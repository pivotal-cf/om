// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/pivotal-cf/om/api"
)

type ApplyChangesService struct {
	CreateInstallationStub        func(bool, bool, bool, []string, api.ApplyErrandChanges) (api.InstallationsServiceOutput, error)
	createInstallationMutex       sync.RWMutex
	createInstallationArgsForCall []struct {
		arg1 bool
		arg2 bool
		arg3 bool
		arg4 []string
		arg5 api.ApplyErrandChanges
	}
	createInstallationReturns struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	createInstallationReturnsOnCall map[int]struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	GetInstallationStub        func(int) (api.InstallationsServiceOutput, error)
	getInstallationMutex       sync.RWMutex
	getInstallationArgsForCall []struct {
		arg1 int
	}
	getInstallationReturns struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	getInstallationReturnsOnCall map[int]struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	GetInstallationLogsStub        func(int) (api.InstallationsServiceOutput, error)
	getInstallationLogsMutex       sync.RWMutex
	getInstallationLogsArgsForCall []struct {
		arg1 int
	}
	getInstallationLogsReturns struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	getInstallationLogsReturnsOnCall map[int]struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	InfoStub        func() (api.Info, error)
	infoMutex       sync.RWMutex
	infoArgsForCall []struct {
	}
	infoReturns struct {
		result1 api.Info
		result2 error
	}
	infoReturnsOnCall map[int]struct {
		result1 api.Info
		result2 error
	}
	ListInstallationsStub        func() ([]api.InstallationsServiceOutput, error)
	listInstallationsMutex       sync.RWMutex
	listInstallationsArgsForCall []struct {
	}
	listInstallationsReturns struct {
		result1 []api.InstallationsServiceOutput
		result2 error
	}
	listInstallationsReturnsOnCall map[int]struct {
		result1 []api.InstallationsServiceOutput
		result2 error
	}
	RunningInstallationStub        func() (api.InstallationsServiceOutput, error)
	runningInstallationMutex       sync.RWMutex
	runningInstallationArgsForCall []struct {
	}
	runningInstallationReturns struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	runningInstallationReturnsOnCall map[int]struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}
	UpdateStagedDirectorPropertiesStub        func(api.DirectorProperties) error
	updateStagedDirectorPropertiesMutex       sync.RWMutex
	updateStagedDirectorPropertiesArgsForCall []struct {
		arg1 api.DirectorProperties
	}
	updateStagedDirectorPropertiesReturns struct {
		result1 error
	}
	updateStagedDirectorPropertiesReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ApplyChangesService) CreateInstallation(arg1 bool, arg2 bool, arg3 bool, arg4 []string, arg5 api.ApplyErrandChanges) (api.InstallationsServiceOutput, error) {
	var arg4Copy []string
	if arg4 != nil {
		arg4Copy = make([]string, len(arg4))
		copy(arg4Copy, arg4)
	}
	fake.createInstallationMutex.Lock()
	ret, specificReturn := fake.createInstallationReturnsOnCall[len(fake.createInstallationArgsForCall)]
	fake.createInstallationArgsForCall = append(fake.createInstallationArgsForCall, struct {
		arg1 bool
		arg2 bool
		arg3 bool
		arg4 []string
		arg5 api.ApplyErrandChanges
	}{arg1, arg2, arg3, arg4Copy, arg5})
	fake.recordInvocation("CreateInstallation", []interface{}{arg1, arg2, arg3, arg4Copy, arg5})
	fake.createInstallationMutex.Unlock()
	if fake.CreateInstallationStub != nil {
		return fake.CreateInstallationStub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.createInstallationReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) CreateInstallationCallCount() int {
	fake.createInstallationMutex.RLock()
	defer fake.createInstallationMutex.RUnlock()
	return len(fake.createInstallationArgsForCall)
}

func (fake *ApplyChangesService) CreateInstallationCalls(stub func(bool, bool, bool, []string, api.ApplyErrandChanges) (api.InstallationsServiceOutput, error)) {
	fake.createInstallationMutex.Lock()
	defer fake.createInstallationMutex.Unlock()
	fake.CreateInstallationStub = stub
}

func (fake *ApplyChangesService) CreateInstallationArgsForCall(i int) (bool, bool, bool, []string, api.ApplyErrandChanges) {
	fake.createInstallationMutex.RLock()
	defer fake.createInstallationMutex.RUnlock()
	argsForCall := fake.createInstallationArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *ApplyChangesService) CreateInstallationReturns(result1 api.InstallationsServiceOutput, result2 error) {
	fake.createInstallationMutex.Lock()
	defer fake.createInstallationMutex.Unlock()
	fake.CreateInstallationStub = nil
	fake.createInstallationReturns = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) CreateInstallationReturnsOnCall(i int, result1 api.InstallationsServiceOutput, result2 error) {
	fake.createInstallationMutex.Lock()
	defer fake.createInstallationMutex.Unlock()
	fake.CreateInstallationStub = nil
	if fake.createInstallationReturnsOnCall == nil {
		fake.createInstallationReturnsOnCall = make(map[int]struct {
			result1 api.InstallationsServiceOutput
			result2 error
		})
	}
	fake.createInstallationReturnsOnCall[i] = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) GetInstallation(arg1 int) (api.InstallationsServiceOutput, error) {
	fake.getInstallationMutex.Lock()
	ret, specificReturn := fake.getInstallationReturnsOnCall[len(fake.getInstallationArgsForCall)]
	fake.getInstallationArgsForCall = append(fake.getInstallationArgsForCall, struct {
		arg1 int
	}{arg1})
	fake.recordInvocation("GetInstallation", []interface{}{arg1})
	fake.getInstallationMutex.Unlock()
	if fake.GetInstallationStub != nil {
		return fake.GetInstallationStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getInstallationReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) GetInstallationCallCount() int {
	fake.getInstallationMutex.RLock()
	defer fake.getInstallationMutex.RUnlock()
	return len(fake.getInstallationArgsForCall)
}

func (fake *ApplyChangesService) GetInstallationCalls(stub func(int) (api.InstallationsServiceOutput, error)) {
	fake.getInstallationMutex.Lock()
	defer fake.getInstallationMutex.Unlock()
	fake.GetInstallationStub = stub
}

func (fake *ApplyChangesService) GetInstallationArgsForCall(i int) int {
	fake.getInstallationMutex.RLock()
	defer fake.getInstallationMutex.RUnlock()
	argsForCall := fake.getInstallationArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ApplyChangesService) GetInstallationReturns(result1 api.InstallationsServiceOutput, result2 error) {
	fake.getInstallationMutex.Lock()
	defer fake.getInstallationMutex.Unlock()
	fake.GetInstallationStub = nil
	fake.getInstallationReturns = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) GetInstallationReturnsOnCall(i int, result1 api.InstallationsServiceOutput, result2 error) {
	fake.getInstallationMutex.Lock()
	defer fake.getInstallationMutex.Unlock()
	fake.GetInstallationStub = nil
	if fake.getInstallationReturnsOnCall == nil {
		fake.getInstallationReturnsOnCall = make(map[int]struct {
			result1 api.InstallationsServiceOutput
			result2 error
		})
	}
	fake.getInstallationReturnsOnCall[i] = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) GetInstallationLogs(arg1 int) (api.InstallationsServiceOutput, error) {
	fake.getInstallationLogsMutex.Lock()
	ret, specificReturn := fake.getInstallationLogsReturnsOnCall[len(fake.getInstallationLogsArgsForCall)]
	fake.getInstallationLogsArgsForCall = append(fake.getInstallationLogsArgsForCall, struct {
		arg1 int
	}{arg1})
	fake.recordInvocation("GetInstallationLogs", []interface{}{arg1})
	fake.getInstallationLogsMutex.Unlock()
	if fake.GetInstallationLogsStub != nil {
		return fake.GetInstallationLogsStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getInstallationLogsReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) GetInstallationLogsCallCount() int {
	fake.getInstallationLogsMutex.RLock()
	defer fake.getInstallationLogsMutex.RUnlock()
	return len(fake.getInstallationLogsArgsForCall)
}

func (fake *ApplyChangesService) GetInstallationLogsCalls(stub func(int) (api.InstallationsServiceOutput, error)) {
	fake.getInstallationLogsMutex.Lock()
	defer fake.getInstallationLogsMutex.Unlock()
	fake.GetInstallationLogsStub = stub
}

func (fake *ApplyChangesService) GetInstallationLogsArgsForCall(i int) int {
	fake.getInstallationLogsMutex.RLock()
	defer fake.getInstallationLogsMutex.RUnlock()
	argsForCall := fake.getInstallationLogsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ApplyChangesService) GetInstallationLogsReturns(result1 api.InstallationsServiceOutput, result2 error) {
	fake.getInstallationLogsMutex.Lock()
	defer fake.getInstallationLogsMutex.Unlock()
	fake.GetInstallationLogsStub = nil
	fake.getInstallationLogsReturns = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) GetInstallationLogsReturnsOnCall(i int, result1 api.InstallationsServiceOutput, result2 error) {
	fake.getInstallationLogsMutex.Lock()
	defer fake.getInstallationLogsMutex.Unlock()
	fake.GetInstallationLogsStub = nil
	if fake.getInstallationLogsReturnsOnCall == nil {
		fake.getInstallationLogsReturnsOnCall = make(map[int]struct {
			result1 api.InstallationsServiceOutput
			result2 error
		})
	}
	fake.getInstallationLogsReturnsOnCall[i] = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) Info() (api.Info, error) {
	fake.infoMutex.Lock()
	ret, specificReturn := fake.infoReturnsOnCall[len(fake.infoArgsForCall)]
	fake.infoArgsForCall = append(fake.infoArgsForCall, struct {
	}{})
	fake.recordInvocation("Info", []interface{}{})
	fake.infoMutex.Unlock()
	if fake.InfoStub != nil {
		return fake.InfoStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.infoReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) InfoCallCount() int {
	fake.infoMutex.RLock()
	defer fake.infoMutex.RUnlock()
	return len(fake.infoArgsForCall)
}

func (fake *ApplyChangesService) InfoCalls(stub func() (api.Info, error)) {
	fake.infoMutex.Lock()
	defer fake.infoMutex.Unlock()
	fake.InfoStub = stub
}

func (fake *ApplyChangesService) InfoReturns(result1 api.Info, result2 error) {
	fake.infoMutex.Lock()
	defer fake.infoMutex.Unlock()
	fake.InfoStub = nil
	fake.infoReturns = struct {
		result1 api.Info
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) InfoReturnsOnCall(i int, result1 api.Info, result2 error) {
	fake.infoMutex.Lock()
	defer fake.infoMutex.Unlock()
	fake.InfoStub = nil
	if fake.infoReturnsOnCall == nil {
		fake.infoReturnsOnCall = make(map[int]struct {
			result1 api.Info
			result2 error
		})
	}
	fake.infoReturnsOnCall[i] = struct {
		result1 api.Info
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) ListInstallations() ([]api.InstallationsServiceOutput, error) {
	fake.listInstallationsMutex.Lock()
	ret, specificReturn := fake.listInstallationsReturnsOnCall[len(fake.listInstallationsArgsForCall)]
	fake.listInstallationsArgsForCall = append(fake.listInstallationsArgsForCall, struct {
	}{})
	fake.recordInvocation("ListInstallations", []interface{}{})
	fake.listInstallationsMutex.Unlock()
	if fake.ListInstallationsStub != nil {
		return fake.ListInstallationsStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.listInstallationsReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) ListInstallationsCallCount() int {
	fake.listInstallationsMutex.RLock()
	defer fake.listInstallationsMutex.RUnlock()
	return len(fake.listInstallationsArgsForCall)
}

func (fake *ApplyChangesService) ListInstallationsCalls(stub func() ([]api.InstallationsServiceOutput, error)) {
	fake.listInstallationsMutex.Lock()
	defer fake.listInstallationsMutex.Unlock()
	fake.ListInstallationsStub = stub
}

func (fake *ApplyChangesService) ListInstallationsReturns(result1 []api.InstallationsServiceOutput, result2 error) {
	fake.listInstallationsMutex.Lock()
	defer fake.listInstallationsMutex.Unlock()
	fake.ListInstallationsStub = nil
	fake.listInstallationsReturns = struct {
		result1 []api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) ListInstallationsReturnsOnCall(i int, result1 []api.InstallationsServiceOutput, result2 error) {
	fake.listInstallationsMutex.Lock()
	defer fake.listInstallationsMutex.Unlock()
	fake.ListInstallationsStub = nil
	if fake.listInstallationsReturnsOnCall == nil {
		fake.listInstallationsReturnsOnCall = make(map[int]struct {
			result1 []api.InstallationsServiceOutput
			result2 error
		})
	}
	fake.listInstallationsReturnsOnCall[i] = struct {
		result1 []api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) RunningInstallation() (api.InstallationsServiceOutput, error) {
	fake.runningInstallationMutex.Lock()
	ret, specificReturn := fake.runningInstallationReturnsOnCall[len(fake.runningInstallationArgsForCall)]
	fake.runningInstallationArgsForCall = append(fake.runningInstallationArgsForCall, struct {
	}{})
	fake.recordInvocation("RunningInstallation", []interface{}{})
	fake.runningInstallationMutex.Unlock()
	if fake.RunningInstallationStub != nil {
		return fake.RunningInstallationStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.runningInstallationReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ApplyChangesService) RunningInstallationCallCount() int {
	fake.runningInstallationMutex.RLock()
	defer fake.runningInstallationMutex.RUnlock()
	return len(fake.runningInstallationArgsForCall)
}

func (fake *ApplyChangesService) RunningInstallationCalls(stub func() (api.InstallationsServiceOutput, error)) {
	fake.runningInstallationMutex.Lock()
	defer fake.runningInstallationMutex.Unlock()
	fake.RunningInstallationStub = stub
}

func (fake *ApplyChangesService) RunningInstallationReturns(result1 api.InstallationsServiceOutput, result2 error) {
	fake.runningInstallationMutex.Lock()
	defer fake.runningInstallationMutex.Unlock()
	fake.RunningInstallationStub = nil
	fake.runningInstallationReturns = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) RunningInstallationReturnsOnCall(i int, result1 api.InstallationsServiceOutput, result2 error) {
	fake.runningInstallationMutex.Lock()
	defer fake.runningInstallationMutex.Unlock()
	fake.RunningInstallationStub = nil
	if fake.runningInstallationReturnsOnCall == nil {
		fake.runningInstallationReturnsOnCall = make(map[int]struct {
			result1 api.InstallationsServiceOutput
			result2 error
		})
	}
	fake.runningInstallationReturnsOnCall[i] = struct {
		result1 api.InstallationsServiceOutput
		result2 error
	}{result1, result2}
}

func (fake *ApplyChangesService) UpdateStagedDirectorProperties(arg1 api.DirectorProperties) error {
	fake.updateStagedDirectorPropertiesMutex.Lock()
	ret, specificReturn := fake.updateStagedDirectorPropertiesReturnsOnCall[len(fake.updateStagedDirectorPropertiesArgsForCall)]
	fake.updateStagedDirectorPropertiesArgsForCall = append(fake.updateStagedDirectorPropertiesArgsForCall, struct {
		arg1 api.DirectorProperties
	}{arg1})
	fake.recordInvocation("UpdateStagedDirectorProperties", []interface{}{arg1})
	fake.updateStagedDirectorPropertiesMutex.Unlock()
	if fake.UpdateStagedDirectorPropertiesStub != nil {
		return fake.UpdateStagedDirectorPropertiesStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.updateStagedDirectorPropertiesReturns
	return fakeReturns.result1
}

func (fake *ApplyChangesService) UpdateStagedDirectorPropertiesCallCount() int {
	fake.updateStagedDirectorPropertiesMutex.RLock()
	defer fake.updateStagedDirectorPropertiesMutex.RUnlock()
	return len(fake.updateStagedDirectorPropertiesArgsForCall)
}

func (fake *ApplyChangesService) UpdateStagedDirectorPropertiesCalls(stub func(api.DirectorProperties) error) {
	fake.updateStagedDirectorPropertiesMutex.Lock()
	defer fake.updateStagedDirectorPropertiesMutex.Unlock()
	fake.UpdateStagedDirectorPropertiesStub = stub
}

func (fake *ApplyChangesService) UpdateStagedDirectorPropertiesArgsForCall(i int) api.DirectorProperties {
	fake.updateStagedDirectorPropertiesMutex.RLock()
	defer fake.updateStagedDirectorPropertiesMutex.RUnlock()
	argsForCall := fake.updateStagedDirectorPropertiesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ApplyChangesService) UpdateStagedDirectorPropertiesReturns(result1 error) {
	fake.updateStagedDirectorPropertiesMutex.Lock()
	defer fake.updateStagedDirectorPropertiesMutex.Unlock()
	fake.UpdateStagedDirectorPropertiesStub = nil
	fake.updateStagedDirectorPropertiesReturns = struct {
		result1 error
	}{result1}
}

func (fake *ApplyChangesService) UpdateStagedDirectorPropertiesReturnsOnCall(i int, result1 error) {
	fake.updateStagedDirectorPropertiesMutex.Lock()
	defer fake.updateStagedDirectorPropertiesMutex.Unlock()
	fake.UpdateStagedDirectorPropertiesStub = nil
	if fake.updateStagedDirectorPropertiesReturnsOnCall == nil {
		fake.updateStagedDirectorPropertiesReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateStagedDirectorPropertiesReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ApplyChangesService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createInstallationMutex.RLock()
	defer fake.createInstallationMutex.RUnlock()
	fake.getInstallationMutex.RLock()
	defer fake.getInstallationMutex.RUnlock()
	fake.getInstallationLogsMutex.RLock()
	defer fake.getInstallationLogsMutex.RUnlock()
	fake.infoMutex.RLock()
	defer fake.infoMutex.RUnlock()
	fake.listInstallationsMutex.RLock()
	defer fake.listInstallationsMutex.RUnlock()
	fake.runningInstallationMutex.RLock()
	defer fake.runningInstallationMutex.RUnlock()
	fake.updateStagedDirectorPropertiesMutex.RLock()
	defer fake.updateStagedDirectorPropertiesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ApplyChangesService) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
