package posts

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
)

// Go-сообщество склоняется к тому, что интерфейс должен объявляться там, где он потребляется, а не там, где реализуется.
type IPostService interface {
	CreatePost(CreatePostDto) (models.Post, error)
	GetPosts() ([]models.Post, error)
	GetPostById(uint) (models.Post, error)
	UpdatePost(uint, UpdatePostDto) (models.Post, error)
	DeletePost(uint) (bool, error)
}

// 11:33

type PostController struct {
	g *gin.Engine
	// Вот тут контроллер знает о типе сервиса
	// По сути тип должны быть на уровне портов,
	// и неявно реализовываться адаптерами
	postService IPostService
	path        string
}

func NewPostController(g *gin.Engine, path string, service IPostService) *PostController {
	return &PostController{g: g, postService: service, path: path}
}

func (controller *PostController) AssignRoutes() {
	post := controller.g.Group(controller.path)
	{
		post.GET("", controller.getPosts)
		post.POST("", controller.createPost)
		post.GET("/:id", controller.GetPostByID)
		post.PATCH("/:id", controller.updatePost)
		post.DELETE("/:id", controller.deletePost)
	}
}

func (controller *PostController) getPosts(c *gin.Context) {
	posts, err := controller.postService.GetPosts()

	if err != nil {
		respondServiceError(c, "getPosts", err)
		return
	}

	c.JSON(200, gin.H{
		"posts": posts,
	})
}

func (controller *PostController) GetPostByID(c *gin.Context) {
	var req PostIdPathParam

	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, "getPostById", err)
		return
	}

	post, err := controller.postService.GetPostById(req.ID)

	if err != nil {
		respondServiceError(c, "getPostById", err)
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) createPost(c *gin.Context) {
	var body CreatePostDto

	// c.Bind` сам пишет 400 и игнорирует ошибку в твоём коде
	// c.ShouldBindJSON - тут отлавливается связь с required gin
	// Понять как он работает и все дела.
	if err := c.ShouldBindJSON(&body); err != nil {
		respondValidationError(c, "createPost", err)
		return
	}

	post, err := controller.postService.CreatePost(body)

	if err != nil {
		respondServiceError(c, "createPost", err)
		return
	}

	c.JSON(201, gin.H{
		"post": post,
	})
}

func (controller *PostController) updatePost(c *gin.Context) {
	// Get id from URL
	var req PostIdPathParam
	// ShouldBindUri делает и биндинг, и конвертацию в int, и валидацию.
	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, "updatePost", err)
		return
	}

	var body UpdatePostDto

	// Get body from req
	if err := c.ShouldBindJSON(&body); err != nil {
		respondValidationError(c, "updatePost", err)
		return
	}

	post, err := controller.postService.UpdatePost(req.ID, body)

	if err != nil {
		respondServiceError(c, "updatePost", err)
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) deletePost(c *gin.Context) {
	// Get id from URL
	var req PostIdPathParam

	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, "deletePost", err)
		return
	}

	success, err := controller.postService.DeletePost(req.ID)

	if err != nil {
		respondServiceError(c, "deletePost", err)
		return
	}

	c.JSON(200, gin.H{
		"success": success,
	})
}

func respondInvalidID(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	ctx.JSON(http.StatusBadRequest, gin.H{"errors": []initializers.ClientError{{
		Message: "id поста должен быть положительным числом",
	}}})
}

func respondValidationError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	ctx.JSON(http.StatusBadRequest, gin.H{
		"errors": initializers.FormatValidationErrors(err),
	})
}

func respondServiceError(ctx *gin.Context, op string, err error) {
	log.Printf("controller: %s, %v", op, err)

	// Использование sentinel
	if errors.Is(err, ErrPostNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Пост не найден",
		})

		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "Непредвиденная ошибка, попробуйте позже",
	})
}
