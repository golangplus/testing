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
	"strconv"
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
			t.Errorf("%s%s is expected to be %s, but got %v", assertPos(0), name,
				strconv.Quote(fmt.Sprint(exp)), strconv.Quote(fmt.Sprint(act)))
		} else {
			t.Errorf("%s%s is expected to be %s(type %v), but got %v(type %v)",
				assertPos(0), name,
				strconv.Quote(fmt.Sprint(exp)), expTp,
				strconv.Quote(fmt.Sprint(act)), actTp)
		}
		return false
	}
	return true
}

func NotEqual(t testing.TB, name string, act, exp interface{}) bool {
	if act == exp {
		t.Errorf("%s%s is not expected to be %s", assertPos(0), name,
			strconv.Quote(fmt.Sprint(exp)))
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
			t.Errorf("%s%s is expected to be %s, but got %v", assertPos(0), name,
				strconv.Quote(fmt.Sprint(exp)), strconv.Quote(fmt.Sprint(act)))
			return false
		} else {
			t.Errorf("%s%s is expected to be\n%s\n, but got\n%v", assertPos(0), name,
				strconv.Quote(fmt.Sprint(exp)), strconv.Quote(fmt.Sprint(act)))
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
