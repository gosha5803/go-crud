package health

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IHealthController interface {
	AssignRoutes()
}

type HealthModule struct {
	controller IHealthController
}

func NewHealthModule(g *gin.Engine, DB *gorm.DB) *HealthModule {
	service := NewHealthService(DB)
	controller := NewHealthController(g, "/health", service)

	return &HealthModule{controller: controller}
}

func (m *HealthModule) Init() {
	m.controller.AssignRoutes()
}