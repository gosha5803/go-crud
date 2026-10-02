package user

// 10:50
import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/initializers"
)

type IUserService interface {
	Auth(AuthReqDto) (User, error)
}

type UserController struct {
	g       *gin.Engine
	path    string
	servcie IUserService
}

func NewUserController(g *gin.Engine, path string, servcie IUserService) *UserController {
	return &UserController{g: g, path: path, servcie: servcie}
}

func (c *UserController) AssignRoutes() {
	preffix := c.g.Group(c.path)

	preffix.POST("/login", c.login)
}

// start 11 45

func (c *UserController) login(ctx *gin.Context) {
	var req AuthReqDto

	if err := ctx.ShouldBindJSON(&req); err != nil {
		respondValidationError(ctx, "userController: login", err)
		return
	}

	result, err := c.servcie.Auth(req)

	if err != nil {
		// TODO тут наверное свой хендлер ошибок
		respondServiceError(ctx, "user controller", err)
		return
	}

	resp := AuthResDto{UserID: result.ID, Message: "На вашу почту было отправлено письмо для подтверждения"}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// TODO вынос?
func respondValidationError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	ctx.JSON(http.StatusBadRequest, gin.H{
		"errors": initializers.FormatValidationErrors(err),
	})
}

func respondServiceError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	if errors.Is(err, ErrEmailAlreadyUsed) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Пользователь с данным email уже существует",
		})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "Непредвиденная ошибка, попробуйте позже",
	})
}

// start 13 33
