// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/adigunhammedolalekan/storm (interfaces: DockerService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockDockerService is a mock of DockerService interface
type MockDockerService struct {
	ctrl     *gomock.Controller
	recorder *MockDockerServiceMockRecorder
}

// MockDockerServiceMockRecorder is the mock recorder for MockDockerService
type MockDockerServiceMockRecorder struct {
	mock *MockDockerService
}

// NewMockDockerService creates a new mock instance
func NewMockDockerService(ctrl *gomock.Controller) *MockDockerService {
	mock := &MockDockerService{ctrl: ctrl}
	mock.recorder = &MockDockerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDockerService) EXPECT() *MockDockerServiceMockRecorder {
	return m.recorder
}

// BuildImage mocks base method
func (m *MockDockerService) BuildImage(arg0 context.Context, arg1, arg2 string, arg3 io.Reader) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildImage", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildImage indicates an expected call of BuildImage
func (mr *MockDockerServiceMockRecorder) BuildImage(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildImage", reflect.TypeOf((*MockDockerService)(nil).BuildImage), arg0, arg1, arg2, arg3)
}

// PushImage mocks base method
func (m *MockDockerService) PushImage(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PushImage", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PushImage indicates an expected call of PushImage
func (mr *MockDockerServiceMockRecorder) PushImage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushImage", reflect.TypeOf((*MockDockerService)(nil).PushImage), arg0, arg1)
}
