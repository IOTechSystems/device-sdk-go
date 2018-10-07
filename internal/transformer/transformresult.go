// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"math"
	"strconv"
)


func TransformGetResult(cv *model.CommandValue, pv models.PropertyValue) error {
	v, err := strconv.ParseFloat(cv.ValueToString(), 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the CommandValue %s cannot be parsed to float64 for calculation: %v", cv.String(), err))
		return err
	}

	if pv.Base != "" {
		v, err = transformGetBase(v, pv.Base)
	}

	if pv.Scale != "" {
		v, err = transformGetScale(v, pv.Scale)
	}

	if pv.Offset != "" {
		v, err = transformGetOffset(v, pv.Offset)
	}

	err = replaceNumericValue(cv, v)
	return err
}

func MapCommandValue(value *model.CommandValue, mappings map[string]string) (*model.CommandValue, bool) {
	newValue, ok := mappings[value.ValueToString()]
	var result *model.CommandValue
	if ok {
		result = model.NewStringValue(value.Origin, newValue)
	}
	return result, ok
}

func transformGetBase(v float64, base string) (float64, error) {
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the base %s of PropertyValue cannot be parsed to float64: %v", base, err))
		return v, err
	}

	v = math.Pow(b, v)
	return v, err
}

func transformGetScale(v float64, scale string) (float64, error) {
	s, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", scale, err))
		return v, err
	}

	v = v * s
	return v, err
}

func transformGetOffset(v float64, offset string) (float64, error) {
	o, err := strconv.ParseFloat(offset, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to float64: %v", offset, err))
		return v, err
	}

	v = v + o
	return v, err
}

func replaceNumericValue(cv *model.CommandValue, value interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("binary.Write failed: %v", err))
	} else {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func CheckAssertion(cv *model.CommandValue, assertion string, device *models.Device) error {
	if assertion != "" && cv.ValueToString() != assertion {
		device.OperatingState = models.Disabled
		go common.DevCli.UpdateOpStateByName(device.Name, models.Disabled)
		msg := fmt.Sprintf("assertion (%s) failed with value: %s", assertion, cv.ValueToString())
		common.LogCli.Error(msg)
		return fmt.Errorf(msg)
	}
	return nil
}
