// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package common

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildAddr(t *testing.T) {
	addr := BuildAddr("test.xyz", "8000")

	if addr != "http://test.xyz:8000" {
		t.Errorf("Expected 'http://test.xyz:8000' but got: %s", addr)
	}
}

func ExampleBuildAddr() {
	fmt.Println(strings.Index("chicken", "ken"))
	fmt.Println(strings.Index("chicken", "dmr"))
	// Output:
	// 4
	// -1
}
