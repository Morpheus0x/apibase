package errx_test

import (
	"errors"
	"fmt"
	"testing"

	"gopkg.cc/apibase/errx"
)

func TestNewFunctions(t *testing.T) {
	typeT1 := errx.NewType("t1")
	typeT2 := errx.NewType("t2")

	tests := []struct {
		name   string
		fn     func() error
		want   string
		isType error // nil if no type comparison is expected
	}{
		{
			name: "New",
			fn:   func() error { return errx.New("simple") },
			want: "simple",
		},
		{
			name: "Newf",
			fn:   func() error { return errx.Newf("f%d", 1) },
			want: "f1",
		},
		{
			name:   "NewWithType",
			fn:     func() error { return errx.NewWithType(typeT1, "msg") },
			want:   "t1: msg",
			isType: typeT1,
		},
		{
			name: "Wrap",
			fn:   func() error { return errx.Wrap(fmt.Errorf("orig"), "wrap") },
			want: "wrap: orig",
		},
		{
			name:   "WrapWithType",
			fn:     func() error { return errx.WrapWithType(typeT2, fmt.Errorf("orig"), "wrap") },
			want:   "t2: wrap: orig",
			isType: typeT2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if got := err.Error(); got != tt.want {
				t.Fatalf("Error() = %q; want %q", got, tt.want)
			}
			if tt.isType != nil && !errors.Is(err, tt.isType) {
				t.Errorf("errors.Is(%v, %v) = false; want true", err, tt.isType)
			}
		})
	}
}

func TestTypeComparisonFails(t *testing.T) {
	typeT4 := errx.NewType("t4")
	err := errx.NewWithType(typeT4, "msg")

	if errors.Is(err, errx.NewType("t4")) { // different pointer → should be false
		t.Errorf("errors.Is returned true for a different type instance")
	}
}

func TestErrorsIsWithNested(t *testing.T) {
	typeT3 := errx.NewType("t3")
	base := errx.NewWithType(typeT3, "base")
	nested := errx.Wrap(base, "wrap")

	if !errors.Is(nested, typeT3) {
		t.Errorf("expected nested to match type t3")
	}
}

func TestBaseError(t *testing.T) {
	// Test New function
	err := errx.New("test error")
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}

	// Test Newf function
	err = errx.Newf("test error %d", 123)
	if err.Error() != "test error 123" {
		t.Errorf("Expected 'test error 123', got '%s'", err.Error())
	}

	// Test NewWithType function
	typeErr := errx.NewType("validation")
	err = errx.NewWithType(typeErr, "field required")
	expected := "validation: field required"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test NewWithTypef function
	err = errx.NewWithTypef(typeErr, "field %s required", "username")
	expected = "validation: field username required"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test Wrap function
	baseErr := errors.New("base error")
	err = errx.Wrap(baseErr, "wrapped error")
	expected = "wrapped error: base error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test Wrapf function
	err = errx.Wrapf(baseErr, "wrapped error %d", 456)
	expected = "wrapped error 456: base error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test WrapWithType function
	err = errx.WrapWithType(typeErr, baseErr, "validation failed")
	expected = "validation: validation failed: base error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test WrapWithTypef function
	err = errx.WrapWithTypef(typeErr, baseErr, "validation failed for %s", "email")
	expected = "validation: validation failed for email: base error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test errors.Is comparison
	err = errx.NewWithType(typeErr, "test error")
	if !errors.Is(err, typeErr) {
		t.Errorf("errors.Is should return true for matching error types")
	}

	// Test nested error wrapping
	innerErr := errx.New("inner error")
	outerErr := errx.Wrap(innerErr, "outer error")
	if !errors.Is(outerErr, innerErr) {
		t.Errorf("errors.Is should find nested errors")
	}
}

func TestBaseErrorRecursion(t *testing.T) {
	typeErr1 := errx.NewType("type1")
	typeErr2 := errx.NewType("type2")

	// Test recursive error wrapping
	err1 := errx.NewWithType(typeErr1, "error 1")
	err2 := errx.WrapWithType(typeErr2, err1, "error 2")
	err3 := errx.Wrap(err2, "error 3")

	expected := "error 3: type2: error 2: type1: error 1"
	if err3.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err3.Error())
	}

	// Test errors.Is with recursive structure
	if !errors.Is(err3, typeErr1) {
		t.Errorf("errors.Is should find nested BaseErrorType")
	}
	if !errors.Is(err3, typeErr2) {
		t.Errorf("errors.Is should find nested BaseErrorType")
	}
}

func TestBaseErrorEmptyFields(t *testing.T) {
	// Test with empty text
	err := errx.NewWithType(errx.NewType("error"), "")
	expected := "error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test with nil nested error
	err = errx.Wrap(nil, "test")
	expected = "test"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test with nil error type
	err = errx.NewWithType(nil, "test")
	expected = "test"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}
