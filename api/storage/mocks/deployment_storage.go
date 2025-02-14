// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	models "github.com/caraml-dev/merlin/models"
	mock "github.com/stretchr/testify/mock"
)

// DeploymentStorage is an autogenerated mock type for the DeploymentStorage type
type DeploymentStorage struct {
	mock.Mock
}

// GetFirstSuccessModelVersionPerModel provides a mock function with given fields:
func (_m *DeploymentStorage) GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error) {
	ret := _m.Called()

	var r0 map[models.ID]models.ID
	if rf, ok := ret.Get(0).(func() map[models.ID]models.ID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[models.ID]models.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListInModel provides a mock function with given fields: model
func (_m *DeploymentStorage) ListInModel(model *models.Model) ([]*models.Deployment, error) {
	ret := _m.Called(model)

	var r0 []*models.Deployment
	if rf, ok := ret.Get(0).(func(*models.Model) []*models.Deployment); ok {
		r0 = rf(model)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Deployment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.Model) error); ok {
		r1 = rf(model)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: deployment
func (_m *DeploymentStorage) Save(deployment *models.Deployment) (*models.Deployment, error) {
	ret := _m.Called(deployment)

	var r0 *models.Deployment
	if rf, ok := ret.Get(0).(func(*models.Deployment) *models.Deployment); ok {
		r0 = rf(deployment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Deployment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.Deployment) error); ok {
		r1 = rf(deployment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDeploymentStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewDeploymentStorage creates a new instance of DeploymentStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDeploymentStorage(t mockConstructorTestingTNewDeploymentStorage) *DeploymentStorage {
	mock := &DeploymentStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
