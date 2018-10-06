// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// ValueType indicates the type of result being passed back
// from a ProtocolDriver instance.
type ValueType int

const (
	// Bool indicates that the result is a bool,
	// stored in CommandValue's boolRes member.
	Bool ValueType = iota
	// String indicates that the result is a string,
	// stored in CommandValue's stringRes member.
	String
	// Uint8 indicates that the result is a uint8 that
	// is stored in CommandValue's NumericRes member.
	Uint8
	// Uint16 indicates that the result is a uint16 that
	// is stored in CommandValue's NumericRes member.
	Uint16
	// Uint32 indicates that the result is a uint32 that
	// is stored in CommandValue's NumericRes member.
	Uint32
	// Uint64 indicates that the result is a uint64 that
	// is stored in CommandValue's NumericRes member.
	Uint64
	// Int8 indicates that the result is a int8 that
	// is stored in CommandValue's NumericRes member.
	Int8
	// Int16 indicates that the result is a int16 that
	// is stored in CommandValue's NumericRes member.
	Int16
	// Int32 indicates that the result is a int32 that
	// is stored in CommandValue's NumericRes member.
	Int32
	// Int64 indicates that the result is a int64 that
	// is stored in CommandValue's NumericRes member.
	Int64
	// Float32 indicates that the result is a float32 that
	// is stored in CommandValue's NumericRes member.
	Float32
	// Float64 indicates that the result is a float64 that
	// is stored in CommandValue's NumericRes member.
	Float64
)

type CommandValue struct {
	// Origin is an int64 value which indicates the time the reading
	// contained in the CommandValue was read by the ProtocolDriver
	// instance.
	Origin int64
	// Type is a ValueType value which indicates what type of
	// result was returned from the ProtocolDriver instance in
	// response to HandleCommand being called to handle a single
	// ResourceOperation.
	Type ValueType
	// NumericValue is a byte slice with a maximum capacity of
	// 64 bytes, used to hold a numeric result returned by a
	// ProtocolDriver instance. The value can be converted to
	// its native type by referring to the the value of ResType.
	NumericValue []byte
	// stringValue is a string value returned as a result by a ProtocolDriver instance.
	stringValue string
}

func NewBoolValue(origin int64, value bool) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Bool}
	encodeResult(cr, value)
	fmt.Printf("result: %v\n", cr)
	return
}

func NewStringValue(origin int64, value string) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: String, stringValue: value}

	fmt.Printf("result: %v\n", cr)
	return
}

// NewUint8Value creates a CommandValue of Type Uint8 with the given value.
func NewUint8Value(origin int64, value uint8) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Uint8}
	encodeResult(cr, value)
	return
}

// NewUint16Value creates a CommandValue of Type Uint16 with the given value.
func NewUint16Value(origin int64, value uint16) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Uint16}
	encodeResult(cr, value)
	return
}

// NewUint32Value creates a CommandValue of Type Uint32 with the given value.
func NewUint32Value(origin int64, value uint32) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Uint32}
	encodeResult(cr, value)
	return
}

// NewUint64Value creates a CommandValue of Type Uint64 with the given value.
func NewUint64Value(origin int64, value uint64) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Uint64}
	encodeResult(cr, value)
	return
}

// NewInt8Value creates a CommandValue of Type Int8 with the given value.
func NewInt8Value(origin int64, value int8) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Int8}
	encodeResult(cr, value)
	return
}

// NewInt16Value creates a CommandValue of Type Int16 with the given value.
func NewInt16Value(origin int64, value int16) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Int16}
	encodeResult(cr, value)
	return
}

// NewInt32Value creates a CommandValue of Type Int32 with the given value.
func NewInt32Value(origin int64, value int32) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Int32}
	encodeResult(cr, value)
	return
}

// NewInt64Value creates a CommandValue of Type Int64 with the given value.
func NewInt64Value(origin int64, value int64) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Int64}
	encodeResult(cr, value)
	return
}

// NewFloat32Value creates a CommandValue of Type Float32 with the given value.
func NewFloat32Value(origin int64, value float32) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Float32}
	encodeResult(cr, value)
	return
}

// NewFloat64Value creates a CommandValue of Type Float64 with the given value.
func NewFloat64Value(origin int64, value float64) (cr *CommandValue) {
	cr = &CommandValue{Origin: origin, Type: Float64}
	encodeResult(cr, value)
	return
}

func encodeResult(cr *CommandValue, value interface{}) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v", err)
	}

	cr.NumericValue = buf.Bytes()
}

func decodeResult(reader io.Reader, value interface{}) error {
	err := binary.Read(reader, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Read failed: %v", err)
	}
	return err
}

// Reading returns a new Reading instance created from the the given CommandValue.
func (cr *CommandValue) Reading(devName string, vdName string) *models.Reading {

	reading := &models.Reading{Name: vdName, Device: devName}
	reading.Value = cr.toString()

	// if result has a non-zero Origin, use it
	if cr.Origin > 0 {
		reading.Origin = cr.Origin
	} else {
		reading.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	}

	return reading
}

// String returns a string representation of a CommandValue instance.
func (cr *CommandValue) toString() (str string) {
	if cr.Type == String {
		str = cr.stringValue
		return
	}

	reader := bytes.NewReader(cr.NumericValue)

	switch cr.Type {
	case Bool:
		var res bool
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatBool(res)
	case Uint8:
		var res uint8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint16:
		var res uint16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint32:
		var res uint32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint64:
		var res uint64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(res, 10)
	case Int8:
		var res int8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int16:
		var res int16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int32:
		var res int32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int64:
		var res int64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(res, 10)
	case Float32:
		var res float32
		binary.Read(reader, binary.BigEndian, &res)
		str = strconv.FormatFloat(float64(res), 'f', -1, 32)
	case Float64:
		var res float64
		binary.Read(reader, binary.BigEndian, &res)
		str = strconv.FormatFloat(res, 'f', -1, 64)
	}

	return
}

// String returns a string representation of a CommandValue instance.
func (cr *CommandValue) String() (str string) {

	originStr := fmt.Sprintf("%d\n", cr.Origin)

	var typeStr string

	switch cr.Type {
	case Bool:
		typeStr = "Bool: "
	case String:
		typeStr = "String: "
	case Uint8:
		typeStr = "Uint8: "
	case Uint16:
		typeStr = "Uint16: "
	case Uint32:
		typeStr = "Uint32: "
	case Uint64:
		typeStr = "Uint64: "
	case Int8:
		typeStr = "Int8: "
	case Int16:
		typeStr = "Int16: "
	case Int32:
		typeStr = "Int32: "
	case Int64:
		typeStr = "Int64: "
	case Float32:
		typeStr = "Float32: "
	case Float64:
		typeStr = "Float64: "
	}

	resultStr := typeStr + cr.toString()

	str = originStr + resultStr

	return
}

// TransformResult applies transforms specified in the given
// PropertyValue instance.
func (cr *CommandValue) TransformResult(models.PropertyValue) bool {

	// TODO: implement base, scale, offset & assertions
	return true
}

func (cr *CommandValue) BoolValue() (bool, error) {
	var result bool
	if cr.Type != Bool {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) StringValue() (string, error) {
	result := cr.stringValue
	if cr.Type != String {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	return result, nil
}

func (cr *CommandValue) Uint8Value() (uint8, error) {
	var result uint8
	if cr.Type != Uint8 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Uint16Value() (uint16, error) {
	var result uint16
	if cr.Type != Uint16 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Uint32Value() (uint32, error) {
	var result uint32
	if cr.Type != Uint32 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Uint64Value() (uint64, error) {
	var result uint64
	if cr.Type != Uint64 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Int8Value() (int8, error) {
	var result int8
	if cr.Type != Int8 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Int16Value() (int16, error) {
	var result int16
	if cr.Type != Int16 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Int32Value() (int32, error) {
	var result int32
	if cr.Type != Int32 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Int64Value() (int64, error) {
	var result int64
	if cr.Type != Int64 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Float32Value() (float32, error) {
	var result float32
	if cr.Type != Float32 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}

func (cr *CommandValue) Float64Value() (float64, error) {
	var result float64
	if cr.Type != Float64 {
		return result, fmt.Errorf("the data type is not %T", result)
	}
	err := decodeResult(bytes.NewReader(cr.NumericValue), &result)
	return result, err
}
