package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type Users struct {
	Admin     *users.AdminService
	JWTSecret []byte
}

type createUserBody struct {
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required"`
	Roles    []string `json:"roles" binding:"required,min=1"`
}

func rolesIncludeAdmin(roleNames []string) bool {
	for _, roleName := range roleNames {
		if strings.TrimSpace(strings.ToLower(roleName)) == domain.RolAdministrador {
			return true
		}
	}
	return false
}

func tokenRolesIncludeAdmin(tokenRoles []string) bool {
	for _, roleName := range tokenRoles {
		if roleName == domain.RolAdministrador {
			return true
		}
	}
	return false
}

// Post creates a user. Empty DB: no token, body must include administrador. Otherwise: admin JWT required.
func (handler *Users) Post(c *gin.Context) {
	var payload createUserBody
	if bindErr := c.ShouldBindJSON(&payload); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: name, email, password, and roles required"})
		return
	}

	ctx := c.Request.Context()
	userCount, countErr := handler.Admin.CountUsers(ctx)
	if countErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": countErr.Error()})
		return
	}

	if !handler.authorizeUserCreation(c, userCount, payload.Roles) {
		return
	}

	handler.respondCreateUser(c, ctx, payload)
}

func (handler *Users) authorizeUserCreation(c *gin.Context, userCount int64, roles []string) bool {
	if userCount == 0 {
		return handler.authorizeBootstrapFirstUser(c, roles)
	}
	return handler.authorizeAdminBearer(c)
}

func (handler *Users) authorizeBootstrapFirstUser(c *gin.Context, roles []string) bool {
	if rolesIncludeAdmin(roles) {
		return true
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "first user must include role administrador (no token until a user exists)",
	})
	return false
}

func (handler *Users) authorizeAdminBearer(c *gin.Context) bool {
	if len(handler.JWTSecret) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured: JWT_SECRET missing"})
		return false
	}
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization: Bearer <token> required"})
		return false
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	_, tokenRoles, parseErr := auth.ParseToken(tokenString, handler.JWTSecret)
	if parseErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return false
	}
	if !tokenRolesIncludeAdmin(tokenRoles) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin role required to create users"})
		return false
	}
	return true
}

func (handler *Users) respondCreateUser(c *gin.Context, ctx context.Context, payload createUserBody) {
	createdUser, createErr := handler.Admin.Create(ctx, payload.Name, payload.Email, payload.Password, payload.Roles)
	if createErr != nil {
		switch {
		case errors.Is(createErr, users.ErrEmailAlreadyRegistered):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		case errors.Is(createErr, users.ErrNoRoles):
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one role is required"})
		case errors.Is(createErr, users.ErrPasswordTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		case errors.Is(createErr, users.ErrUnknownRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown or invalid role"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": createErr.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, createdUser)
}

func (handler *Users) GetList(c *gin.Context) {
	role := c.Query("role")
	var (
		userList []domain.User
		listErr  error
	)
	if role != "" {
		userList, listErr = handler.Admin.ListByRole(c.Request.Context(), role)
	} else {
		userList, listErr = handler.Admin.List(c.Request.Context())
	}
	if listErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": listErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": userList})
}
