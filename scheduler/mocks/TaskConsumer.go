// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	scheduler "github.com/smousa/forgerock-scheduler/scheduler"
	mock "github.com/stretchr/testify/mock"
)

// TaskConsumer is an autogenerated mock type for the TaskConsumer type
type TaskConsumer struct {
	mock.Mock
}

// Consume provides a mock function with given fields: ctx, w
func (_m *TaskConsumer) Consume(ctx context.Context, w scheduler.TaskWorker) error {
	ret := _m.Called(ctx, w)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, scheduler.TaskWorker) error); ok {
		r0 = rf(ctx, w)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
