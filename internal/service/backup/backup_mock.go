// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/backup/backup.go
//
// Generated by this command:
//
//	mockgen -source=internal/service/backup/backup.go -destination=internal/service/backup/backup_mock.go -package=backup
//

// Package backup is a generated GoMock package.
package backup

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockIBackupService is a mock of IBackupService interface.
type MockIBackupService struct {
	ctrl     *gomock.Controller
	recorder *MockIBackupServiceMockRecorder
	isgomock struct{}
}

// MockIBackupServiceMockRecorder is the mock recorder for MockIBackupService.
type MockIBackupServiceMockRecorder struct {
	mock *MockIBackupService
}

// NewMockIBackupService creates a new mock instance.
func NewMockIBackupService(ctrl *gomock.Controller) *MockIBackupService {
	mock := &MockIBackupService{ctrl: ctrl}
	mock.recorder = &MockIBackupServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIBackupService) EXPECT() *MockIBackupServiceMockRecorder {
	return m.recorder
}

// LoadData mocks base method.
func (m *MockIBackupService) LoadData() (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadData")
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadData indicates an expected call of LoadData.
func (mr *MockIBackupServiceMockRecorder) LoadData() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadData", reflect.TypeOf((*MockIBackupService)(nil).LoadData))
}

// SaveData mocks base method.
func (m *MockIBackupService) SaveData(data map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveData", data)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveData indicates an expected call of SaveData.
func (mr *MockIBackupServiceMockRecorder) SaveData(data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveData", reflect.TypeOf((*MockIBackupService)(nil).SaveData), data)
}