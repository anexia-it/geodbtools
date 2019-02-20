// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anexia-it/geodbtools (interfaces: Format)

// Package geodbtools is a generated GoMock package.
package geodbtools

import (
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockFormat is a mock of Format interface
type MockFormat struct {
	ctrl     *gomock.Controller
	recorder *MockFormatMockRecorder
}

// MockFormatMockRecorder is the mock recorder for MockFormat
type MockFormatMockRecorder struct {
	mock *MockFormat
}

// NewMockFormat creates a new mock instance
func NewMockFormat(ctrl *gomock.Controller) *MockFormat {
	mock := &MockFormat{ctrl: ctrl}
	mock.recorder = &MockFormatMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFormat) EXPECT() *MockFormatMockRecorder {
	return m.recorder
}

// DetectFormat mocks base method
func (m *MockFormat) DetectFormat(arg0 ReaderSource) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetectFormat", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// DetectFormat indicates an expected call of DetectFormat
func (mr *MockFormatMockRecorder) DetectFormat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetectFormat", reflect.TypeOf((*MockFormat)(nil).DetectFormat), arg0)
}

// FormatName mocks base method
func (m *MockFormat) FormatName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatName")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatName indicates an expected call of FormatName
func (mr *MockFormatMockRecorder) FormatName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatName", reflect.TypeOf((*MockFormat)(nil).FormatName))
}

// NewReaderAt mocks base method
func (m *MockFormat) NewReaderAt(arg0 ReaderSource) (Reader, Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewReaderAt", arg0)
	ret0, _ := ret[0].(Reader)
	ret1, _ := ret[1].(Metadata)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// NewReaderAt indicates an expected call of NewReaderAt
func (mr *MockFormatMockRecorder) NewReaderAt(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewReaderAt", reflect.TypeOf((*MockFormat)(nil).NewReaderAt), arg0)
}

// NewWriter mocks base method
func (m *MockFormat) NewWriter(arg0 io.Writer, arg1 DatabaseType, arg2 IPVersion) (Writer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewWriter", arg0, arg1, arg2)
	ret0, _ := ret[0].(Writer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewWriter indicates an expected call of NewWriter
func (mr *MockFormatMockRecorder) NewWriter(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewWriter", reflect.TypeOf((*MockFormat)(nil).NewWriter), arg0, arg1, arg2)
}