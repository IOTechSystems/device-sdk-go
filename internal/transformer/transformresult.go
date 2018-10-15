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
	"math"
	"strconv"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/model"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TransformReadResult(cv *model.CommandValue, pv models.PropertyValue) error {
	var err error

	if pv.Base != "" {
		err = transformReadBase(cv, pv.Base)
		if err != nil {
			return err
		}
	}

	if pv.Scale != "" {
		err = transformReadScale(cv, pv.Scale)
		if err != nil {
			return err
		}
	}

	if pv.Offset != "" {
		err = transformReadOffset(cv, pv.Offset)
	}
	return err
}

func transformReadBase(cv *model.CommandValue, base string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the base %s of PropertyValue cannot be parsed to float64: %v", base, err))
		return err
	}

	v = math.Pow(b, v)
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func transformReadScale(cv *model.CommandValue, scale string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	s, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", scale, err))
		return err
	}

	v = v * s
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func transformReadOffset(cv *model.CommandValue, offset string) error {
	v, err := commandValueToFloat64(cv)
	if err != nil {
		return err
	}
	o, err := strconv.ParseFloat(offset, 64)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to float64: %v", offset, err))
		return err
	}

	v = v + o
	err = replaceCommandValueFromFloat64(cv, v)
	return err
}

func commandValueToFloat64(cv *model.CommandValue) (float64, error) {
	var value float64
	var err error = nil
	switch cv.Type {
	case model.Uint8:
		v, err := cv.Uint8Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Uint16:
		v, err := cv.Uint16Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Uint32:
		v, err := cv.Uint32Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Uint64:
		v, err := cv.Uint64Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Int8:
		v, err := cv.Int8Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Int16:
		v, err := cv.Int16Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Int32:
		v, err := cv.Int32Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Int64:
		v, err := cv.Int64Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Float32:
		v, err := cv.Float32Value()
		if err != nil {
			return 0, err
		}
		value = float64(v)
	case model.Float64:
		value, err = cv.Float64Value()
		if err != nil {
			return 0, err
		}
	default:
		err = fmt.Errorf("wrong data type of CommandValue to convert to float64: %s", cv.String())
	}
	return value, nil
}

func replaceCommandValueFromFloat64(cv *model.CommandValue, f64 float64) error {
	value, err := convertNumericFromFloat64(f64, cv.Type)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("binary.Write failed: %v", err))
	} else {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func convertNumericFromFloat64(f64 float64, t model.ValueType) (interface{}, error) {
	switch t {
	case model.Uint8:
		return uint8(f64), nil
	case model.Uint16:
		return uint16(f64), nil
	case model.Uint32:
		return uint32(f64), nil
	case model.Uint64:
		return uint64(f64), nil
	case model.Int8:
		return int8(f64), nil
	case model.Int16:
		return int16(f64), nil
	case model.Int32:
		return int32(f64), nil
	case model.Int64:
		return int64(f64), nil
	case model.Float32:
		return float32(f64), nil
	case model.Float64:
		return f64, nil
	default:
		return 0, fmt.Errorf("wrong data type of CommandValue to convert from float64: %v", t)
	}
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

func MapCommandValue(value *model.CommandValue) (*model.CommandValue, bool) {
	mappings := value.RO.Mappings
	newValue, ok := mappings[value.ValueToString()]
	var result *model.CommandValue
	if ok {
		result = model.NewStringValue(value.RO, value.Origin, newValue)
	}
	return result, ok
}
