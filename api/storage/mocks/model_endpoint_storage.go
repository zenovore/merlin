// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/caraml-dev/merlin/models"
	mock "github.com/stretchr/testify/mock"
)

// ModelEndpointStorage is an autogenerated mock type for the ModelEndpointStorage type
type ModelEndpointStorage struct {
	mock.Mock
}

// FindByID provides a mock function with given fields: ctx, id
func (_m *ModelEndpointStorage) FindByID(ctx context.Context, id models.ID) (*models.ModelEndpoint, error) {
	ret := _m.Called(ctx, id)

	var r0 *models.ModelEndpoint
	if rf, ok := ret.Get(0).(func(context.Context, models.ID) *models.ModelEndpoint); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.ModelEndpoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListModelEndpoints provides a mock function with given fields: ctx, modelID
func (_m *ModelEndpointStorage) ListModelEndpoints(ctx context.Context, modelID models.ID) ([]*models.ModelEndpoint, error) {
	ret := _m.Called(ctx, modelID)

	var r0 []*models.ModelEndpoint
	if rf, ok := ret.Get(0).(func(context.Context, models.ID) []*models.ModelEndpoint); ok {
		r0 = rf(ctx, modelID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.ModelEndpoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ID) error); ok {
		r1 = rf(ctx, modelID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListModelEndpointsInProject provides a mock function with given fields: ctx, projectID, region
func (_m *ModelEndpointStorage) ListModelEndpointsInProject(ctx context.Context, projectID models.ID, region string) ([]*models.ModelEndpoint, error) {
	ret := _m.Called(ctx, projectID, region)

	var r0 []*models.ModelEndpoint
	if rf, ok := ret.Get(0).(func(context.Context, models.ID, string) []*models.ModelEndpoint); ok {
		r0 = rf(ctx, projectID, region)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.ModelEndpoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ID, string) error); ok {
		r1 = rf(ctx, projectID, region)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, newModelEndpoint, prevModelEndpoint
func (_m *ModelEndpointStorage) Save(ctx context.Context, newModelEndpoint *models.ModelEndpoint, prevModelEndpoint *models.ModelEndpoint) error {
	ret := _m.Called(ctx, newModelEndpoint, prevModelEndpoint)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.ModelEndpoint, *models.ModelEndpoint) error); ok {
		r0 = rf(ctx, newModelEndpoint, prevModelEndpoint)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewModelEndpointStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewModelEndpointStorage creates a new instance of ModelEndpointStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewModelEndpointStorage(t mockConstructorTestingTNewModelEndpointStorage) *ModelEndpointStorage {
	mock := &ModelEndpointStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
