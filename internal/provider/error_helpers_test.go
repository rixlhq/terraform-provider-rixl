package provider

import (
	"errors"
	"fmt"
	"testing"
)

type mockStatusError struct {
	StatusCode int64
	Msg        string
}

func (e *mockStatusError) Error() string {
	return e.Msg
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "substring HTTP 404",
			err:      errors.New("request failed with HTTP 404 Not Found"),
			expected: true,
		},
		{
			name:     "substring status 404",
			err:      errors.New("response status 404: entity not found"),
			expected: true,
		},
		{
			name:     "status code struct field",
			err:      &mockStatusError{StatusCode: 404, Msg: "not found"},
			expected: true,
		},
		{
			name:     "wrapped status code error",
			err:      fmt.Errorf("outer error: %w", &mockStatusError{StatusCode: 404, Msg: "nested"}),
			expected: true,
		},
		{
			name:     "different status code",
			err:      &mockStatusError{StatusCode: 500, Msg: "internal error"},
			expected: false,
		},
		{
			name:     "generic non-404 error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isNotFound(tc.err)
			if got != tc.expected {
				t.Errorf("isNotFound(%v) = %v; want %v", tc.err, got, tc.expected)
			}
		})
	}
}
