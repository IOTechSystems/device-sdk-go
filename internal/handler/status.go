// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"io"
	"net/http"
)

func StatusHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "pong")
}
