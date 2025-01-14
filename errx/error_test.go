package errx

import (
	"errors"
	"testing"
)

func TestBaseError(t *testing.T) {
	tests := []struct {
		name     string
		errType  string
		text     string
		nested   error
		expected string
	}{
		{"Basic", "", "Test error", nil, "Test error"},
		{"With type", "Type1", "", nil, "Type1"},
		{"With nested", "", "Outer error", errors.New("inner error"), "Outer error: inner error"},
		{"Nested with same type", "", "Outer error", &BaseError{errType: "Type2", text: "Inner error"}, "Outer error: Type2: Inner error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &BaseError{
				errType: tt.errType,
				text:    tt.text,
				nested:  tt.nested,
			}

			if err.Error() != tt.expected {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.expected)
			}
		})
	}
}

func TestBaseErrorUnwrap(t *testing.T) {
	innerError := New("Inner error")
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{"No nested", New("Test error"), nil},
		{"With nested", Wrap(innerError, "Outer error"), innerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errors.Unwrap(tt.err) != tt.expected {
				t.Errorf("Unwrap() = %v, want %v", errors.Unwrap(tt.err), tt.expected)
			}
		})
	}
}

func TestBaseErrorIs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  *BaseError
		expected bool
	}{
		{"Basic type", NewWithType(NewType("Type1"), "Test error"), NewType("Type1"), true},
		{"Wrong type", NewWithType(NewType("Type1"), "Test error"), NewType("Type2"), false},
		{"Outer type match", WrapWithType(NewType("Type1"), NewWithType(NewType("Type2"), "Inner"), "Outer"), NewType("Type1"), true},
		{"Inner type not match", WrapWithType(NewType("Type1"), NewWithType(NewType("Type2"), "Inner"), "Outer"), NewType("Type2"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For Is method, we only care about the top-level error
			if tt.err.(*BaseError).Is(tt.errType) != tt.expected {
				t.Errorf("Error: %v, Is() = %v, want %v", tt.err.(*BaseError), !tt.expected, tt.expected)
			}

		})
	}
}

func TestBaseErrorIsRecursiveBuiltIn(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  *BaseError
		expected bool
	}{
		{"Basic type", NewWithType(NewType("Type1"), "Test error"), NewType("Type1"), true},
		{"Wrong type", NewWithType(NewType("Type1"), "Test error"), NewType("Type2"), false},
		{"Outer type match", WrapWithType(NewType("Type1"), NewWithType(NewType("Type2"), "Inner"), "Outer"), NewType("Type1"), true},
		{"Inner type match", WrapWithType(NewType("Type1"), NewWithType(NewType("Type2"), "Inner"), "Outer"), NewType("Type2"), true},
		{"No type match", WrapWithType(NewType("Type1"), NewWithType(NewType("Type2"), "Inner"), "Outer"), NewType("Type3"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For Is method, we only care about the top-level error
			if errors.Is(tt.err, tt.errType) != tt.expected {
				t.Errorf("Error: %v, Is() = %v, want %v", tt.err.(*BaseError), !tt.expected, tt.expected)
			}

		})
	}
}

func TestNewType(t *testing.T) {
	err := NewType("TestType")
	if err.errType != "TestType" {
		t.Errorf("NewType() = %v, want %v", err, "TestType")
	}
	if err.text != "" || err.nested != nil {
		t.Errorf("NewType() side effects! BaseError: %v", err)
	}
}

func TestNew(t *testing.T) {
	err := New("Test error").(*BaseError)
	if err.text != "Test error" {
		t.Errorf("New() = %v, want %v", err, "Test error")
	}
	if err.errType != "" || err.nested != nil {
		t.Errorf("New() side effects! BaseError: %v", err)
	}
}

func TestNewf(t *testing.T) {
	err := Newf("Test %s", "error").(*BaseError)
	if err.text != "Test error" {
		t.Errorf("Newf() = %v, want %v", err, "Test error")
	}
	if err.errType != "" || err.nested != nil {
		t.Errorf("Newf() side effects! BaseError: %v", err)
	}
}

func TestNewWithType(t *testing.T) {
	err := NewWithType(NewType("TestType"), "Test error").(*BaseError)
	if err.errType != "TestType" || err.text != "Test error" {
		t.Errorf("NewWithType() = %v, want %v", err, "TestType: Test error")
	}
	if err.Error() != "TestType: Test error" {
		t.Errorf("NewWithType().Error() = %v, want %v", err.Error(), "TestType: Test error")
	}
	if err.errType != "TestType" || err.nested != nil {
		t.Errorf("Newf() side effects! BaseError: %v", err)
	}
}

func TestNewWithTypef(t *testing.T) {
	err := NewWithTypef(NewType("TestType"), "Test %s", "error").(*BaseError)
	if err.errType != "TestType" || err.text != "Test error" {
		t.Errorf("NewWithTypef() = %v, want %v", err, "TestType: Test error")
	}
	if err.Error() != "TestType: Test error" {
		t.Errorf("NewWithType().Error() = %v, want %v", err.Error(), "TestType: Test error")
	}
	if err.errType != "TestType" || err.nested != nil {
		t.Errorf("Newf() side effects! BaseError: %v", err)
	}
}

func TestWrap(t *testing.T) {
	inner := New("Inner")
	err := Wrap(inner, "Outer")
	if err.Error() != "Outer: Inner" {
		t.Errorf("Wrap() = %v, want %v", err, "Outer: Inner")
	}
	if !errors.Is(errors.Unwrap(err), inner) {
		t.Errorf("Wrap() nested error = %s, want %s", errors.Unwrap(err).Error(), "Inner")
	}
}

func TestWrapf(t *testing.T) {
	inner := New("Inner")
	err := Wrapf(inner, "Outer %s", "error")
	if err.Error() != "Outer error: Inner" {
		t.Errorf("Wrapf() = %v, want %v", err, "Outer error: Inner")
	}
	if !errors.Is(errors.Unwrap(err), inner) {
		t.Errorf("Wrap() nested error = %s, want %s", errors.Unwrap(err).Error(), "Inner")
	}
}

func TestWrapWithType(t *testing.T) {
	err := WrapWithType(NewType("TestType"), errors.New("Original"), "Wrapped")
	if err.(*BaseError).errType != "TestType" || err.(*BaseError).text != "Wrapped" {
		t.Errorf("WrapWithType() = %v, want %v", err, "TestType:Wrapped")
	}
}

func TestWrapWithTypef(t *testing.T) {
	err := WrapWithTypef(NewType("TestType"), errors.New("Original"), "Wrapped %s", "error")
	if err.(*BaseError).errType != "TestType" || err.(*BaseError).text != "Wrapped error" {
		t.Errorf("WrapWithTypef() = %v, want %v", err, "TestType:Wrapped error")
	}
}
