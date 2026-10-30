package onboarding

import (
	"testing"
	"time"
)

func TestInviteStatusAndClaimable(t *testing.T) {
	now := time.Now()
	live := Invite{Email: "Ada@Example.com", ExpiresAt: now.Add(time.Hour)}
	if live.Status(now) != InviteLive || !live.Claimable(now, " ada@example.com ") {
		t.Error("live invite should be claimable by case-insensitive matching email")
	}
	if live.Claimable(now, "other@example.com") {
		t.Error("a different email must not be able to claim")
	}

	expired := Invite{ExpiresAt: now.Add(-time.Second)}
	if expired.Status(now) != InviteExpired {
		t.Error("past expiry should be expired")
	}

	used := now
	claimed := Invite{ExpiresAt: now.Add(-time.Hour), ClaimedAt: &used}
	if claimed.Status(now) != InviteClaimed {
		t.Error("claimed should win over expired")
	}
	revoked := Invite{ExpiresAt: now.Add(time.Hour), RevokedAt: &used}
	if revoked.Status(now) != InviteRevoked || revoked.Claimable(now, "") {
		t.Error("revoked invite must not be claimable")
	}
}

func TestHashInviteToken(t *testing.T) {
	if HashInviteToken("abc") != HashInviteToken("abc") || HashInviteToken("abc") == HashInviteToken("abd") {
		t.Error("hash must be deterministic and input-sensitive")
	}
	if len(HashInviteToken("abc")) != 64 {
		t.Error("expected hex sha-256")
	}
}
