// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package assert provides some assertion functions for testing.

Return values: true if the assert holds, false otherwise.
*/
package assert

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Set this to false to avoid include file position in logs.
var IncludeFilePosition = true

func isTestFuncName(name string) bool {
	p := strings.LastIndex(name, ".")
	if p < 0 {
		return false
	}

	name = name[p+1:]
	return strings.HasPrefix(name, "Test")
}

// Default: skip == 0
func assertPos(skip int) string {
	if !IncludeFilePosition {
		return ""
	}

	res := ""
	for i := 0; i < 5; i++ {
		pc, file, line, ok := runtime.Caller(skip + 2)
		if !ok {
			return ""
		}

		res = fmt.Sprintf("%s:%d: ", path.Base(file), line) + res

		if isTestFuncName(runtime.FuncForPC(pc).Name()) {
			break
		}

		skip++
	}
	return res
}

func Equal(t testing.TB, name string, act, exp interface{}) bool {
	if act != exp {
		expTp, actTp := reflect.ValueOf(exp).Type(), reflect.ValueOf(act).Type()
		var expMsg, actMsg string
		if expTp == actTp {
			expMsg, actMsg = fmt.Sprintf("%q", fmt.Sprint(exp)), fmt.Sprintf("%q", fmt.Sprint(act))
		} else {
			expMsg = fmt.Sprintf("%q(type %v)", fmt.Sprint(exp), expTp)
			actMsg = fmt.Sprintf("%q(type %v)", fmt.Sprint(act), actTp)
		}
		msg := fmt.Sprintf("%s%s is expected to be %s, but got %s", assertPos(0), name, expMsg, actMsg)
		if len(msg) >= 80 {
			msg = fmt.Sprintf("%s%s is expected to be\n  %s\nbut got\n  %s", assertPos(0), name, expMsg, actMsg)
		}
		t.Error(msg)
		return false
	}
	return true
}

// @param expToFunc could be a func with a single input value and a bool return, or a bool value directly.
func ValueShould(t testing.TB, name string, act interface{}, expToFunc interface{}, descIfFailed string) bool {
	expFunc := reflect.ValueOf(expToFunc)
	actValue := reflect.ValueOf(act)
	var succ bool
	if expFunc.Kind() == reflect.Bool {
		succ = expFunc.Bool()
	} else if expFunc.Kind() == reflect.Func {
		if expFunc.Type().NumIn() != 1 {
			t.Errorf("%sassert: expToFunc must have one parameter", assertPos(0))
			return false
		}

		if expFunc.Type().NumOut() != 1 {
			t.Errorf("%sassert: expToFunc must have one return value", assertPos(0))
			return false
		}

		if expFunc.Type().Out(0).Kind() != reflect.Bool {
			t.Errorf("%sassert: expToFunc must return a bool", assertPos(0))
			return false
		}

		succ = expFunc.Call([]reflect.Value{actValue})[0].Bool()
	} else {
		t.Errorf("%sassert: expToFunc must be a func or a bool", assertPos(0))
		return false
	}

	if !succ {
		t.Errorf("%s%s %s: %q(type %v)", assertPos(0), name, descIfFailed,
			fmt.Sprint(act), actValue.Type())
	}
	return succ
}

func NotEqual(t testing.TB, name string, act, exp interface{}) bool {
	if act == exp {
		t.Errorf("%s%s is not expected to be %q", assertPos(0), name, fmt.Sprint(exp))
		return false
	}
	return true
}

func True(t testing.TB, name string, act bool) bool {
	if !act {
		t.Errorf("%s%s unexpectedly got false", assertPos(0), name)
	}
	return act
}

func Should(t testing.TB, vl bool, showIfFailed string) bool {
	if !vl {
		t.Errorf("%s%s", assertPos(0), showIfFailed)
	}
	return vl
}

func False(t testing.TB, name string, act bool) bool {
	if act {
		t.Errorf("%s%s unexpectedly got true", assertPos(0), name)
	}
	return !act
}

func StringEqual(t testing.TB, name string, act, exp interface{}) bool {
	actS, expS := fmt.Sprintf("%+v", act), fmt.Sprintf("%+v", exp)
	if actS != expS {
		msg := fmt.Sprintf("%s%s is expected to be %q, but got %q", assertPos(0), name,
			fmt.Sprint(exp), fmt.Sprint(act))
		if len(msg) >= 80 {
			msg = fmt.Sprintf("%s%s is expected to be\n  %q\nbut got\n  %q", assertPos(0), name,
				fmt.Sprint(exp), fmt.Sprint(act))
		}
		t.Error(msg)
		return false
	}
	return true
}

func NoError(t testing.TB, err error) bool {
	if err != nil {
		t.Errorf("%s%v", assertPos(0), err)
		return false
	}
	return true
}

func Panic(t testing.TB, name string, f func()) bool {
	if !func() (res bool) {
		defer func() {
			res = recover() != nil
		}()

		f()
		return
	}() {
		t.Errorf("%s%s does not panic as expected.", assertPos(0), name)
		return false
	}

	return true
}
