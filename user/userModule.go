package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IUserController interface {
	AssignRoutes()
}

type UserModule struct {
	userController IUserController
}

func NewUserModule(g *gin.Engine, DB *gorm.DB) *UserModule {
	userService := NewUserService(DB)
	userController := NewUserController(g, "/users", userService)

	return &UserModule{
		userController: userController,
	}
}

func (m *UserModule) Init() {
	m.userController.AssignRoutes()
}
