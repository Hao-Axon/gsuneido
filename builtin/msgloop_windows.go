// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

// +build !portable

package builtin

import (
	"runtime"

	"github.com/apmckinlay/gsuneido/builtin/goc"
	. "github.com/apmckinlay/gsuneido/runtime"
	"golang.org/x/sys/windows"
)

var _ = builtin("MessageLoop(hwnd = 0)", func(t *Thread, args []Value) Value {
	goc.MessageLoop(uintptr(ToInt(args[0])))
	return nil
})

var uiThreadId uint32

func init() {
	// inject dependencies
	goc.Callback2 = callback2
	goc.Callback3 = callback3
	goc.Callback4 = callback4
	goc.SunAPP = sunAPP
	goc.UpdateUI = updateUI2
	UpdateUI = updateUI // runtime
	Interrupt = goc.Interrupt

	// used to detect calls from other threads (not allowed)
	runtime.LockOSThread()
	uiThreadId = windows.GetCurrentThreadId()

	goc.Init()
}

func Run() {
	goc.Run()
}

type timerSpec struct {
	hwnd Value
	id   Value
	ms   Value // nil for KillTimer
	cb   Value
	ret  chan Value
}

// timerChan is used for cross thread SetTimer and KillTimer requests
var timerChan = make(chan timerSpec, 1)
