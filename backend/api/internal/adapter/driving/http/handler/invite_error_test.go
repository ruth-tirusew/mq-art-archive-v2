package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/domain/apperrors"
)

func inviteErrorResponse(err error) (int, string) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/v1/invites", nil)
	writeInviteError(c, err)
	return w.Code, w.Body.String()
}

func TestWriteInviteErrorTellsTheAdminWhyItConflicts(t *testing.T) {
	code, body := inviteErrorResponse(fmt.Errorf("%w: %s already has an Artiv account", apperrors.ErrConflict, "ada@example.com"))
	if code != http.StatusConflict || !strings.Contains(body, "ada@example.com already has an Artiv account") {
		t.Fatalf("got %d %s", code, body)
	}
	if strings.Contains(body, "conflict:") {
		t.Errorf("the sentinel prefix should not leak into the message: %s", body)
	}
}

func TestWriteInviteErrorKeepsRepositoryErrorsPrivate(t *testing.T) {
	code, body := inviteErrorResponse(fmt.Errorf("insert invite: pq: duplicate key idx_x: %w", apperrors.ErrConflict))
	if code != http.StatusConflict || strings.Contains(body, "pq") || strings.Contains(body, "idx_x") {
		t.Fatalf("internals must not reach the client: %d %s", code, body)
	}
	code, _ = inviteErrorResponse(errors.New("boom"))
	if code != http.StatusInternalServerError {
		t.Errorf("unknown errors stay a 500, got %d", code)
	}
}
