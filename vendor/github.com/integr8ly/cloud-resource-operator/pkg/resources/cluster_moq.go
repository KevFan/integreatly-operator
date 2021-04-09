// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	"sync"
)

// Ensure, that PodCommanderMock does implement PodCommander.
// If this is not the case, regenerate this file with moq.
var _ PodCommander = &PodCommanderMock{}

// PodCommanderMock is a mock implementation of PodCommander.
//
// 	func TestSomethingThatUsesPodCommander(t *testing.T) {
//
// 		// make and configure a mocked PodCommander
// 		mockedPodCommander := &PodCommanderMock{
// 			ExecIntoPodFunc: func(dpl *appsv1.Deployment, cmd string) error {
// 				panic("mock out the ExecIntoPod method")
// 			},
// 		}
//
// 		// use mockedPodCommander in code that requires PodCommander
// 		// and then make assertions.
//
// 	}
type PodCommanderMock struct {
	// ExecIntoPodFunc mocks the ExecIntoPod method.
	ExecIntoPodFunc func(dpl *appsv1.Deployment, cmd string) error

	// calls tracks calls to the methods.
	calls struct {
		// ExecIntoPod holds details about calls to the ExecIntoPod method.
		ExecIntoPod []struct {
			// Dpl is the dpl argument value.
			Dpl *appsv1.Deployment
			// Cmd is the cmd argument value.
			Cmd string
		}
	}
	lockExecIntoPod sync.RWMutex
}

// ExecIntoPod calls ExecIntoPodFunc.
func (mock *PodCommanderMock) ExecIntoPod(dpl *appsv1.Deployment, cmd string) error {
	if mock.ExecIntoPodFunc == nil {
		panic("PodCommanderMock.ExecIntoPodFunc: method is nil but PodCommander.ExecIntoPod was just called")
	}
	callInfo := struct {
		Dpl *appsv1.Deployment
		Cmd string
	}{
		Dpl: dpl,
		Cmd: cmd,
	}
	mock.lockExecIntoPod.Lock()
	mock.calls.ExecIntoPod = append(mock.calls.ExecIntoPod, callInfo)
	mock.lockExecIntoPod.Unlock()
	return mock.ExecIntoPodFunc(dpl, cmd)
}

// ExecIntoPodCalls gets all the calls that were made to ExecIntoPod.
// Check the length with:
//     len(mockedPodCommander.ExecIntoPodCalls())
func (mock *PodCommanderMock) ExecIntoPodCalls() []struct {
	Dpl *appsv1.Deployment
	Cmd string
} {
	var calls []struct {
		Dpl *appsv1.Deployment
		Cmd string
	}
	mock.lockExecIntoPod.RLock()
	calls = mock.calls.ExecIntoPod
	mock.lockExecIntoPod.RUnlock()
	return calls
}
