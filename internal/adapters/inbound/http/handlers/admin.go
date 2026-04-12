package handlers

import (
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/gin-gonic/gin"
)

type Admin struct {
	Platform *admin.PlatformOverviewService
}

func NewAdminHandler(platform *admin.PlatformOverviewService) *Admin {
	return &Admin{Platform: platform}
}

func (handler *Admin) GetOverview(c *gin.Context) {
	overview, err := handler.Platform.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}
