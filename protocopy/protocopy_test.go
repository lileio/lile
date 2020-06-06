package protocopy_test

import (
	"testing"
	"time"

	"github.com/lileio/lile/v2/protocopy"
	"github.com/lileio/lile/v2/protocopy/test"
	"github.com/stretchr/testify/assert"
)

type Nested struct {
	Name   string
	Amount uint32
	Ok     string
}

type OriginalBitOfEverything struct {
	SingleNested             Nested
	UUID                     string
	Nested                   []*Nested
	FloatValue               float32
	DoubleValue              float64
	Int64Value               int64
	Uint64Value              uint64
	Int32Value               int32
	Fixed64Value             uint64
	Fixed32Value             uint32
	BoolValue                bool
	StringValue              string
	BytesValue               []byte
	Uint32Value              uint32
	EnumValue                string
	EnumValueInt             int
	Sfixed32Value            int32
	Sfixed64Value            int64
	Sint32Value              int32
	Sint64Value              int64
	RepeatedStringValue      []string
	MapValue                 map[string]string
	MappedStringValue        map[string]string
	MappedNestedValue        map[string]*Nested
	NonConventionalNameValue string
	TimestampValue           time.Time
	TimestampValuePointer    *time.Time
	DurationValue            time.Duration
	// repeated values. they are comma-separated in path
	PathRepeatedFloatValue    []float32
	PathRepeatedDoubleValue   []float64
	PathRepeatedInt64Value    []int64
	PathRepeatedUint64Value   []uint64
	PathRepeatedInt32Value    []int32
	PathRepeatedFixed64Value  []uint64
	PathRepeatedFixed32Value  []uint32
	PathRepeatedBoolValue     []bool
	PathRepeatedStringValue   []string
	PathRepeatedBytesValue    [][]byte
	PathRepeatedUint32Value   []uint32
	PathRepeatedEnumValue     []string
	PathRepeatedSfixed32Value []int32
	PathRepeatedSfixed64Value []int64
	PathRepeatedSint32Value   []int32
	PathRepeatedSint64Value   []int64
}

func TestABitOfEverything(t *testing.T) {
	tm := time.Now()
	testObj := OriginalBitOfEverything{
		SingleNested: Nested{
			Name:   "Alex B",
			Amount: 32,
			Ok:     test.ABitOfEverything_Nested_FALSE.String(),
		},
		UUID:                "123e4567-e89b-12d3-a456-426614174000",
		Nested:              []*Nested{{Name: "Alex B"}},
		FloatValue:          1.5,
		DoubleValue:         2.5,
		Int64Value:          4294967296,
		Uint64Value:         9223372036854775807,
		Int32Value:          -2147483648,
		Fixed64Value:        9223372036854775807,
		Fixed32Value:        4294967295,
		BoolValue:           true,
		StringValue:         "somelovely string",
		BytesValue:          []byte("strings are basically bytes"),
		EnumValue:           test.NumericEnum_ONE.String(),
		EnumValueInt:        int(test.NumericEnum_ONE),
		Uint32Value:         4294967295,
		Sfixed32Value:       2147483647,
		Sfixed64Value:       -4611686018427387904,
		Sint32Value:         2147483647,
		Sint64Value:         4611686018427387903,
		RepeatedStringValue: []string{"hi", "there"},
		MapValue: map[string]string{
			"1": test.NumericEnum_ONE.String(),
			"0": test.NumericEnum_ZERO.String(),
		},
		MappedStringValue: map[string]string{
			"help": "test",
			"you":  "222test",
		},
		MappedNestedValue: map[string]*Nested{
			"somekey": {
				Name:   "Alex B",
				Amount: 20,
				Ok:     test.ABitOfEverything_Nested_TRUE.String(),
			},
		},
		NonConventionalNameValue: "camelCase",
		TimestampValue:           time.Now(),
		TimestampValuePointer:    &tm,
		DurationValue:            time.Minute,
		PathRepeatedFloatValue: []float32{
			1.5,
			-1.5,
		},
		PathRepeatedDoubleValue: []float64{
			2.5,
			-2.5,
		},
		PathRepeatedInt64Value: []int64{
			4294967296,
			-4294967296,
		},
		PathRepeatedUint64Value: []uint64{
			0,
			9223372036854775807,
		},
		PathRepeatedInt32Value: []int32{
			2147483647,
			-2147483648,
		},
		PathRepeatedFixed64Value: []uint64{
			0,
			9223372036854775807,
		},
		PathRepeatedFixed32Value: []uint32{
			0,
			4294967295,
		},
		PathRepeatedBoolValue: []bool{
			true,
			false,
		},
		PathRepeatedStringValue: []string{
			"foo",
			"bar",
		},
		PathRepeatedBytesValue: [][]byte{
			[]byte{0x00},
			[]byte{0xFF},
		},
		PathRepeatedUint32Value: []uint32{
			0,
			4294967295,
		},
		PathRepeatedEnumValue: []string{
			test.NumericEnum_ONE.String(),
			test.NumericEnum_ZERO.String(),
		},
		PathRepeatedSfixed32Value: []int32{
			2147483647,
			-2147483648,
		},
		PathRepeatedSfixed64Value: []int64{
			4294967296,
			-4294967296,
		},
		PathRepeatedSint32Value: []int32{
			2147483647,
			-2147483648,
		},
		PathRepeatedSint64Value: []int64{
			4611686018427387903,
			-4611686018427387904,
		},
	}

	out := &test.ABitOfEverything{}
	err := protocopy.ToProto(testObj, out)
	assert.Nil(t, err)

	// Test using a nil pointer too
	var outnil test.ABitOfEverything
	err = protocopy.ToProto(testObj, &outnil)
	assert.Nil(t, err)
}
