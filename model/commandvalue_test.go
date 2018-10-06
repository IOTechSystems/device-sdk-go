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
	"math"
	"testing"
	"time"
)

// Test NewBoolValue function.
func TestNewBoolValue(t *testing.T) {
	var value bool

	cr := NewBoolValue(0, value)

	if cr.Type != Bool {
		t.Errorf("NewBoolValue: invalid Type: %v", cr.Type)
	}

	if value == true {
		t.Errorf("NewBoolValue: invalid value: true")
	}

	v, err := cr.BoolValue()
	if err != nil {
		t.Errorf("NewBoolValue: failed to get bool value")
	}
	if v != value {
		t.Errorf("NewBoolValue: bool value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("bool reading: %v\n", reading)
	if reading.Value != "false" {
		t.Errorf("NewBoolValue: invalid reading Value: %s", reading.Value)
	}

	value = true

	cr = NewBoolValue(0, value)

	if cr.Type != Bool {
		t.Errorf("NewBoolValue: invalid Type: %v #2", cr.Type)
	}

	if value == false {
		t.Errorf("NewBoolValue: invalid value: false")
	}

	v, err = cr.BoolValue()
	if err != nil {
		t.Errorf("NewBoolValue: failed to get bool value")
	}
	if v != value {
		t.Errorf("NewBoolValue: bool value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("bool reading: %v\n", reading)
	if reading.Value != "true" {
		t.Errorf("NewBoolValue: invalid reading Value: %s", reading.Value)
	}
}

// Test NewStringValue function.
func TestNewStringValue(t *testing.T) {
	var value string

	cr := NewStringValue(0, value)
	if cr.Type != String {
		t.Errorf("NewStringValue: invalid Type: %v", cr.Type)
	}

	v, err := cr.StringValue()
	if err != nil {
		t.Errorf("NewStringValue: failed to get string value")
	}
	if v != value {
		t.Errorf("NewStringValue: string value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("string reading: %v\n", reading)

	value = "this is a real string"
	cr = NewStringValue(0, value)
	if cr.Type != String {
		t.Errorf("NewStringValue: invalid Type: %v #2", cr.Type)
	}

	if value != cr.stringValue {
		t.Errorf("NewStringValue: cr.stringValue: %s doesn't match value: %s", cr.stringValue, value)
	}

	v, err = cr.StringValue()
	if err != nil {
		t.Errorf("NewStringValue: failed to get string value")
	}
	if v != value {
		t.Errorf("NewStringValue: string value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("string reading #2: %v\n", reading)
	if reading.Value != "this is a real string" {
		t.Errorf("NewStringValue: invalid reading Value: %s", reading.Value)
	}
}

// Test NewUint8Value function.
func TestNewUint8Value(t *testing.T) {
	var value uint8

	cr := NewUint8Value(0, value)
	if cr.Type != Uint8 {
		t.Errorf("NewUint8Value: invalid Type: %v", cr.Type)
	}

	var res uint8
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint8Value: cr.Uint8Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Uint8Value()
	if err != nil {
		t.Errorf("NewUint8Value: failed to get uint8 value")
	}
	if v != value {
		t.Errorf("NewUint8Value: uint8 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint8 reading: %v\n", reading)

	value = 42
	cr = NewUint8Value(0, value)
	if cr.Type != Uint8 {
		t.Errorf("NewUint8Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint8Value: cr.Uint8Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Uint8Value()
	if err != nil {
		t.Errorf("NewUint8Value: failed to get uint8 value")
	}
	if v != value {
		t.Errorf("NewUint8Value: uint8 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint8 reading #2: %v\n", reading)
	if reading.Value != "42" {
		t.Errorf("NewUint8Value: invalid reading Value: %s", reading.Value)
	}
}

// Test NewUint16Value function.
func TestNewUint16Value(t *testing.T) {
	var value uint16

	cr := NewUint16Value(0, value)
	if cr.Type != Uint16 {
		t.Errorf("NewUint16Value: invalid Type: %v", cr.Type)
	}

	var res uint16
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint16Value: cr.Uint16Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Uint16Value()
	if err != nil {
		t.Errorf("NewUint16Value: failed to get uint16 value")
	}
	if v != value {
		t.Errorf("NewUint16Value: uint16 value is incorrect")
	}

	value = 65535
	cr = NewUint16Value(0, value)
	if cr.Type != Uint16 {
		t.Errorf("NewUint16Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint16Value: cr.Uint16Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Uint16Value()
	if err != nil {
		t.Errorf("NewUint16Value: failed to get uint16 value")
	}
	if v != value {
		t.Errorf("NewUint16Value: uint16 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint16 reading: %v\n", reading)
	if reading.Value != "65535" {
		t.Errorf("NewUint16Value: invalid reading Value: %s", reading.Value)
	}
}

// Test NewUint32Value function.
func TestNewUint32Value(t *testing.T) {
	var value uint32

	cr := NewUint32Value(0, value)
	if cr.Type != Uint32 {
		t.Errorf("NewUint32Value: invalid Type: %v", cr.Type)
	}

	var res uint32
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint32Value: cr.Uint32Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Uint32Value()
	if err != nil {
		t.Errorf("NewUint32Value: failed to get uint32 value")
	}
	if v != value {
		t.Errorf("NewUint32Value: uint32 value is incorrect")
	}

	value = 4294967295
	cr = NewUint32Value(0, value)
	if cr.Type != Uint32 {
		t.Errorf("NewUint32Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint32Value: cr.Uint32Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Uint32Value()
	if err != nil {
		t.Errorf("NewUint32Value: failed to get uint32 value")
	}
	if v != value {
		t.Errorf("NewUint32Value: uint32 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint32 reading: %v\n", reading)
	if reading.Value != "4294967295" {
		t.Errorf("NewUint32Value: invalid reading Value: %s", reading.Value)
	}
}

// Test NewUint64Value function.
func TestNewUint64Value(t *testing.T) {
	var value uint64
	var origin int64 = 42

	cr := NewUint64Value(origin, value)
	if cr.Type != Uint64 {
		t.Errorf("NewUint64Value: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewUint64Value: invalid Origin: %d", cr.Origin)
	}

	var res uint64
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint64Value: cr.Uint64Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Uint64Value()
	if err != nil {
		t.Errorf("NewUint64Value: failed to get uint64 value")
	}
	if v != value {
		t.Errorf("NewUint64Value: uint64 value is incorrect")
	}

	value = 18446744073709551615

	cr = NewUint64Value(0, value)
	if cr.Type != Uint64 {
		t.Errorf("NewUint64Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint64Value: cr.Uint64Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Uint64Value()
	if err != nil {
		t.Errorf("NewUint64Value: failed to get uint64 value")
	}
	if v != value {
		t.Errorf("NewUint64Value: uint64 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint64 reading: %v\n", reading)
	if reading.Value != "18446744073709551615" {
		t.Errorf("NewUint64Value: invalid reading Value: %s", reading.Value)
	}
}

// Test NewInt8Value function.
func TestNewInt8Value(t *testing.T) {
	var value int8 = -128

	cr := NewInt8Value(0, value)
	if cr.Type != Int8 {
		t.Errorf("NewInt8Value: invalid Type: %v", cr.Type)
	}

	var res int8
	buf := bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt8Value: cr.Int8Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Int8Value()
	if err != nil {
		t.Errorf("NewInt8Value: failed to get int8 value")
	}
	if v != value {
		t.Errorf("NewInt8Value: int8 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int8 reading: %v\n", reading)
	if reading.Value != "-128" {
		t.Errorf("NewInt8Value #1: invalid reading Value: %s", reading.Value)
	}

	value = 127
	cr = NewInt8Value(0, value)
	if cr.Type != Int8 {
		t.Errorf("NewInt8Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt8Value: cr.Int8Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Int8Value()
	if err != nil {
		t.Errorf("NewInt8Value: failed to get int8 value")
	}
	if v != value {
		t.Errorf("NewInt8Value: int8 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int8 reading: %v\n", reading)
	if reading.Value != "127" {
		t.Errorf("NewInt8Value #2: invalid reading Value: %s", reading.Value)
	}
}

// Test NewInt16Value function.
func TestNewInt16Value(t *testing.T) {
	var value int16 = -32768

	cr := NewInt16Value(0, value)
	if cr.Type != Int16 {
		t.Errorf("NewInt16Value: invalid Type: %v", cr.Type)
	}

	var res int16
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt16Value: cr.Int16Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Int16Value()
	if err != nil {
		t.Errorf("NewInt16Value: failed to get int16 value")
	}
	if v != value {
		t.Errorf("NewInt16Value: int16 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int16 reading: %v\n", reading)
	if reading.Value != "-32768" {
		t.Errorf("NewInt16Value #1: invalid reading Value: %s", reading.Value)
	}

	value = 32767
	cr = NewInt16Value(0, value)
	if cr.Type != Int16 {
		t.Errorf("NewInt16Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt16Value: cr.Int16Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Int16Value()
	if err != nil {
		t.Errorf("NewInt16Value: failed to get int16 value")
	}
	if v != value {
		t.Errorf("NewInt16Value: int16 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int16 reading: %v\n", reading)
	if reading.Value != "32767" {
		t.Errorf("NewInt16Value #2: invalid reading Value: %s", reading.Value)
	}
}

// Test NewInt32Value function.
func TestNewInt32Value(t *testing.T) {
	var value int32 = -2147483648

	cr := NewInt32Value(0, value)
	if cr.Type != Int32 {
		t.Errorf("NewInt32Value: invalid Type: %v", cr.Type)
	}

	var res int32
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt32Value: cr.Int32Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Int32Value()
	if err != nil {
		t.Errorf("NewInt32Value: failed to get int32 value")
	}
	if v != value {
		t.Errorf("NewInt32Value: int32 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int32 reading: %v\n", reading)
	if reading.Value != "-2147483648" {
		t.Errorf("NewInt32Value #1: invalid reading Value: %s", reading.Value)
	}

	value = 2147483647
	cr = NewInt32Value(0, value)
	if cr.Type != Int32 {
		t.Errorf("NewInt32Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt32Value: cr.Int32Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Int32Value()
	if err != nil {
		t.Errorf("NewInt32Value: failed to get int32 value")
	}
	if v != value {
		t.Errorf("NewInt32Value: int32 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int32 reading: %v\n", reading)
	if reading.Value != "2147483647" {
		t.Errorf("NewInt32Value #2: invalid reading Value: %s", reading.Value)
	}
}

// Test NewInt64Value function.
func TestNewInt64Value(t *testing.T) {
	var value int64 = -9223372036854775808
	var origin int64 = 42

	cr := NewInt64Value(origin, value)
	if cr.Type != Int64 {
		t.Errorf("NewInt64Value: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewInt64Value: invalid Origin: %d", cr.Origin)
	}

	var res int64
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt64Value: cr.Int64Value: %d doesn't match value: %d", value, res)
	}

	v, err := cr.Int64Value()
	if err != nil {
		t.Errorf("NewInt64Value: failed to get int64 value")
	}
	if v != value {
		t.Errorf("NewInt64Value: int64 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int64 reading: %v\n", reading)
	if reading.Value != "-9223372036854775808" {
		t.Errorf("NewInt64Value #1: invalid reading Value: %s", reading.Value)
	}

	value = 9223372036854775807

	cr = NewInt64Value(0, value)
	if cr.Type != Int64 {
		t.Errorf("NewInt64Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt64Value: cr.Int64Value: %d doesn't match value: %d (#2)", value, res)
	}

	v, err = cr.Int64Value()
	if err != nil {
		t.Errorf("NewInt64Value: failed to get int64 value")
	}
	if v != value {
		t.Errorf("NewInt64Value: int64 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("int64 reading: %v\n", reading)
	if reading.Value != "9223372036854775807" {
		t.Errorf("NewInt64Value #2: invalid reading Value: %s", reading.Value)
	}
}

// Test NewFloat32Value function.
func TestNewFloat32Value(t *testing.T) {
	var value float32 = math.SmallestNonzeroFloat32
	var origin int64 = time.Now().UnixNano() / int64(time.Millisecond)

	cr := NewFloat32Value(origin, value)
	if cr.Type != Float32 {
		t.Errorf("NewFloat32Value: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewFloat32Value: invalid Origin: %d", cr.Origin)
	}

	var res float32
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat32Value: cr.Int64Value: %v doesn't match value: %v", value, res)
	}

	v, err := cr.Float32Value()
	if err != nil {
		t.Errorf("NewFloat32Value: failed to get float32 value")
	}
	if v != value {
		t.Errorf("NewFloat32Value: float32 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("float32 reading: %v\n", reading)
	if reading.Value != "0.000000000000000000000000000000000000000000001" {
		t.Errorf("NewFloat32Value #1: invalid reading Value: %s", reading.Value)
	}

	value = math.MaxFloat32

	cr = NewFloat32Value(0, value)
	if cr.Type != Float32 {
		t.Errorf("NewFloat32Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat32Value: cr.Float32Value: %v doesn't match value: %v (#2)", value, res)
	}

	v, err = cr.Float32Value()
	if err != nil {
		t.Errorf("NewFloat32Value: failed to get float32 value")
	}
	if v != value {
		t.Errorf("NewFloat32Value: float32 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("float32 reading: %v\n", reading)
	if reading.Value != "340282350000000000000000000000000000000" {
		t.Errorf("NewFloat32Value #2: invalid reading Value: %s", reading.Value)
	}
}

// Test NewFloat64Value function.
func TestNewFloat64Value(t *testing.T) {
	var value float64 = math.SmallestNonzeroFloat64
	var origin int64 = time.Now().UnixNano() / int64(time.Millisecond)

	cr := NewFloat64Value(origin, value)
	if cr.Type != Float64 {
		t.Errorf("NewFloat64Value: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewFloat64Value: invalid Origin: %d", cr.Origin)
	}

	var res float64
	buf := bytes.NewReader(cr.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat64Value: cr.Int64Value: %v doesn't match value: %v", value, res)
	}

	v, err := cr.Float64Value()
	if err != nil {
		t.Errorf("NewFloat64Value: failed to get float64 value")
	}
	if v != value {
		t.Errorf("NewFloat64Value: float64 value is incorrect")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("float64 reading: %v\n", reading)
	if reading.Value != "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005" {
		t.Errorf("NewFloat64Value #1: invalid reading Value: %s", reading.Value)
	}

	value = math.MaxFloat64

	cr = NewFloat64Value(0, value)
	if cr.Type != Float64 {
		t.Errorf("NewFloat64Value: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericValue)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat64Value: cr.Float64Value: %v doesn't match value: %v (#2)", value, res)
	}

	v, err = cr.Float64Value()
	if err != nil {
		t.Errorf("NewFloat64Value: failed to get float64 value")
	}
	if v != value {
		t.Errorf("NewFloat64Value: float64 value is incorrect")
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("float64 reading: %v\n", reading)
	if reading.Value != "179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("NewFloat64Value #2: invalid reading Value: %s", reading.Value)
	}
}
