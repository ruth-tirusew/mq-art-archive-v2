package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
)

type UserAdminHandler struct{ users inbound.AuthService }

func NewUserAdminHandler(users inbound.AuthService) *UserAdminHandler {
	return &UserAdminHandler{users: users}
}

func (h *UserAdminHandler) List(c *gin.Context) {
	limit := queryLimit(c, 50)
	offset := queryOffset(c)
	var role *identity.Role
	if raw := c.Query("role"); raw != "" {
		value := identity.Role(raw)
		role = &value
	}
	users, total, err := h.users.ListUsers(c.Request.Context(), role, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	data := make([]any, 0, len(users))
	for i := range users {
		data = append(data, toUserResponse(&users[i]))
	}
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, data)
}

func (h *UserAdminHandler) UpdateRole(c *gin.Context) {
	actorID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Role identity.Role `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, err := h.users.UpdateUserRole(c.Request.Context(), actorID, userID, req.Role)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}
