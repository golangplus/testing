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
	"sort"
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

type sortInterfaceStruct struct {
	LenF  int
	LessF func(i, j int) bool
	SwapF func(i, j int)
}

func (is sortInterfaceStruct) Len() int           { return is.LenF }
func (is sortInterfaceStruct) Less(i, j int) bool { return is.LessF(i, j) }
func (is sortInterfaceStruct) Swap(i, j int)      { is.SwapF(i, j) }

func collectAndSortMapKeys(vl reflect.Value) (keys []reflect.Value, keyStrs []string) {
	keys = vl.MapKeys()
	keyStrs = make([]string, len(keys))
	for i, key := range keys {
		keyStrs[i] = fmt.Sprintf("%+v", key)
	}
	sort.Sort(sortInterfaceStruct{
		LenF: len(keys),
		LessF: func(i, j int) bool {
			return keyStrs[i] < keyStrs[j]
		},
		SwapF: func(i, j int) {
			keyStrs[i], keyStrs[j] = keyStrs[j], keyStrs[i]
			keys[i], keys[j] = keys[j], keys[i]
		},
	})
	return
}

func collectMapDiffKeys(act, exp reflect.Value, actKeys, expKeys []reflect.Value,
	actKeyStrs, expKeyStrs []string) (extraKeys, diffKeys, missingKeys []reflect.Value) {
	i, j := 0, 0
	for i < len(expKeys) && j < len(actKeys) {
		switch {
		case expKeyStrs[i] < actKeyStrs[j]:
			missingKeys = append(missingKeys, expKeys[i])
			i++

		case expKeyStrs[i] > actKeyStrs[j]:
			extraKeys = append(extraKeys, actKeys[j])
			j++

		default:
			expKeyStr, actKeyStr := expKeyStrs[i], actKeyStrs[j]
			// Processing keys with equal string representation. (They are not necessarily equal).
			for ; i < len(expKeys) && expKeyStrs[i] == actKeyStr; i++ {
				expKey := expKeys[i]
				actValue := act.MapIndex(expKey)
				if !actValue.IsValid() {
					// expKey does not exist in act.
					missingKeys = append(missingKeys, expKey)
					continue
				}
				expValue := exp.MapIndex(expKey)
				if reflect.DeepEqual(expValue.Interface(), actValue.Interface()) {
					continue
				}
				diffKeys = append(diffKeys, expKey)
			}
			for ; j < len(actKeys) && expKeyStr == actKeyStrs[j]; j++ {
				actKey := actKeys[j]
				expValue := exp.MapIndex(actKey)
				if expValue.IsValid() {
					// Should have handled in last loop
					continue
				}
				extraKeys = append(extraKeys, actKey)
			}
		}
	}
	missingKeys = append(missingKeys, expKeys[i:]...)
	extraKeys = append(extraKeys, actKeys[j:]...)
	return
}

func mapValueToStr(val reflect.Value) string {
	s := fmt.Sprintf("%+v", val.Interface())
	if s == "{}" {
		// If the value is an empty struct, the map is used as a set, not showing the value
		return ""
	}
	return ": " + fmt.Sprintf("%q", s)
}

func mapDiff(skip int, t testing.TB, name string, act, exp reflect.Value) {
	// Collect and sort map keys for exp and act.
	actKeys, actKeyStrs := collectAndSortMapKeys(act)
	expKeys, expKeyStrs := collectAndSortMapKeys(exp)

	// Collect extra/diff/missing keys.
	extraKeys, diffKeys, missingKeys :=
		collectMapDiffKeys(act, exp, actKeys, expKeys, actKeyStrs, expKeyStrs)

	// Output results
	title := fmt.Sprintf("%sUnexpected %s: ", assertPos(skip), name)
	if len(expKeys) == len(actKeys) {
		title = fmt.Sprintf("%sboth %d entries", title, len(expKeys))
	} else {
		title = fmt.Sprintf("%sexp %d, act %d entries", title, len(expKeys), len(actKeys))
	}
	t.Error(title)
	t.Log("  Difference(expected ---  actual +++)")

	if len(missingKeys) > 0 {
		for _, key := range missingKeys {
			t.Logf("    --- %q%s", fmt.Sprintf("%+v", key.Interface()), mapValueToStr(exp.MapIndex(key)))
		}
	}
	if len(diffKeys) > 0 {
		for _, key := range diffKeys {
			t.Logf("    --- %q%s", fmt.Sprintf("%+v", key.Interface()), mapValueToStr(exp.MapIndex(key)))
			t.Logf("    +++ %q%s", fmt.Sprintf("%+v", key.Interface()), mapValueToStr(act.MapIndex(key)))
		}
	}
	if len(extraKeys) > 0 {
		for _, key := range extraKeys {
			t.Logf("    +++ %q%s", fmt.Sprintf("%+v", key.Interface()), mapValueToStr(act.MapIndex(key)))
		}
	}
}

func sameTypeDiff(skip int, t testing.TB, name string, act, exp reflect.Value) {
	switch exp.Kind() {
	case reflect.Map:
		mapDiff(skip+1, t, name, act, exp)
		return
	case reflect.Slice:
	}

	expMsg := fmt.Sprintf("%q", fmt.Sprintf("%+v", exp.Interface()))
	actMsg := fmt.Sprintf("%q", fmt.Sprintf("%+v", act.Interface()))
	msg := fmt.Sprintf("%s%s is expected to be %s, but got %s", assertPos(skip), name, expMsg, actMsg)
	if len(msg) >= 80 {
		msg = fmt.Sprintf("%s%s is expected to be\n  %s\nbut got\n  %s", assertPos(skip), name, expMsg, actMsg)
	}
	t.Error(msg)
}

func safeValueType(vl reflect.Value) reflect.Type {
	if !vl.IsValid() {
		return nil
	}
	return vl.Type()
}

func Equal(t testing.TB, name string, act, exp interface{}) bool {
	if reflect.DeepEqual(exp, act) {
		return true
	}
	expVl, actVl := reflect.ValueOf(exp), reflect.ValueOf(act)

	if safeValueType(expVl) == safeValueType(actVl) {
		sameTypeDiff(1, t, name, actVl, expVl)
		return false
	}

	expMsg := fmt.Sprintf("%q(type %v)", fmt.Sprintf("%+v", exp), safeValueType(expVl))
	actMsg := fmt.Sprintf("%q(type %v)", fmt.Sprintf("%+v", act), safeValueType(actVl))
	msg := fmt.Sprintf("%s%s is expected to be %s, but got %s", assertPos(0), name, expMsg, actMsg)
	if len(msg) >= 80 {
		msg = fmt.Sprintf("%s%s is expected to be\n  %s\nbut got\n  %s", assertPos(0), name, expMsg, actMsg)
	}
	t.Error(msg)
	return false
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

func ShouldOrDie(t testing.TB, vl bool, showIfFailed string) {
	if !vl {
		t.Fatalf("%s%s", assertPos(0), showIfFailed)
	}
}

func False(t testing.TB, name string, act bool) bool {
	if act {
		t.Errorf("%s%s unexpectedly got true", assertPos(0), name)
	}
	return !act
}

func sliceToStrings(a reflect.Value) []string {
	l := make([]string, a.Len())
	for i := 0; i < a.Len(); i++ {
		l[i] = fmt.Sprintf("%+v", a.Index(i).Interface())
	}
	return l
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func linesEqual(skip int, t testing.TB, name string, act, exp reflect.Value) bool {
	actS, expS := sliceToStrings(act), sliceToStrings(exp)
	if stringSliceEqual(actS, expS) {
		return true
	}

	title := fmt.Sprintf("%sUnexpected %s: ", assertPos(skip), name)
	if len(expS) == len(actS) {
		title = fmt.Sprintf("%sboth %d lines", title, len(expS))
	} else {
		title = fmt.Sprintf("%sexp %d, act %d lines", title, len(expS), len(actS))
	}
	t.Error(title)
	t.Log("  Difference(expected ---  actual +++)")

	_, expMat, actMat := match(len(expS), len(actS), func(expI, actI int) int {
		if expS[expI] == actS[actI] {
			return 0
		}
		return 2
	}, func(int) int {
		return 1
	}, func(int) int {
		return 1
	})
	for i, j := 0, 0; i < len(expS) || j < len(actS); {
		switch {
		case j >= len(actS) || i < len(expS) && expMat[i] < 0:
			t.Logf("    --- %3d: %q", i+1, expS[i])
			i++
		case i >= len(expS) || j < len(actS) && actMat[j] < 0:
			t.Logf("    +++ %3d: %q", j+1, actS[j])
			j++
		default:
			if expS[i] != actS[j] {
				t.Logf("    --- %3d: %q", i+1, expS[i])
				t.Logf("    +++ %3d: %q", j+1, actS[j])
			} // else
			i++
			j++
		}
	}

	return false
}

// StringEqual compares the string representation of the values.
// If act and exp are both slices, they were matched by elements and the results are
// presented in a diff style (if not totally equal).
func StringEqual(t testing.TB, name string, act, exp interface{}) bool {
	actV, expV := reflect.ValueOf(act), reflect.ValueOf(exp)
	if actV.Kind() == reflect.Slice && expV.Kind() == reflect.Slice {
		return linesEqual(1, t, name, actV, expV)
	}

	actS, expS := fmt.Sprintf("%+v", act), fmt.Sprintf("%+v", exp)
	if actS == expS {
		return true
	}

	if strings.ContainsRune(actS, '\n') || strings.ContainsRune(expS, '\n') {
		return linesEqual(1, t, name,
			reflect.ValueOf(strings.Split(actS, "\n")),
			reflect.ValueOf(strings.Split(expS, "\n")))
	}

	msg := fmt.Sprintf("%s%s is expected to be %q, but got %q", assertPos(0), name,
		fmt.Sprint(exp), fmt.Sprint(act))
	if len(msg) >= 80 {
		msg = fmt.Sprintf("%s%s is expected to be\n  %q\nbut got\n  %q", assertPos(0), name,
			fmt.Sprint(exp), fmt.Sprint(act))
	}
	t.Error(msg)
	return false
}

func NoError(t testing.TB, err error) bool {
	if err != nil {
		t.Errorf("%s%v", assertPos(0), err)
		return false
	}
	return true
}

func NoErrorOrDie(t testing.TB, err error) {
	if err != nil {
		t.Fatalf("%s%v", assertPos(0), err)
	}
}

func Error(t testing.TB, err error) bool {
	if err == nil {
		t.Error("Expecting error but nil got!")
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
