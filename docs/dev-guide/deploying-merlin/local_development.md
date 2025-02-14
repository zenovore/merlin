# Local Development

In this guide, we will deploy Merlin on a local Minikube cluster.

If you already have existing development cluster, you can run [`quick_install.sh`](../../../scripts/quick_install.sh) to install Merlin and it's components.

## Prerequesites

1. Kubernetes v1.22.7
2. Minikube v1.16.0 with LoadBalancer enabled
3. Istio v1.12.4
4. Knative v1.3.2
5. Cert Manager v1.9.1
6. Kserve v0.8.0
8. Minio v7.0.2

## Provision Minikube cluster

First, you need to have Minikube installed on your machine. To install it, please follow this [documentation](https://minikube.sigs.k8s.io/docs/start/). You also need to have a [driver](https://minikube.sigs.k8s.io/docs/drivers/) to run Minikube cluster. This guide uses VirtualBox driver.

Next, create a new Minikube cluster with Kubernetes v1.22.7:

```bash
export CLUSTER_NAME=dev
minikube start --cpus=4 --memory=8192 --kubernetes-version=v1.22.7 --driver=virtualbox
```

Lastly, we need to enable Minikube's LoadBalancer services by running `minikube tunnel` in another terminal.

## Install Istio

We recommend installing Istio without service mesh (sidecar injection disabled). We also need to enable Istio Kubernetes Ingress enabled so we can access Merlin API and UI.

```bash
export ISTIO_VERSION=1.12.3

curl --location https://git.io/getLatestIstio | sh -

cat << EOF > ./istio-config.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: default
  hub: gcr.io/istio-testing
  tag: latest
  revision: 1-12-4
  meshConfig:
    accessLogFile: /dev/stdout
    enableTracing: true
  components:
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  values:
    global:
      proxy:
        autoInject: disabled
    gateways:
        istio-ingressgateway:
            runAsRoot: true
  components:
    ingressGateways:
      - name: istio-ingressgateway
        enabled: true
        k8s:
          resources:
            requests:
              cpu: 20m
              memory: 64Mi
            limits:
              memory: 128Mi
      - name: cluster-local-gateway
        enabled: true
        label:
          istio: cluster-local-gateway
          app: cluster-local-gateway
        k8s:
          resources:
            requests:
              cpu: 20m
              memory: 64Mi
            limits:
              memory: 128Mi
          service:
            type: ClusterIP
            ports:
              - port: 15020
                name: status-port
              - port: 80
                name: http2
              - port: 443
                name: https
EOF
istio-${ISTIO_VERSION}/bin/istioctl manifest apply -f istio-config.yaml
```

## Install Knative

In this step, we install Knative Serving and configure it to use Istio as ingress controller.

```bash
export KNATIVE_VERSION=v1.3.2
export KNATIVE_NET_ISTIO_VERSION=v1.3.0

kubectl apply --filename=https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename=https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml

export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cat <<EOF > ./patch-config-domain.json
{
  "data": {
    "${INGRESS_HOST}.nip.io": ""
  }
}
EOF
kubectl patch configmap/config-domain --namespace=knative-serving --type=merge --patch="$(cat patch-config-domain.json)"

# Install Knative Net Istio
kubectl apply --filename=https://github.com/knative-sandbox/net-istio/releases/download/knative-${KNATIVE_NET_ISTIO_VERSION}/release.yaml
```

## Install Cert Manager

```bash
export CERT_MANAGER_VERSION=v1.9.1

kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
```

## Install Kserve

Kserve manages the deployment of Merlin models.

```bash
export KSERVE_VERSION=v0.8.0

kubectl apply --filename=https://github.com/kserve/kserve/releases/download/${KSERVE_VERSION}/kserve.yaml
kubectl apply --filename=https://github.com/kserve/kserve/releases/download/${KSERVE_VERSION}/kserve-runtimes.yaml

cat <<EOF > ./patch-config-inferenceservice.json
{
  "data": {
    "storageInitializer": "{\n    \"image\" : \"ghcr.io/ariefrahmansyah/kfserving-storage-init:latest\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"100m\",\n    \"cpuLimit\": \"1\"\n}",
  }
}
EOF
kubectl patch configmap/inferenceservice-config --namespace=kfserving-system --type=merge --patch="$(cat patch-config-inferenceservice.json)"
```

> Notes that we change Kserve's Storage Initializer image here so it can download the model artifacts from Minio.

## Setup cluster credentials

```bash
cat <<EOF | yq e -P - > k8s_config.yaml
{
  "k8s_config": {
    "name": "dev",
    "cluster": {
      "server": "https://kubernetes.default.svc.cluster.local:443",
      "certificate-authority-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/ca.crt | base64)"
    },
    "user": {
      "client-certificate-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/profiles/minikube/client.crt | base64)",
      "client-key-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/profiles/minikube/client.key | base64)"
    }
  }
}
EOF
```

## Install Minio

Minio is used by MLflow to store model artifacts.

```bash
export MINIO_VERSION=7.0.2

cat <<EOF > minio-values.yaml
replicas: 1
persistence:
  enabled: false
resources:
  requests:
    cpu: 25m
    memory: 64Mi
livenessProbe:
  initialDelaySeconds: 30
defaultBucket:
  enabled: true
  name: mlflow
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: istio
  path: /*
  hosts:
    - 'minio.minio.${INGRESS_HOST}.nip.io'
EOF

kubectl create namespace minio
helm repo add minio https://helm.min.io/
helm install minio minio/minio --version=${MINIO_VERSION} --namespace=minio --values=minio-values.yaml --wait --timeout=600s
```

## Install MLP

MLP and Merlin use Google Sign-in to authenticate the user to access the API and UI. Please follow [this documentation](https://developers.google.com/identity/protocols/oauth2/javascript-implicit-flow) to create Google Authorization credential. You must specify Javascript origins and redirect URIs with both `http://mlp.mlp.${INGRESS_HOST}.nip.io` and `http://merlin.mlp.${INGRESS_HOST}.nip.io`. After you get the client ID, specify it into `OAUTH_CLIENT_ID`.

```bash
export OAUTH_CLIENT_ID="<put your oauth client id here>"

kubectl create namespace mlp

git clone git@github.com:caraml-dev/mlp.git

helm install mlp ./mlp/chart --namespace=mlp --values=./mlp/chart/values-e2e.yaml \
  --set mlp.image.tag=main \
  --set mlp.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set mlp.oauthClientID=${OAUTH_CLIENT_ID} \
  --set mlp.mlflowTrackingUrl=http://mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlp.ingress.enabled=true \
  --set mlp.ingress.class=istio \
  --set mlp.ingress.host=mlp.mlp.${INGRESS_HOST}.nip.io \
  --set mlp.ingress.path="/*" \
  --wait --timeout=5m

cat <<EOF > mlp-ingress.yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: mlp
  namespace: mlp
  annotations:
    kubernetes.io/ingress.class: istio
  labels:
    app: mlp
spec:
  rules:
    - host: 'mlp.mlp.${INGRESS_HOST}.nip.io'
      http:
        paths:
          - path: /*
            backend:
              serviceName: mlp
              servicePort: 8080
EOF
```

## Install Merlin

```bash
export MERLIN_VERSION=82ca798 # TODO: update to use new merlin version once vault dependency removed

output=$(yq e -o json '.k8s_config' k8s_config.yaml | jq -r -M -c .)
yq '.merlin.environmentConfigs[0] *= load("k8s_config.yaml")' ../charts/merlin/values-e2e.yaml > ../charts/merlin/values-e2e-with-k8s_config.yaml
output="$output" yq '.merlin.imageBuilder.k8sConfig |= strenv(output)' -i ../charts/merlin/values-e2e-with-k8s_config.yaml

helm upgrade --install merlin ../charts/merlin --namespace=mlp --values=../charts/merlin/values-e2e-with-k8s_config.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.oauthClientID=${OAUTH_CLIENT_ID} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --timeout=5m \
  --wait
```

### Check Merlin installation

```bash
kubectl get po -n mlp
NAME                             READY   STATUS    RESTARTS   AGE
merlin-64c9c75dfc-djs4t          1/1     Running   0          12m
merlin-mlflow-5c7dd6d9df-g2s6v   1/1     Running   0          12m
merlin-postgresql-0              1/1     Running   0          12m
mlp-6877d8567-msqg9              1/1     Running   0          15m
mlp-postgresql-0                 1/1     Running   0          15m
```

Once everything is Running, you can open Merlin in <http://merlin.mlp.${INGRESS_HOST}.nip.io/merlin>. From here, you can run Jupyter notebook examples by setting `merlin.set_url("merlin.mlp.${INGRESS_HOST}.nip.io")`.
