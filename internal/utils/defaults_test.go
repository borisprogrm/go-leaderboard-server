package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	type TestStruct struct {
		StringVal string `default:"test"`
		BoolVal   bool   `default:"true"`
		IntVal    int    `default:"-2147483647"`
		Int64Val  int64  `default:"-9223372036854775808"`
		Int32Val  int32  `default:"-2147483648"`
		Int16Val  int16  `default:"-32768"`
		Int8Val   int8   `default:"-128"`
		UIntVal   uint   `default:"4294967295"`
		UInt64Val uint64 `default:"18446744073709551615"`
		UInt32Val uint32 `default:"4294967295"`
		UInt16Val uint16 `default:"65535"`
		UInt8Val  uint8  `default:"255"`

		StringVal2 string `default:"test"`
		IntVal2    int    `default:"2147483647"`
		StringVal3 string
		ArrayVal   [3]int
	}

	type TestNestedStruct1 struct {
		StringVal string `default:"test"`
		IntVal    int    `default:"1000"`
	}

	type TestNestedStruct2 struct{}

	type TestStruct2 struct {
		TestNestedStruct1
		TestNestedStruct2
		IntVal2 int `default:"100"`
	}

	type TestStruct3 struct {
		*TestNestedStruct1
		*TestNestedStruct2
		IntVal2 int `default:"100"`
		BoolVal bool
	}

	type TestStruct4 struct {
		SliceVal []int `default:"123"`
	}

	type TestStruct5 struct {
		IntPtrVal *int `default:"123"`
	}

	t.Run("struct", func(t *testing.T) {
		c := &TestStruct{
			StringVal2: "text",
			IntVal2:    100,
		}
		expected := &TestStruct{
			StringVal: "test",
			BoolVal:   true,
			IntVal:    -2147483647,
			Int64Val:  -9223372036854775808,
			Int32Val:  -2147483648,
			Int16Val:  -32768,
			Int8Val:   -128,
			UIntVal:   4294967295,
			UInt64Val: 18446744073709551615,
			UInt32Val: 4294967295,
			UInt16Val: 65535,
			UInt8Val:  255,

			StringVal2: "text",
			IntVal2:    100,
			StringVal3: "",
			ArrayVal:   [3]int{0, 0, 0},
		}
		err := ApplyDefaults(c)
		require.NoError(t, err)
		require.Equal(t, expected, c)
	})

	t.Run("nested struct", func(t *testing.T) {
		c := &TestStruct2{
			TestNestedStruct1: TestNestedStruct1{},
			TestNestedStruct2: TestNestedStruct2{},
		}
		expected := &TestStruct2{
			TestNestedStruct1: TestNestedStruct1{
				StringVal: "test",
				IntVal:    1000,
			},
			TestNestedStruct2: TestNestedStruct2{},
			IntVal2:           100,
		}
		err := ApplyDefaults(c)
		require.NoError(t, err)
		require.Equal(t, expected, c)
	})

	t.Run("nested struct with ptr", func(t *testing.T) {
		c := &TestStruct3{
			TestNestedStruct1: &TestNestedStruct1{},
		}
		expected := &TestStruct3{
			TestNestedStruct1: &TestNestedStruct1{
				StringVal: "test",
				IntVal:    1000,
			},
			TestNestedStruct2: nil,
			IntVal2:           100,
			BoolVal:           false,
		}
		err := ApplyDefaults(c)
		require.NoError(t, err)
		require.Equal(t, expected, c)
	})

	t.Run("errors", func(t *testing.T) {
		var err error

		c1 := TestStruct{}
		err = ApplyDefaults(c1)
		require.EqualError(t, err, "pointer to a struct is required")

		c2 := "abc"
		err = ApplyDefaults(&c2)
		require.EqualError(t, err, "pointer to a struct is required")

		c3 := &TestStruct4{}
		err = ApplyDefaults(c3)
		require.EqualError(t, err, "unsupported field type")

		c4 := &TestStruct5{}
		err = ApplyDefaults(c4)
		require.EqualError(t, err, "unsupported field type")
	})

}
