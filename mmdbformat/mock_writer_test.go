// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anexia-it/geodbtools (interfaces: Writer)

// Package mmdbformat is a generated GoMock package.
package mmdbformat

import (
	geodbtools "github.com/anexia-it/geodbtools"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockWriter is a mock of Writer interface
type MockWriter struct {
	ctrl     *gomock.Controller
	recorder *MockWriterMockRecorder
}

// MockWriterMockRecorder is the mock recorder for MockWriter
type MockWriterMockRecorder struct {
	mock *MockWriter
}

// NewMockWriter creates a new mock instance
func NewMockWriter(ctrl *gomock.Controller) *MockWriter {
	mock := &MockWriter{ctrl: ctrl}
	mock.recorder = &MockWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWriter) EXPECT() *MockWriterMockRecorder {
	return m.recorder
}

// WriteDatabase mocks base method
func (m *MockWriter) WriteDatabase(arg0 geodbtools.Metadata, arg1 *geodbtools.RecordTree) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteDatabase", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteDatabase indicates an expected call of WriteDatabase
func (mr *MockWriterMockRecorder) WriteDatabase(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteDatabase", reflect.TypeOf((*MockWriter)(nil).WriteDatabase), arg0, arg1)
}
