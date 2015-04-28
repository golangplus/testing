// Copyright 2015 The Golang Plus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package assert

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/golangplus/testing"
)

func TestSuccess(t *testing.T) {
	True(t, "return value", Equal(t, "v", 1, 1))
	True(t, "return value", NotEqual(t, "v", 1, 4))
	True(t, "return value", True(t, "bool", true))
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
	NotEqual(t, "v", 1, 1)
	True(t, "v", false)
	False(t, "v", true)
	StringEqual(t, "s", 1, "2")

	// OUTPUT:
	// v is expected to be "2", but got "1"
	// v is expected to be "1"(type string), but got "1"(type int)
	// v is not expected to be "1"
	// v unexpectedly got false
	// v unexpectedly got true
	// s is expected to be "2", but got "1"
}

func nextLine() string {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", path.Base(file), line+1)
}

func TestFilePosition(t *testing.T) {
	var b bytes.Buffer
	bt := &testingp.WriterTB{
		Writer: &b,
	}

	pos := nextLine()
	Equal(bt, "v", 1, 2)
	Equal(t, "log", b.String(), fmt.Sprintf("%s: v is expected to be \"2\", but got \"1\"\n", pos))
}
