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
	"math"
	"strconv"
)

func TransformPutParameter(cv *model.CommandValue, pv models.PropertyValue) error {
	var err error

	if pv.Offset != "" {
		err = transformPutOffset(cv, pv.Offset)
		if err != nil {
			return err
		}
	}

	if pv.Scale != "" {
		err = transformPutScale(cv, pv.Scale)
		if err != nil {
			return err
		}
	}

	if pv.Base != "" {
		err = transformPutBase(cv, pv.Base)
	}
	return err
}

func transformPutBase(cv *model.CommandValue, base string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", scale, err))
		return err
	}

	v = math.Log (v) / math.Log (b)
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func transformPutScale(cv *model.CommandValue, scale string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	s, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", scale, err))
		return err
	}

	if s == 0 {
		return fmt.Errorf("scale is 0")
	}
	v = v / s
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func transformPutOffset(cv *model.CommandValue, offset string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	o, err := strconv.ParseFloat(offset, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to float64: %v", offset, err))
		return err
	}

	v = v - o
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}
