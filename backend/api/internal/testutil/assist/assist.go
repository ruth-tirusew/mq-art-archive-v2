package assist

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func ErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("errors.Is: got %v, want %v", err, target)
	}
}

func Equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func NotEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want == got {
		t.Fatalf("got %v, should not equal %v", got, want)
	}
}

func Contains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("got %q, want substring %q", s, substr)
	}
}

func GreaterOrEqual(t *testing.T, got, want int) {
	t.Helper()
	if got < want {
		t.Fatalf("got %d, want >= %d", got, want)
	}
}

func Len(t *testing.T, want, got int) {
	t.Helper()
	if got != want {
		t.Fatalf("len: got %d, want %d", got, want)
	}
}

func NotNil[T any](t *testing.T, v *T) {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
}

func WithinDuration(t *testing.T, want, got time.Time, delta time.Duration) {
	t.Helper()
	diff := want.Sub(got)
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		t.Fatalf("time %v not within %v of %v", got, delta, want)
	}
}
