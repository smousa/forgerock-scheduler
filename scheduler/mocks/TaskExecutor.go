// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	api "github.com/smousa/forgerock-scheduler/api"

	mock "github.com/stretchr/testify/mock"
)

// TaskExecutor is an autogenerated mock type for the TaskExecutor type
type TaskExecutor struct {
	mock.Mock
}

// Sleep provides a mock function with given fields: ctx, sleep
func (_m *TaskExecutor) Sleep(ctx context.Context, sleep *api.SleepAction) error {
	ret := _m.Called(ctx, sleep)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *api.SleepAction) error); ok {
		r0 = rf(ctx, sleep)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
