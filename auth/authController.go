package auth

// 10:50
import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/user"
)

type IAuthService interface {
	Auth(AuthReqDto) (user.User, error)
	ActivateUser(string) error
}

type AuthController struct {
	g       *gin.Engine
	path    string
	servcie IAuthService
}

func NewAuthController(g *gin.Engine, path string, servcie IAuthService) *AuthController {
	return &AuthController{g: g, path: path, servcie: servcie}
}

func (c *AuthController) AssignRoutes() {
	preffix := c.g.Group(c.path)

	preffix.POST("/register", c.register)
	preffix.GET("/activate", c.activateUser)
}

// start 11 45

func (c *AuthController) activateUser(ctx *gin.Context) {
	var req ActivateUserDto

	if err := ctx.ShouldBindQuery(&req); err != nil {
		respondValidationError(ctx, "userController: activateUser", err)
		return
	}

	if err := c.servcie.ActivateUser(req.Token); err != nil {
		respondServiceError(ctx, "userController: activateUser", err)
		return
	}

	response := UserActivatedDto{
		Message: "Поздравляем! Ваш аккаунт успешно подтверждён.",
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": response,
	})

}

func (c *AuthController) register(ctx *gin.Context) {
	var req AuthReqDto

	if err := ctx.ShouldBindJSON(&req); err != nil {
		respondValidationError(ctx, "userController: login", err)
		return
	}

	_, err := c.servcie.Auth(req)

	if err != nil {
		respondServiceError(ctx, "user controller", err)
		return
	}

	resp := AuthResDto{
		Message: "На вашу почту было отправена ссылка для подтверждения. Подтвердите вашу почту в течение 24 часов.",
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// TODO дубль от контроллера постов
func respondValidationError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	ctx.JSON(http.StatusBadRequest, gin.H{
		"errors": initializers.FormatValidationErrors(err),
	})
}

func respondServiceError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	var httpErr *HTTPErr

	if errors.As(err, &httpErr) {
		ctx.JSON(httpErr.Status, gin.H{
			"error": httpErr.Message,
		})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "Непредвиденная ошибка, попробуйте позже",
	})
}
