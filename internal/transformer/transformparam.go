// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"strconv"
)

func TransformPutParameter(cv *model.CommandValue, pv models.PropertyValue) error {
	v, err := strconv.ParseFloat(cv.ValueToString(), 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the CommandValue %s cannot be parsed to float64 for calculation: %v", cv.String(), err))
		return err
	}

	if pv.Scale != "" {
		v, err = transformPutScale(v, pv.Scale)
	}

	if pv.Offset != "" {
		v, err = transformPutOffset(v, pv.Offset)
	}

	err = replaceNumericValue(cv, v)
	return err
}

func transformPutScale(v float64, scale string) (float64, error) {
	s, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", scale, err))
		return v, err
	}

	if s == 0 {
		return v, fmt.Errorf("scale is 0")
	}
	v = v / s
	return v, err
}

func transformPutOffset(v float64, offset string) (float64, error) {
	o, err := strconv.ParseFloat(offset, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to float64: %v", offset, err))
		return v, err
	}

	v = v - o
	return v, err
}
