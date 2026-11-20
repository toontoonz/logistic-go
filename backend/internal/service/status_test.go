package service

import "testing"

func TestCanTransition(t *testing.T) {
	// Table-driven tests are the usual Go style: one function, many cases.
	tests := []struct {
		from, to string
		ok       bool
	}{
		{StatusAssigned, StatusPickedUp, true},
		{StatusPickedUp, StatusInTransit, true},
		{StatusInTransit, StatusDelivered, true},
		{StatusPending, StatusAssigned, false}, // assignment is its own action
		{StatusPending, StatusDelivered, false},
		{StatusAssigned, StatusDelivered, false},
		{StatusDelivered, StatusInTransit, false},
		{StatusPickedUp, StatusPickedUp, false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if got := CanTransition(tt.from, tt.to); got != tt.ok {
				t.Fatalf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.ok)
			}
		})
	}
}
