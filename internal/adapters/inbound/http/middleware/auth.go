package middleware

import (
	"net/http"
	"strings"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/gin-gonic/gin"
)

const ctxUserID = "authUserID"
const ctxRoles = "authRoles"

// Autenticar valida Authorization: Bearer <jwt>.
func Autenticar(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(jwtSecret) == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "servidor sin JWT_SECRET"})
			return
		}
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "falta token"})
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		uid, roles, err := auth.ParseToken(raw, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		c.Set(ctxUserID, uid)
		c.Set(ctxRoles, roles)
		c.Next()
	}
}

// ExigeRol exige un rol global (ej. domain.RolAdministrador) después de Autenticar.
func ExigeRol(nombre string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ctxRoles)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "sin permiso"})
			return
		}
		roles, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "sin permiso"})
			return
		}
		for _, r := range roles {
			if r == nombre {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "sin permiso para esta acción"})
	}
}
