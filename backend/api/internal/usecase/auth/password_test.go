package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/mq/api/internal/domain/apperrors"
)

func TestValidatePassword_bounds(t *testing.T) {
	if err := validatePassword(strings.Repeat("a", maxPasswordBytes)); err != nil {
		t.Fatalf("72 bytes should be fine: %v", err)
	}
	if err := validatePassword(strings.Repeat("a", maxPasswordBytes+1)); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("73 bytes should be rejected, got %v", err)
	}
	if err := validatePassword("short"); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("too short should be rejected, got %v", err)
	}
}
