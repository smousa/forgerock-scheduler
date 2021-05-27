// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	scheduler "github.com/smousa/forgerock-scheduler/scheduler"
	mock "github.com/stretchr/testify/mock"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// Add provides a mock function with given fields: _a0, _a1
func (_m *Store) Add(_a0 context.Context, _a1 *scheduler.Job) (string, error) {
	ret := _m.Called(_a0, _a1)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *scheduler.Job) string); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *scheduler.Job) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: ctx, jobID, status
func (_m *Store) UpdateStatus(ctx context.Context, jobID string, status *scheduler.Status) {
	_m.Called(ctx, jobID, status)
}

// WaitForTaskToComplete provides a mock function with given fields: ctx, jobID, taskName
func (_m *Store) WaitForTaskToComplete(ctx context.Context, jobID string, taskName string) (*scheduler.Status, error) {
	ret := _m.Called(ctx, jobID, taskName)

	var r0 *scheduler.Status
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *scheduler.Status); ok {
		r0 = rf(ctx, jobID, taskName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*scheduler.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, jobID, taskName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
