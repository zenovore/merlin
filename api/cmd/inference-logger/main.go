package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/kelseyhightower/envconfig"
	nrconfig "github.com/newrelic/newrelic-client-go/v2/pkg/config"
	nrlog "github.com/newrelic/newrelic-client-go/v2/pkg/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	network "knative.dev/networking/pkg"
	pkgnet "knative.dev/pkg/network"
	pkghandler "knative.dev/pkg/network/handlers"
	"knative.dev/pkg/signals"
	"knative.dev/serving/pkg/queue"
	"knative.dev/serving/pkg/queue/health"
	"knative.dev/serving/pkg/queue/readiness"

	"github.com/caraml-dev/merlin/pkg/inference-logger/liveness"
	merlinlogger "github.com/caraml-dev/merlin/pkg/inference-logger/logger"
)

var (
	logUrl           = flag.String("log-url", "localhost:8002", "The URL to send request/response logs to")
	port             = flag.String("port", "9081", "Logger port")
	componentPort    = flag.String("component-port", "8080", "Component port")
	workers          = flag.Int("workers", 5, "Number of workers")
	logMode          = flag.String("log-mode", string(merlinlogger.LogModeAll), "Whether to log 'request', 'response' or 'all'")
	inferenceService = flag.String("inference-service", "my-model-1", "The InferenceService name to add as header to log events")
	namespace        = flag.String("namespace", "my-project", "The namespace to add as header to log events")
)

const (
	HystrixTimeout                = 5000
	HystrixMaxConcurrentRequests  = 2000
	HystrixRequestVolumeThreshold = 2000
	HystrixSleepWindow            = 2000
	HystrixErrorPercentThreshold  = 10

	// Internal queue config
	QueueMinBatchSize = 1
	QueueMaxBatchSize = 32
	QueueCapacity     = 1000

	// Duration the /wait-for-drain handler should wait before returning.
	// This is to give networking a little bit more time to remove the pod
	// from its configuration and propagate that to all loadbalancers and nodes.
	drainSleepDuration = 30 * time.Second

	// NewRelic client configuration
	NewRelicLogLevel = "info"
)

type config struct {
	// Making the below fields optional since raw deployment wont have them
	ServingReadinessProbe string `split_words:"true" required:"false"`

	UnixSocketPath string `split_words:"true" required:"false" default:"@/kserve/agent.sock"`
}

func main() {
	flag.Parse()

	l, _ := zap.NewProduction()
	log := l.Sugar()

	var env config
	if err := envconfig.Process("", &env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Setup probe to run for checking user container healthiness.
	probe := func() bool { return true }
	if env.ServingReadinessProbe != "" {
		probe = buildProbe(log, env.ServingReadinessProbe).ProbeContainer
	}

	if *logUrl == "" {
		log.Info("log-url argument must not be empty.")
		os.Exit(-1)
	}

	_, err := url.Parse(*logUrl)
	if err != nil {
		log.Info("Malformed log-url", "URL", *logUrl)
		os.Exit(-1)
	}
	loggingMode := merlinlogger.LogMode(*logMode)
	switch loggingMode {
	case merlinlogger.LogModeAll, merlinlogger.LogModeRequestOnly, merlinlogger.LogModeResponseOnly:
	default:
		log.Info("Malformed log-mode", "mode", *logMode)
		os.Exit(-1)
	}

	target := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", *componentPort),
	}

	workerConfig := &merlinlogger.WorkerConfig{
		MinBatchSize: QueueMinBatchSize,
		MaxBatchSize: QueueMaxBatchSize,
	}

	logSink := getLogSink(*logUrl, log)
	dispatcher := createLoggerDispatcher(workerConfig, logSink, log)
	dispatcher.Start()

	// Create handler chain.
	// Note: innermost handlers are specified first, ie. the last handler in the chain will be executed first.
	drainer, mainServer := buildServer(target, dispatcher, loggingMode, probe, log)

	ctx := signals.NewContext()
	servers := map[string]*http.Server{
		"main": mainServer,
	}
	errCh := make(chan error)
	listenCh := make(chan struct{})
	log.Infof("Listening at :%s", *port)
	for name, server := range servers {
		go func(name string, s *http.Server) {
			l, err := net.Listen("tcp", s.Addr)
			if err != nil {
				errCh <- fmt.Errorf("%s server failed to listen: %w", name, err)
				return
			}

			// Notify the unix socket setup that the tcp socket for the main server is ready.
			if s == mainServer {
				close(listenCh)
			}

			// Don't forward ErrServerClosed as that indicates we're already shutting down.
			if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("%s server failed to serve: %w", name, err)
			}
		}(name, server)
	}

	// Listen on a unix socket so that the exec probe can avoid having to go
	// through the full tcp network stack.
	go func() {
		// Only start listening on the unix socket once the tcp socket for the
		// main server is setup.
		// This avoids the unix socket path succeeding before the tcp socket path
		// is actually working and thus it avoids a race.
		<-listenCh

		l, err := net.Listen("unix", env.UnixSocketPath)
		if err != nil {
			errCh <- fmt.Errorf("failed to listen to unix socket: %w", err)
			return
		}
		if err := http.Serve(l, mainServer.Handler); err != nil {
			errCh <- fmt.Errorf("serving failed on unix socket: %w", err)
		}
	}()

	// Blocks until we actually receive a TERM signal or one of the servers
	// exit unexpectedly. We fold both signals together because we only want
	// to act on the first of those to reach here.
	select {
	case err := <-errCh:
		log.Errorw("Failed to bring up agent, shutting down.", zap.Error(err))
		// This extra flush is needed because defers are not handled via os.Exit calls.
		_ = log.Sync()
		_ = os.Stdout.Sync()
		_ = os.Stderr.Sync()
		os.Exit(1)
	case <-ctx.Done():
		log.Info("Received TERM signal, attempting to gracefully shutdown servers.")
		log.Infof("Sleeping %v to allow K8s propagation of non-ready state", drainSleepDuration)
		drainer.Drain()
		dispatcher.Stop()

		for serverName, srv := range servers {
			log.Info("Shutting down server: ", serverName)
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Errorw("Failed to shutdown server", zap.String("server", serverName), zap.Error(err))
			}
		}
		log.Info("Shutdown complete, exiting...")
	}
}

func buildServer(target *url.URL, dispatcher *merlinlogger.Dispatcher, loggingMode merlinlogger.LogMode, probe func() bool, log *zap.SugaredLogger) (*pkghandler.Drainer, *http.Server) {
	maxIdleConns := 1000 // TODO: somewhat arbitrary value for CC=0, needs experimental validation.

	httpProxy := httputil.NewSingleHostReverseProxy(target)
	httpProxy.Transport = pkgnet.NewAutoTransport(maxIdleConns /* max-idle */, maxIdleConns /* max-idle-per-host */)
	// nolint:staticcheck
	httpProxy.ErrorHandler = pkgnet.ErrorHandler(log)
	httpProxy.BufferPool = network.NewBufferPool()
	httpProxy.FlushInterval = network.FlushInterval

	var composedHandler http.Handler = httpProxy
	composedHandler = merlinlogger.NewLoggerHandler(dispatcher, loggingMode, composedHandler, log)

	inner := queue.ForwardedShimHandler(composedHandler)
	composedHandler = inner

	drainer := &pkghandler.Drainer{
		QuietPeriod: drainSleepDuration,
		// Add Activator probe header to the drainer so it can handle probes directly from activator
		HealthCheckUAPrefixes: []string{network.ActivatorUserAgent},
		Inner:                 composedHandler,
		HealthCheck:           health.ProbeHandler(probe, false),
	}
	composedHandler = liveness.NewProbe(inner, drainer)
	mainServer := pkgnet.NewServer(":"+*port, composedHandler)
	return drainer, mainServer
}

func createLoggerDispatcher(
	workerConfig *merlinlogger.WorkerConfig,
	logSink merlinlogger.LogSink,
	log *zap.SugaredLogger,
) *merlinlogger.Dispatcher {
	// TODO: Remove default console logging
	dispatcher := merlinlogger.NewDispatcher(*workers, QueueCapacity, workerConfig, log, logSink, merlinlogger.NewConsoleSink(log))
	return dispatcher
}

func buildProbe(logger *zap.SugaredLogger, probeJSON string) *readiness.Probe {
	coreProbe, err := readiness.DecodeProbe(probeJSON)
	if err != nil {
		logger.Fatalw("Agent failed to parse readiness probe", zap.Error(err))
	}
	newProbe := readiness.NewProbe(coreProbe)
	if newProbe.InitialDelaySeconds == 0 {
		newProbe.InitialDelaySeconds = 10
	}
	return newProbe
}

func getServiceName(projectName string, modelName string) string {
	return fmt.Sprintf("%s-%s", projectName, modelName)
}

func getModelNameAndVersion(inferenceServiceName string) (modelName string, modelVersion string) {
	if !strings.Contains(inferenceServiceName, "-") {
		return inferenceServiceName, "1"
	}

	idx := strings.LastIndex(inferenceServiceName, "-")
	modelName = inferenceServiceName[:idx]
	modelVersion = inferenceServiceName[idx+1:]
	return
}

func getTopicName(serviceName string) string {
	return fmt.Sprintf("merlin-%s-inference-log", serviceName)
}

func getLogSink(
	logUrl string,
	log *zap.SugaredLogger,
) merlinlogger.LogSink {
	var sinkKind merlinlogger.LoggerSinkKind
	host := logUrl
	for _, loggerSinkKind := range merlinlogger.LoggerSinkKinds {
		if strings.HasPrefix(logUrl, loggerSinkKind) {
			sinkKind = loggerSinkKind
			host = strings.TrimPrefix(logUrl, loggerSinkKind+":")
		}
	}

	// Map inputs to respective variables for logging
	projectName := *namespace
	modelName, modelVersion := getModelNameAndVersion(*inferenceService)
	serviceName := getServiceName(projectName, modelName)
	topicName := getTopicName(serviceName)

	switch sinkKind {
	case merlinlogger.NewRelic:
		apiKeySplits := strings.Split(logUrl, "?")
		apiKey := apiKeySplits[len(apiKeySplits)-1]

		// https://github.com/newrelic/newrelic-client-go/blob/main/pkg/logs/logs.go
		// Initialize the client configuration
		cfg := nrconfig.New()
		cfg.LicenseKey = apiKey
		cfg.LogLevel = NewRelicLogLevel
		cfg.Compression = nrconfig.Compression.Gzip
		// Initialize the client
		var logClient merlinlogger.NewRelicLogsClient
		nrLogClient := nrlog.New(cfg)
		logClient = &nrLogClient

		return merlinlogger.NewNewRelicSink(log, logClient, serviceName, projectName, modelName, modelVersion)
	case merlinlogger.Kafka:
		// Initialize the producer
		var kafkaProducer merlinlogger.KafkaProducer
		// Create Kafka Producer
		kafkaProducer, err := kafka.NewProducer(
			&kafka.ConfigMap{
				"bootstrap.servers": host,
				"message.max.bytes": merlinlogger.MaxMessageBytes,
				"compression.type":  merlinlogger.CompressionType,
			},
		)
		if err != nil {
			log.Info(err)
			return nil
		}
		// Test that we are able to query the broker on the topic. If the topic
		// does not already exist on the broker, this should create it.
		_, err = kafkaProducer.GetMetadata(&topicName, false, merlinlogger.ConnectTimeoutMS)
		if err != nil {
			log.Info(err)
			return nil
		}

		return merlinlogger.NewKafkaSink(log, kafkaProducer, serviceName, projectName, modelName, modelVersion, topicName)
	default:
		return merlinlogger.NewConsoleSink(log)
	}
}
