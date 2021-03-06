// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import "fmt"

var printBuiltin = &SuBuiltinRaw{printBuiltinFn, BuiltinParams{ParamSpec: ParamSpecAt}}

func printBuiltinFn(t *Thread, as *ArgSpec, args []Value) Value {
	iter := NewArgsIter(as, args)
	sep := ""
	for {
		name, value := iter()
		if value == nil {
			break
		}
		fmt.Print(sep)
		if name != nil {
			print(t, name)
			fmt.Print(": ")
		}
		print(t, value)
		sep = " "
	}
	fmt.Println()
	return nil
}

func print(t *Thread, v Value) {
	if s, ok := v.ToStr(); ok {
		fmt.Print(s)
	} else {
		fmt.Print(Display(t, v))
	}
}

type Displayable interface {
	Display(t *Thread) string
}

func Display(t *Thread, val Value) string {
	if d, ok := val.(Displayable); ok {
		return d.Display(t)
	}
	if d, ok := val.(ToStringable); ok {
		return d.ToString(t)
	}
	return val.String()
}
