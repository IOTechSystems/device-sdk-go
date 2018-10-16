// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"
	"strconv"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TransformWriteParameter(cv *model.CommandValue, pv models.PropertyValue) error {
	var err error
	if cv.Type == model.String || cv.Type == model.Bool {
		return nil // do nothing for String and Bool
	}

	if pv.Offset != "" {
		err = transformWriteOffset(cv, pv.Offset)
		if err != nil {
			return err
		}
	}

	if pv.Scale != "" {
		err = transformWriteScale(cv, pv.Scale)
		if err != nil {
			return err
		}
	}

	if pv.Base != "" {
		err = transformWriteBase(cv, pv.Base)
	}
	return err
}

func transformWriteBase(cv *model.CommandValue, base string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", base, err))
		return err
	} else if b == 0 {
		return nil // do nothing if Base = 0
	}

	v = math.Log(v) / math.Log(b)
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func transformWriteScale(cv *model.CommandValue, scale string) error {
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

func transformWriteOffset(cv *model.CommandValue, offset string) error {
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
