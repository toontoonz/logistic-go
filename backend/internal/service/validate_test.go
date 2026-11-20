package service

import (
	"math"
	"testing"
)

func TestValidateCoord(t *testing.T) {
	if err := ValidateCoord(13.7563, 100.5018); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCoord(-90, -180); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCoord(90, 180); err != nil {
		t.Fatal(err)
	}
	for _, tt := range [][2]float64{{90.1, 0}, {-90.1, 0}, {0, 180.1}, {0, -181}, {math.NaN(), 0}, {0, math.Inf(1)}} {
		if err := ValidateCoord(tt[0], tt[1]); err == nil {
			t.Fatalf("expected error for %v, %v", tt[0], tt[1])
		}
	}
}

func TestRequireTextAndPhone(t *testing.T) {
	if _, err := requireText("name", "   ", 10); err == nil {
		t.Fatal("expected empty name to fail")
	}
	got, err := requireText("name", "  Mali  ", 10)
	if err != nil || got != "Mali" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := validatePhone("081-234"); err == nil {
		t.Fatal("expected short phone to fail")
	}
	if _, err := validatePhone("081-234-5678"); err != nil {
		t.Fatal(err)
	}
	if _, err := validatePhone("081-234-5678x"); err == nil {
		t.Fatal("expected invalid character to fail")
	}
}
