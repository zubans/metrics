package grpc_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrconvParsing(t *testing.T) {
	t.Run("Parse gauge values with strconv", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected float64
			valid    bool
		}{
			{"123.45", 123.45, true},
			{"0.0", 0.0, true},
			{"-123.456", -123.456, true},
			{"1e10", 1e10, true},
			{"invalid", 0, false},
			{"", 0, false},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				parsed, err := strconv.ParseFloat(tc.input, 64)
				if tc.valid {
					require.NoError(t, err)
					assert.Equal(t, tc.expected, parsed)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("Parse counter values with strconv", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected int64
			valid    bool
		}{
			{"123", 123, true},
			{"0", 0, true},
			{"-456", -456, true},
			{"9223372036854775807", 9223372036854775807, true}, // max int64
			{"invalid", 0, false},
			{"123.45", 0, false}, // float не должен парситься как int
			{"", 0, false},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				parsed, err := strconv.ParseInt(tc.input, 10, 64)
				if tc.valid {
					require.NoError(t, err)
					assert.Equal(t, tc.expected, parsed)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("Format values with strconv", func(t *testing.T) {
		// Тестируем форматирование значений
		assert.Equal(t, "123.45", strconv.FormatFloat(123.45, 'f', -1, 64))
		assert.Equal(t, "0", strconv.FormatFloat(0.0, 'f', -1, 64))
		assert.Equal(t, "-123.456", strconv.FormatFloat(-123.456, 'f', -1, 64))

		assert.Equal(t, "123", strconv.FormatInt(123, 10))
		assert.Equal(t, "0", strconv.FormatInt(0, 10))
		assert.Equal(t, "-456", strconv.FormatInt(-456, 10))
	})
}
