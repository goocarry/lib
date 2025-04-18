package utils

import "math"

// RoundFloat rounds a floating-point number to a specified precision using a multiplier.
//
// Parameters:
//   - value: The floating-point number to round.
//   - precision: The multiplier representing the precision (e.g., 100 for 2 decimal places, 1000 for 3).
//     Must be positive and non-zero.
//
// Returns:
//   - The rounded floating-point number.
//   - If precision is invalid (non-positive), returns the original value.
//
// Notes:
//   - The function correctly handles both positive and negative numbers.
//   - For financial or high-precision calculations, consider using a library like "github.com/shopspring/decimal"
//     to avoid potential floating-point precision issues inherent to float64.
//
// Examples:
//
//	RoundFloat(6.123456, 100)    // Returns 6.12
//	RoundFloat(-6.123456, 100)   // Returns -6.12
//	RoundFloat(6.125, 1000)      // Returns 6.125
func RoundFloat(value float64, precision float64) float64 {
	if precision <= 0 {
		return value
	}
	return math.Round(value*precision) / precision
}
