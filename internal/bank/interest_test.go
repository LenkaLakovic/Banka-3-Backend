package bank

import (
	"math"
	"testing"
)

func TestBaseAnnualRate(t *testing.T) {
	tests := []struct {
		amount   float64
		expected float64
	}{
		{0, 6.25},
		{500_000, 6.25},
		{500_001, 6.00},
		{1_000_000, 6.00},
		{1_000_001, 5.75},
		{2_000_000, 5.75},
		{2_000_001, 5.50},
		{5_000_000, 5.50},
		{5_000_001, 5.25},
		{10_000_000, 5.25},
		{10_000_001, 5.00},
		{20_000_000, 5.00},
		{20_000_001, 4.75},
		{100_000_000, 4.75},
	}
	for _, tt := range tests {
		got := BaseAnnualRate(tt.amount)
		if got != tt.expected {
			t.Errorf("BaseAnnualRate(%v) = %v, want %v", tt.amount, got, tt.expected)
		}
	}
}

func TestMarginForLoanType(t *testing.T) {
	tests := []struct {
		lt       loan_type
		expected float64
	}{
		{Cash, 1.75},
		{Mortgage, 1.50},
		{Car, 1.25},
		{Refinancing, 1.00},
		{Student, 0.75},
	}
	for _, tt := range tests {
		got := MarginForLoanType(tt.lt)
		if got != tt.expected {
			t.Errorf("MarginForLoanType(%v) = %v, want %v", tt.lt, got, tt.expected)
		}
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestCalculateAnnuity(t *testing.T) {
	// 1,000,000 RSD at 8% for 12 months => ~86,988.43
	got := CalculateAnnuity(1_000_000, 8.0, 12)
	if !almostEqual(got, 86988.43, 0.01) {
		t.Errorf("CalculateAnnuity(1M, 8%%, 12) = %v, want ~86988.43", got)
	}

	// 10,000 RSD at 8% for 12 months => ~869.88
	got = CalculateAnnuity(10_000, 8.0, 12)
	if !almostEqual(got, 869.88, 0.02) {
		t.Errorf("CalculateAnnuity(10000, 8%%, 12) = %v, want ~869.88", got)
	}
}

func TestCalculateAnnuity_ZeroRate(t *testing.T) {
	got := CalculateAnnuity(12000, 0, 12)
	if got != 1000 {
		t.Errorf("CalculateAnnuity(12000, 0, 12) = %v, want 1000", got)
	}
}

func TestCalculateAnnuity_ZeroMonths(t *testing.T) {
	got := CalculateAnnuity(10000, 5.0, 0)
	if got != 0 {
		t.Errorf("CalculateAnnuity(10000, 5, 0) = %v, want 0", got)
	}
}
