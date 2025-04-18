package utils

import "testing"

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		value     float64
		precision float64
		expected  float64
	}{
		{6.123456, 100, 6.12},      // 2 decimal places
		{-6.123456, 100, -6.12},    // 2 decimal places, negative
		{6.125, 1000, 6.125},       // 3 decimal places
		{-6.125, 100, -6.13},       // 2 decimal places, edge case
		{0.0, 100, 0.0},            // Zero input
		{6.123456, 1, 6.0},         // 0 decimal places
		{6.123456, 0, 6.123456},    // Invalid precision (returns original)
		{6.123456, -100, 6.123456}, // Negative precision (returns original)
	}

	for _, tt := range tests {
		result := RoundFloat(tt.value, tt.precision)
		if result != tt.expected {
			t.Errorf("RoundFloat(%v, %v) = %v; expected %v", tt.value, tt.precision, result, tt.expected)
		}
	}
}
