// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anexia-it/geodbtools (interfaces: ReaderSource)

// Package mmdatformat is a generated GoMock package.
package mmdatformat

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockReaderSource is a mock of ReaderSource interface
type MockReaderSource struct {
	ctrl     *gomock.Controller
	recorder *MockReaderSourceMockRecorder
}

// MockReaderSourceMockRecorder is the mock recorder for MockReaderSource
type MockReaderSourceMockRecorder struct {
	mock *MockReaderSource
}

// NewMockReaderSource creates a new mock instance
func NewMockReaderSource(ctrl *gomock.Controller) *MockReaderSource {
	mock := &MockReaderSource{ctrl: ctrl}
	mock.recorder = &MockReaderSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReaderSource) EXPECT() *MockReaderSourceMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockReaderSource) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockReaderSourceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockReaderSource)(nil).Close))
}

// ReadAt mocks base method
func (m *MockReaderSource) ReadAt(arg0 []byte, arg1 int64) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadAt", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadAt indicates an expected call of ReadAt
func (mr *MockReaderSourceMockRecorder) ReadAt(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadAt", reflect.TypeOf((*MockReaderSource)(nil).ReadAt), arg0, arg1)
}

// Size mocks base method
func (m *MockReaderSource) Size() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Size indicates an expected call of Size
func (mr *MockReaderSourceMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockReaderSource)(nil).Size))
}
