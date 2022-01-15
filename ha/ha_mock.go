// Code generated by MockGen. DO NOT EDIT.
// Source: ha.go

// Package ha is a generated GoMock package.
package ha

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockHA is a mock of HA interface.
type MockHA struct {
	ctrl     *gomock.Controller
	recorder *MockHAMockRecorder
}

// MockHAMockRecorder is the mock recorder for MockHA.
type MockHAMockRecorder struct {
	mock *MockHA
}

// NewMockHA creates a new mock instance.
func NewMockHA(ctrl *gomock.Controller) *MockHA {
	mock := &MockHA{ctrl: ctrl}
	mock.recorder = &MockHAMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHA) EXPECT() *MockHAMockRecorder {
	return m.recorder
}

// MushMaster mocks base method.
func (m *MockHA) MushMaster(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MushMaster", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// MushMaster indicates an expected call of MushMaster.
func (mr *MockHAMockRecorder) MushMaster(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MushMaster", reflect.TypeOf((*MockHA)(nil).MushMaster), arg0)
}
