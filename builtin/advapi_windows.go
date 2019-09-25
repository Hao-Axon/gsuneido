package builtin

import (
	"unsafe"

	. "github.com/apmckinlay/gsuneido/runtime"
	"golang.org/x/sys/windows"
)

var advapi32 = windows.NewLazyDLL("advapi32.dll")

// RegOpenKeyEx
var regOpenKeyEx = advapi32.NewProc("RegOpenKeyExA")
var _ = builtin5("RegOpenKeyEx(hKey, lpSubKey, ulOptions, samDesired, phkResult)",
	func(a, b, c, d, e Value) Value {
		var e1 uintptr
		rtn, _, _ := regOpenKeyEx.Call(
			intArg(a),
			uintptr(stringArg(b)),
			intArg(c),
			intArg(d),
			uintptr(unsafe.Pointer(&e1)))
		e.Put(nil, SuStr("x"), IntVal(int(e1))) // phkResult
		return intRet(rtn)
	})

// RegCloseKey
var regCloseKey = advapi32.NewProc("RegCloseKey")
var _ = builtin1("RegCloseKey(hKey)",
	func(a Value) Value {
		rtn, _, _ := regCloseKey.Call(intArg(a))
		return intRet(rtn)
	})

// RegCreateKeyEx
var regCreateKeyEx = advapi32.NewProc("RegCreateKeyExA")
var _ = builtin("RegCreateKeyEx(hKey, lpSubKey, Reserved/*unused*/, lpClass, "+
	"dwOptions, samDesired, lpSecurityAttributes, phkResult, lpdwDisposition)",
	func(_ *Thread, a []Value) Value {
		var h1 uintptr
		rtn, _, _ := regCreateKeyEx.Call(
			intArg(a[0]),
			uintptr(stringArg(a[1])),
			0, // Reserved - must be 0
			uintptr(stringArg(a[3])),
			intArg(a[4]),
			intArg(a[5]),
			0, // lpSecurityAttributes - always null
			uintptr(unsafe.Pointer(&h1)),
			0) // lpdwDisposition - always null
		a[7].Put(nil, SuStr("x"), IntVal(int(h1))) // phkResult
		return intRet(rtn)
	})

// RegQueryValueEx - hard coded for 4 byte data
var regQueryValueEx = advapi32.NewProc("RegQueryValueExA")
var _ = builtin6("RegQueryValueEx(hKey, lpValueName, lpReserved/*unused*/, "+
	"lpType/*unused*/, lpData, lpcbData/*unused*/)",
	func(a, b, c, d, e, f Value) Value {
		var e1 int32   // data
		f1 := int32(4) // cbData = 4 to match int32 data
		rtn, _, _ := regQueryValueEx.Call(
			intArg(a),
			uintptr(stringArg(b)),
			0,                            // lpReserved - must be 0
			0,                            // lpType - NULL
			uintptr(unsafe.Pointer(&e1)), // lpData
			uintptr(unsafe.Pointer(&f1))) // lpcbData
		e.Put(nil, SuStr("x"), IntVal(int(e1))) // data
		return intRet(rtn)
	})

// RegSetValueEx - hard coded for 4 byte data
var regSetValueEx = advapi32.NewProc("RegSetValueExA")
var _ = builtin6("RegSetValueEx(hKey, lpValueName, reserved/*unused*/, "+
	"dwType/*unused*/, lpData, cbData/*unused*/)",
	func(a, b, c, d, e, f Value) Value {
		var e1 int32 // data
		rtn, _, _ := regSetValueEx.Call(
			intArg(a),
			uintptr(stringArg(b)),
			0,                            // reserved - must be 0
			intArg(d),                    // dwType
			uintptr(unsafe.Pointer(&e1)), // lpData
			4)                            // cbData = 4 to match int32 data
		e.Put(nil, SuStr("x"), IntVal(int(e1)))
		return intRet(rtn)
	})
