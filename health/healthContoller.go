package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IHealthService interface {
	GetDBStats() (DBHealthResDTO, error)
}

type HealthController struct {
	g *gin.Engine
	pathPrefix string
	service IHealthService
}

func NewHealthController(g *gin.Engine, pathPrefix string, service IHealthService) *HealthController {
	return &HealthController{g: g, service: service, pathPrefix: pathPrefix}
} 

func (c *HealthController) AssignRoutes() {
	group := c.g.Group(c.pathPrefix)

	group.GET("/health", c.health)
}

func (c *HealthController) health(g *gin.Context) {
	dbStats, err := c.service.GetDBStats()

	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": "Непредвиденная ошибка, попробуйте позже",
		})
	}

	g.JSON(http.StatusOK, gin.H{
		"stats": dbStats,
	})
}