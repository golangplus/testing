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
	"testing"
)

// Set this to false to avoid include file position in logs.
var IncludeFilePosition = true

// Default: skip == 0
func assertPos(skip int) string {
	if !IncludeFilePosition {
		return ""
	}

	_, file, line, ok := runtime.Caller(skip + 2)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d: ", path.Base(file), line)
}

func Equal(t testing.TB, name string, act, exp interface{}) bool {
	if act != exp {
		expTp, actTp := reflect.ValueOf(exp).Type(), reflect.ValueOf(act).Type()
		if expTp == actTp {
			t.Errorf("%s%s is expected to be %q, but got %q", assertPos(0), name,
				fmt.Sprint(exp), fmt.Sprint(act))
		} else {
			t.Errorf("%s%s is expected to be %q(type %v), but got %q(type %v)",
				assertPos(0), name, fmt.Sprint(exp), expTp, fmt.Sprint(act), actTp)
		}
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
		if len(actS)+len(expS) < 70 {
			t.Errorf("%s%s is expected to be %q, but got %q", assertPos(0), name,
				fmt.Sprint(exp), fmt.Sprint(act))
			return false
		} else {
			t.Errorf("%s%s is expected to be\n%q\n, but got\n%q", assertPos(0), name,
				fmt.Sprint(exp), fmt.Sprint(act))
			return false
		}
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
