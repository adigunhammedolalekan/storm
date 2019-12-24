// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/adigunhammedolalekan/storm (interfaces: K8sService)

// Package mocks is a generated GoMock package.
package mocks

import (
	storm "github.com/adigunhammedolalekan/storm"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockK8sService is a mock of K8sService interface
type MockK8sService struct {
	ctrl     *gomock.Controller
	recorder *MockK8sServiceMockRecorder
}

// MockK8sServiceMockRecorder is the mock recorder for MockK8sService
type MockK8sServiceMockRecorder struct {
	mock *MockK8sService
}

// NewMockK8sService creates a new mock instance
func NewMockK8sService(ctrl *gomock.Controller) *MockK8sService {
	mock := &MockK8sService{ctrl: ctrl}
	mock.recorder = &MockK8sServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockK8sService) EXPECT() *MockK8sServiceMockRecorder {
	return m.recorder
}

// DeployService mocks base method
func (m *MockK8sService) DeployService(arg0, arg1 string, arg2 map[string]string, arg3 bool) (*storm.DeploymentResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeployService", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*storm.DeploymentResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeployService indicates an expected call of DeployService
func (mr *MockK8sServiceMockRecorder) DeployService(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeployService", reflect.TypeOf((*MockK8sService)(nil).DeployService), arg0, arg1, arg2, arg3)
}
