package service

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

func requireText(field, value string, max int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", &InputError{Message: field + " is required"}
	}
	if utf8.RuneCountInString(value) > max {
		return "", &InputError{Message: field + " is too long"}
	}
	return value, nil
}

func validatePhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 8 || len(value) > 20 {
		return "", &InputError{Message: "phone must be 8–20 characters"}
	}
	for _, r := range value {
		switch {
		case unicode.IsDigit(r), r == '+', r == '-', r == ' ', r == '(', r == ')':
		default:
			return "", &InputError{Message: "phone has an invalid character"}
		}
	}
	return value, nil
}

// ValidateCoord rejects values that cannot be a WGS84 position.
// NaN must be checked explicitly: every comparison with NaN is false,
// so `lat > 90` alone would let NaN through.
func ValidateCoord(lat, lng float64) error {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return &InputError{Message: "latitude must be between -90 and 90"}
	}
	if math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		return &InputError{Message: "longitude must be between -180 and 180"}
	}
	return nil
}
