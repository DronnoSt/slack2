// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces.go

// Package dl is a generated GoMock package.
package dl

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	slackdump "github.com/rusq/slackdump/v2"
	slack "github.com/slack-go/slack"
)

// MockExporter is a mock of Exporter interface.
type MockExporter struct {
	ctrl     *gomock.Controller
	recorder *MockExporterMockRecorder
}

// MockExporterMockRecorder is the mock recorder for MockExporter.
type MockExporterMockRecorder struct {
	mock *MockExporter
}

// NewMockExporter creates a new mock instance.
func NewMockExporter(ctrl *gomock.Controller) *MockExporter {
	mock := &MockExporter{ctrl: ctrl}
	mock.recorder = &MockExporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExporter) EXPECT() *MockExporterMockRecorder {
	return m.recorder
}

// ProcessFunc mocks base method.
func (m *MockExporter) ProcessFunc(channelName string) slackdump.ProcessFunc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessFunc", channelName)
	ret0, _ := ret[0].(slackdump.ProcessFunc)
	return ret0
}

// ProcessFunc indicates an expected call of ProcessFunc.
func (mr *MockExporterMockRecorder) ProcessFunc(channelName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessFunc", reflect.TypeOf((*MockExporter)(nil).ProcessFunc), channelName)
}

// Start mocks base method.
func (m *MockExporter) Start(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", ctx)
}

// Start indicates an expected call of Start.
func (mr *MockExporterMockRecorder) Start(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockExporter)(nil).Start), ctx)
}

// Stop mocks base method.
func (m *MockExporter) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockExporterMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockExporter)(nil).Stop))
}

// MockStartStopper is a mock of StartStopper interface.
type MockStartStopper struct {
	ctrl     *gomock.Controller
	recorder *MockStartStopperMockRecorder
}

// MockStartStopperMockRecorder is the mock recorder for MockStartStopper.
type MockStartStopperMockRecorder struct {
	mock *MockStartStopper
}

// NewMockStartStopper creates a new mock instance.
func NewMockStartStopper(ctrl *gomock.Controller) *MockStartStopper {
	mock := &MockStartStopper{ctrl: ctrl}
	mock.recorder = &MockStartStopperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStartStopper) EXPECT() *MockStartStopperMockRecorder {
	return m.recorder
}

// Start mocks base method.
func (m *MockStartStopper) Start(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", ctx)
}

// Start indicates an expected call of Start.
func (mr *MockStartStopperMockRecorder) Start(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockStartStopper)(nil).Start), ctx)
}

// Stop mocks base method.
func (m *MockStartStopper) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockStartStopperMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockStartStopper)(nil).Stop))
}

// MockexportDownloader is a mock of exportDownloader interface.
type MockexportDownloader struct {
	ctrl     *gomock.Controller
	recorder *MockexportDownloaderMockRecorder
}

// MockexportDownloaderMockRecorder is the mock recorder for MockexportDownloader.
type MockexportDownloaderMockRecorder struct {
	mock *MockexportDownloader
}

// NewMockexportDownloader creates a new mock instance.
func NewMockexportDownloader(ctrl *gomock.Controller) *MockexportDownloader {
	mock := &MockexportDownloader{ctrl: ctrl}
	mock.recorder = &MockexportDownloaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockexportDownloader) EXPECT() *MockexportDownloaderMockRecorder {
	return m.recorder
}

// DownloadFile mocks base method.
func (m *MockexportDownloader) DownloadFile(dir string, f slack.File) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadFile", dir, f)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadFile indicates an expected call of DownloadFile.
func (mr *MockexportDownloaderMockRecorder) DownloadFile(dir, f interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadFile", reflect.TypeOf((*MockexportDownloader)(nil).DownloadFile), dir, f)
}

// Start mocks base method.
func (m *MockexportDownloader) Start(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", ctx)
}

// Start indicates an expected call of Start.
func (mr *MockexportDownloaderMockRecorder) Start(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockexportDownloader)(nil).Start), ctx)
}

// Stop mocks base method.
func (m *MockexportDownloader) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockexportDownloaderMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockexportDownloader)(nil).Stop))
}
