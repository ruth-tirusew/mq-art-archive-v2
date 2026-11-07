package onboarding

import (
	"testing"

	"github.com/mq/api/internal/testutil/assist"
)

func TestParseStatusFilter(t *testing.T) {
	for raw, want := range map[string]ApprovalStatus{
		"": ApprovalStatusPending, "pending": ApprovalStatusPending,
		"approved": ApprovalStatusApproved, "rejected": ApprovalStatusRejected,
	} {
		got, ok := ParseStatusFilter(raw)
		if !ok {
			t.Fatalf("%q should be accepted", raw)
		}
		assist.NotNil(t, got)
		assist.Equal(t, want, *got)
	}
	all, ok := ParseStatusFilter("all")
	if !ok || all != nil {
		t.Error("all should mean no status filter")
	}
	if _, ok := ParseStatusFilter("bogus"); ok {
		t.Error("an unknown status must be rejected")
	}
}
