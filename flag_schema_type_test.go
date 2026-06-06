package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlag_SchemaType_Bool(t *testing.T) {
	f := &BoolFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "boolean", st.SchemaType())
	_, ok = any(f).(SchemaItemsTyper)
	assert.True(t, ok)
}

func TestFlag_SchemaType_String(t *testing.T) {
	f := &StringFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "string", st.SchemaType())
	_, ok = any(f).(SchemaItemsTyper)
	assert.True(t, ok)
}

func TestFlag_SchemaType_Int(t *testing.T) {
	flags := []Flag{&IntFlag{}, &Int8Flag{}, &Int16Flag{}, &Int32Flag{}, &Int64Flag{}}
	for _, f := range flags {
		st, ok := f.(SchemaTyper)
		assert.True(t, ok)
		assert.Equal(t, "integer", st.SchemaType())
	}
}

func TestFlag_SchemaType_Uint(t *testing.T) {
	flags := []Flag{&UintFlag{}, &Uint8Flag{}, &Uint16Flag{}, &Uint32Flag{}, &Uint64Flag{}}
	for _, f := range flags {
		st, ok := f.(SchemaTyper)
		assert.True(t, ok)
		assert.Equal(t, "integer", st.SchemaType())
	}
}

func TestFlag_SchemaType_Float(t *testing.T) {
	flags := []Flag{&FloatFlag{}, &Float32Flag{}, &Float64Flag{}}
	for _, f := range flags {
		st, ok := f.(SchemaTyper)
		assert.True(t, ok)
		assert.Equal(t, "number", st.SchemaType())
	}
}

func TestFlag_SchemaType_Duration(t *testing.T) {
	f := &DurationFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "string", st.SchemaType())
}

func TestFlag_SchemaType_Timestamp(t *testing.T) {
	f := &TimestampFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "string", st.SchemaType())
}

func TestFlag_SchemaType_Slice(t *testing.T) {
	flags := []Flag{
		&StringSliceFlag{},
		&IntSliceFlag{},
		&FloatSliceFlag{},
	}
	for _, f := range flags {
		st, ok := f.(SchemaTyper)
		assert.True(t, ok)
		assert.Equal(t, "array", st.SchemaType())
	}
}

func TestFlag_SchemaItemsType_Slice(t *testing.T) {
	tests := []struct {
		flag     Flag
		itemType string
	}{
		{&StringSliceFlag{}, "string"},
		{&IntSliceFlag{}, "integer"},
		{&Int8SliceFlag{}, "integer"},
		{&Int16SliceFlag{}, "integer"},
		{&Int32SliceFlag{}, "integer"},
		{&Int64SliceFlag{}, "integer"},
		{&UintSliceFlag{}, "integer"},
		{&Uint8SliceFlag{}, "integer"},
		{&Uint16SliceFlag{}, "integer"},
		{&Uint32SliceFlag{}, "integer"},
		{&Uint64SliceFlag{}, "integer"},
		{&FloatSliceFlag{}, "number"},
		{&Float32SliceFlag{}, "number"},
		{&Float64SliceFlag{}, "number"},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			st, ok := tc.flag.(SchemaItemsTyper)
			assert.True(t, ok)
			assert.Equal(t, tc.itemType, st.SchemaItemsType())
		})
	}
}

func TestFlag_SchemaType_Map(t *testing.T) {
	f := &StringMapFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "object", st.SchemaType())
	sit, ok := any(f).(SchemaItemsTyper)
	assert.True(t, ok)
	assert.Equal(t, "", sit.SchemaItemsType())
}

func TestFlag_SchemaType_Generic(t *testing.T) {
	f := &GenericFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "", st.SchemaType())
}

func TestFlag_SchemaType_BoolWithInverse(t *testing.T) {
	f := &BoolWithInverseFlag{}
	st, ok := any(f).(SchemaTyper)
	assert.True(t, ok)
	assert.Equal(t, "boolean", st.SchemaType())
	sit, ok := any(f).(SchemaItemsTyper)
	assert.True(t, ok)
	assert.Equal(t, "", sit.SchemaItemsType())
}

func TestFlag_SchemaType_NonSliceItemsType(t *testing.T) {
	flags := []Flag{
		&BoolFlag{},
		&StringFlag{},
		&IntFlag{},
		&FloatFlag{},
		&DurationFlag{},
		&TimestampFlag{},
	}
	for _, f := range flags {
		sit, ok := f.(SchemaItemsTyper)
		assert.True(t, ok)
		assert.Equal(t, "", sit.SchemaItemsType())
	}
}

func TestFlag_SchemaType_PreservesPrecision(t *testing.T) {
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &TimestampFlag{Config: TimestampConfig{Layouts: []string{time.RFC3339}}, Value: created}
	assert.Equal(t, "string", f.SchemaType())

	f2 := &DurationFlag{Value: 5 * time.Second}
	assert.Equal(t, "string", f2.SchemaType())
}
