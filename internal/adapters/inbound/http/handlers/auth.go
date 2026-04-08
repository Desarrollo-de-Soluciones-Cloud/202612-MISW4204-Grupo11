package handlers

import (
	"errors"
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	Login *auth.LoginService
}

type loginBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (handler *Auth) PostLogin(c *gin.Context) {
	var payload loginBody
	if bindErr := c.ShouldBindJSON(&payload); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: email and password required"})
		return
	}
	result, loginErr := handler.Login.Login(c.Request.Context(), payload.Email, payload.Password)
	if loginErr != nil {
		if errors.Is(loginErr, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": loginErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}
