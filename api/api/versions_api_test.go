// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/mlflow"
	mlfmocks "github.com/caraml-dev/merlin/mlflow/mocks"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *Response
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model version 1 for version 1"},
			},
		},
		{
			desc: "Should return 500 if error when fetching version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version for given model 1 version 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.GetVersion(&http.Request{}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestListVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		versionService func() *mocks.VersionsService
		queryParameter string
		expected       *Response
	}{
		{
			desc: "Should success get version",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				}, "", nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				},
				headers: map[string]string{},
			},
		},
		{
			desc: "Should success get version with pagination",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return([]*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				}, "NDdfMzQ=", nil)
				return svc
			},
			queryParameter: "limit=30",
			expected: &Response{
				code: http.StatusOK,
				data: []*models.Version{
					{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					},
				},
				headers: map[string]string{
					"Next-Cursor": "NDdfMzQ=",
				},
			},
		},
		{
			desc: "Should return 500 if get version returning error",
			vars: map[string]string{
				"model_id": "1",
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("ListVersions", mock.Anything, models.ID(1), mock.Anything, mock.Anything).Return(nil, "", fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "DB is down"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.ListVersions(&http.Request{URL: &url.URL{RawQuery: tC.queryParameter}}, tC.vars, nil)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestPatchVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		requestBody    interface{}
		vars           map[string]string
		versionService func() *mocks.VersionsService
		expected       *Response
	}{
		{
			desc: "Should success patch version",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				},
			},
		},
		{
			desc: "Should success patch version - patch customer container if model type is CUSTOM",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Image:   "gcr.io/custom-predictor:v0.1",
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-predictor:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					CustomPredictor: &models.CustomPredictor{
						Image:   "gcr.io/custom-predictor:v0.1",
						Command: "./run.sh",
						Args:    "firstArg secondArg",
					},
				},
			},
		},
		{
			desc: "Should success patch version - do nothing when trying patch custom_container where its type is not CUSTOM",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Image:   "gcr.io/custom-predictor:v0.1",
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
				},
			},
		},
		{
			desc: "Return 400 patch version - model type custom but custom predictor object doesn't have image",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{
				CustomPredictor: &models.CustomPredictor{
					Command: "./run.sh",
					Args:    "firstArg secondArg",
				},
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "custom",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL:       "http://mlflow.com",
						CustomPredictor: nil,
					}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "custom predictor image must be set"},
			},
		},
		{
			desc: "Should return 404 if version is not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					nil, gorm.ErrRecordNotFound)
				return svc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model version 1 for version 1"},
			},
		},
		{
			desc: "Should return 500 if version fetching returning error",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error getting model version for given model 1 version 1"},
			},
		},
		{
			desc: "Should return 500 if request body is not valud",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.Model{
				ID: models.ID(1),
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Unable to parse request body"},
			},
		},
		{
			desc: "Should return 500 if save is failing",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			requestBody: &models.VersionPatch{Properties: &models.KV{
				"name":       "model-1",
				"created_by": "anonymous",
			}},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(
					&models.Version{
						ID:      models.ID(1),
						ModelID: models.ID(1),
						Model: &models.Model{
							ID:           models.ID(1),
							Name:         "model-1",
							ProjectID:    models.ID(1),
							Project:      mlp.Project{},
							ExperimentID: 1,
							Type:         "pyfunc",
							MlflowURL:    "http://mlflow.com",
						},
						MlflowURL: "http://mlflow.com",
					}, nil)
				svc.On("Save", mock.Anything, &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "pyfunc",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Properties: models.KV{
						"name":       "model-1",
						"created_by": "anonymous",
					},
				}, mock.Anything).Return(nil, fmt.Errorf("DB is down"))
				return svc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error patching model version for given model 1 version 1"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled: true,
				},
			}
			resp := ctl.PatchVersion(&http.Request{}, tC.vars, tC.requestBody)
			assert.Equal(t, tC.expected, resp)
		})
	}
}

func TestCreateVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		vars           map[string]string
		body           models.VersionPost
		versionService func() *mocks.VersionsService
		mlflowClient   func() *mlfmocks.Client
		modelsService  func() *mocks.ModelsService
		expected       *Response
	}{
		{
			desc: "Should successfully create version",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service.type":     "GO-FOOD",
					"1-targeting_date": "2021-02-01",
					"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
				},
				PythonVersion: "3.10.*",
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:     models.ID(1),
					RunID:       "1",
					ArtifactURI: "artifact/url/run",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL: "http://mlflow.com",
					Labels: models.KV{
						"service.type":     "GO-FOOD",
						"1-targeting_date": "2021-02-01",
						"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverThe",
					},
					PythonVersion: "3.10.*",
				},
			},
		},
		{
			desc: "Should fail label key validation: has emoji inside",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"😊😊😊😊😊": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: start with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"-service_type": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: end with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type-": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label key validation: 64 characters",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverTheL": "GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: has emoji inside",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"emoji-label": "😊😊😊😊😊",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: start with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type": "-GO-FOOD",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: end with non alphanumeric",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"service_type": "GO-FOOD-",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should fail label value validation: 64 characters",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{
				Labels: models.KV{
					"some_valid_key": "TheQuickBrownFoxJumpsOverTheLazyDogTheQuickBrownFoxJumpsOverTheL",
				},
			},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{}, mock.Anything).Return(&models.Version{}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusBadRequest,
				data: Error{Message: "Valid label key/values must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (_), dots (.), and alphanumerics between."},
			},
		},
		{
			desc: "Should successfully create version without labels, default python version",
			vars: map[string]string{
				"model_id": "1",
			},
			body: models.VersionPost{},
			modelsService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					Name:      "model-1",
					ProjectID: models.ID(1),
					Project: mlp.Project{
						MLFlowTrackingURL: "http://www.notinuse.com",
					},
					ExperimentID: 1,
					Type:         "pyfunc",
					MlflowURL:    "http://mlflow.com",
					Endpoints:    nil,
				}, nil)
				return svc
			},
			mlflowClient: func() *mlfmocks.Client {
				svc := &mlfmocks.Client{}
				svc.On("CreateRun", "1").Return(&mlflow.Run{
					Info: mlflow.Info{
						RunID:       "1",
						ArtifactURI: "artifact/url/run",
					},
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("Save", mock.Anything, &models.Version{
					ModelID:       models.ID(1),
					RunID:         "1",
					ArtifactURI:   "artifact/url/run",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				}, mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				}, nil)
				return svc
			},
			expected: &Response{
				code: http.StatusCreated,
				data: &models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Model: &models.Model{
						ID:           models.ID(1),
						Name:         "model-1",
						ProjectID:    models.ID(1),
						Project:      mlp.Project{},
						ExperimentID: 1,
						Type:         "sklearn",
						MlflowURL:    "http://mlflow.com",
					},
					MlflowURL:     "http://mlflow.com",
					PythonVersion: DEFAULT_PYTHON_VERSION,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			versionSvc := tC.versionService()
			modelsSvc := tC.modelsService()
			mlflowClient := tC.mlflowClient()

			ctl := &VersionsController{
				AppContext: &AppContext{
					VersionsService: versionSvc,
					MonitoringConfig: config.MonitoringConfig{
						MonitoringEnabled: true,
						MonitoringBaseURL: "http://grafana",
					},
					AlertEnabled:  true,
					MlflowClient:  mlflowClient,
					ModelsService: modelsSvc,
				},
			}
			resp := ctl.CreateVersion(&http.Request{}, tC.vars, &tC.body)
			assert.Equal(t, tC.expected, resp)
		})
	}
}
