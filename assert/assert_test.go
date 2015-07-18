// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package assert

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/golangplus/testing"
)

func TestFilePosition(t *testing.T) {
	var b bytes.Buffer
	bt := &testingp.WriterTB{
		Writer: &b,
	}

	Equal(bt, "v", 1, 2)
	line := 23 // the line number of the last line
	Equal(t, "log", b.String(), fmt.Sprintf("assert_test.go:%d: v is expected to be \"2\", but got \"1\"\n", line))

	b.Reset()
	Panic(bt, "nonpanic", func() {})
	line = 28 // the line number of the last line
	Equal(t, "log", b.String(), fmt.Sprintf("assert_test.go:%d: nonpanic does not panic as expected.\n", line))

	func(outLine int) {
		b.Reset()
		Equal(bt, "v", 1, 2)
		line := 34 // the line number of the last line
		Equal(t, "log", b.String(), fmt.Sprintf("assert_test.go:%d: assert_test.go:%d: v is expected to be \"2\", but got \"1\"\n", outLine, line))
	}(37) // 37 is the line number of current line
}

func TestSuccess(t *testing.T) {
	True(t, "return value", Equal(t, "v", 1, 1))
	True(t, "return true", Equal(t, "slice", []int{1}, []int{1}))
	True(t, "return true", Equal(t, "map", map[int]int{2: 1}, map[int]int{2: 1}))
	True(t, "return value", ValueShould(t, "s", "abc", func(s string) bool {
		return s == "abc"
	}, "is not abc"))
	True(t, "return value", ValueShould(t, "s", "abc", true, "is not abc"))
	True(t, "return value", NotEqual(t, "v", 1, 4))
	True(t, "return value", True(t, "bool", true))
	True(t, "return value", Should(t, true, "failed"))
	True(t, "return value", False(t, "bool", false))
	True(t, "return value", StringEqual(t, "string", 1, "1"))
}

func ExampleFailed() {
	// Turn off file position because it is hard to check.
	IncludeFilePosition = false

	t := &testingp.WriterTB{
		Writer: os.Stdout,
	}
	Equal(t, "v", 1, 2)
	Equal(t, "v", 1, "1")
	Equal(t, "v", 1, "2")
	Equal(t, "m", map[int]int{1: 2, 3: 4}, map[int]int{2: 1, 3: 5})
	ValueShould(t, "s", "\xff\xfe\xfd", utf8.ValidString, "is not valid UTF8")
	ValueShould(t, "s", "abcd", len("abcd") <= 3, "has more than 3 bytes")
	NotEqual(t, "v", 1, 1)
	True(t, "v", false)
	Should(t, false, "failed")
	False(t, "v", true)
	StringEqual(t, "s", 1, "2")
	StringEqual(t, "s", []int{2, 3}, []string{"1", "2"})
	Panic(t, "nonpanic", func() {})

	// OUTPUT:
	// v is expected to be "2", but got "1"
	// v is expected to be "1"(type string), but got "1"(type int)
	// v is expected to be "2"(type string), but got "1"(type int)
	// Unexpected m: both 2 entries
	//   Difference(expected ---  actual +++)
	//     --- "2": "1"
	//     --- "3": "5"
	//     +++ "3": "4"
	//     +++ "1": "2"
	// s is not valid UTF8: "\xff\xfe\xfd"(type string)
	// s has more than 3 bytes: "abcd"(type string)
	// v is not expected to be "1"
	// v unexpectedly got false
	// failed
	// v unexpectedly got true
	// s is expected to be "2", but got "1"
	// Unexpected s: both 2 lines
	//   Difference(expected ---  actual +++)
	//     ---   1: "1"
	//     +++   2: "3"
	// nonpanic does not panic as expected.
}

func TestPanic(t *testing.T) {
	True(t, "return value", Panic(t, "panic", func() {
		panic("error")
	}))
}
