package posts

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/auth"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
)

// Go-сообщество склоняется к тому, что интерфейс должен объявляться там, где он потребляется, а не там, где реализуется.
type IPostService interface {
	CreatePost(context.Context, string, string, uint) (models.Post, error)
	GetPosts(context.Context) ([]models.Post, error)
	GetPostById(context.Context,uint) (models.Post, error)
	UpdatePost(context.Context, UpdatePostParams) (models.Post, error)
	DeletePost(context.Context, DeletePostByIDParams) (error)
	IncrementViews( context.Context,  uint)(models.Post, error)
}

// 11:33

type PostController struct {
	g *gin.Engine
	// Вот тут контроллер знает о типе сервиса
	// По сути тип должны быть на уровне портов,
	// и неявно реализовываться адаптерами
	postService IPostService
	path        string
	authMW gin.HandlerFunc
}

func NewPostController(
	g *gin.Engine, 
	path string, 
	service IPostService,
	authMW gin.HandlerFunc,
	) *PostController {
	return &PostController{
		g: g, 
		postService: service, 
		path: path,
		authMW: authMW,
	}
}

func (controller *PostController) AssignRoutes() {
	prefix := controller.g.Group(controller.path)
	prefix.Use(controller.authMW)
	{
		prefix.GET("", controller.getPosts)
		prefix.POST("", controller.createPost)
		prefix.GET("/:id", controller.GetPostByID)
		prefix.PATCH("/:id", controller.updatePost)
		prefix.DELETE("/:id", controller.deletePost)
		prefix.POST("/:id/views", controller.incrementViews)
	}
}

func requireUserID(ctx *gin.Context) (uint, bool) {
	userId, extractErr := auth.GetUserIdFromContext(ctx)

	if extractErr != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return 0, false
	}

	return userId, true
}

func (controller *PostController) incrementViews(c *gin.Context) {
	var req PostIdPathParam
	op := "increment views"

	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, op, err)
		return
	}

	ctx := c.Request.Context()
	post, err := controller.postService.IncrementViews(ctx, req.ID)
	
	if err != nil {
		respondServiceError(c, op, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post": post,
	})

}

func (controller *PostController) getPosts(c *gin.Context) {
	ctx := c.Request.Context()
	posts, err := controller.postService.GetPosts(ctx)
	op := "getPosts"

	if err != nil {
		respondServiceError(c, op, err)
		return
	}

	c.JSON(200, gin.H{
		"posts": posts,
	})
}

func (controller *PostController) GetPostByID(c *gin.Context) {
	var req PostIdPathParam
	op := "getPostById"
	
	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, op, err)
		return
	}
	
	ctx := c.Request.Context()
	post, err := controller.postService.GetPostById(ctx, req.ID)

	if err != nil {
		respondServiceError(c, op, err)
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) createPost(c *gin.Context) {
	var body CreatePostDto
	op := "createPost"
	userId, ok := requireUserID(c)

	if !ok {
		return
	}

	// c.Bind` сам пишет 400 и игнорирует ошибку в твоём коде
	// c.ShouldBindJSON - тут отлавливается связь с required gin
	// Понять как он работает и все дела.
	if err := c.ShouldBindJSON(&body); err != nil {
		respondValidationError(c, op, err)
		return
	}

	ctx := c.Request.Context()
	post, err := controller.postService.CreatePost(
		ctx,
		body.Body,
		body.Title,
		userId,
	)
	
	if err != nil {
		respondServiceError(c, op, err)
		return
	}

	c.JSON(201, gin.H{
		"post": post,
	})
}

func (controller *PostController) updatePost(c *gin.Context) {
	userID, ok := requireUserID(c)

	if !ok {
		return
	}

	// Get id from URL
	var req PostIdPathParam
	op := "updatePost"
	// ShouldBindUri делает и биндинг, и конвертацию в int, и валидацию.
	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, op, err)
		return
	}
	
	var body UpdatePostDto
	
	// Get body from req
	if err := c.ShouldBindJSON(&body); err != nil {
		respondValidationError(c, op, err)
		return
	}
	
	ctx := c.Request.Context()
	post, err := controller.postService.UpdatePost(
		ctx, 
		UpdatePostParams{
			postBody: body.Body,
			postTitle: body.Title,
			postID: req.ID,
			userID: userID,
		},
	)

	if err != nil {
		respondServiceError(c, op, err)
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) deletePost(c *gin.Context) {
	userID, ok := requireUserID(c)

	if !ok {
		return
	}
	// Get id from URL
	var req PostIdPathParam
	op := "PostController.deletePost"

	if err := c.ShouldBindUri(&req); err != nil {
		respondInvalidID(c, op, err)
		return
	}

	ctx := c.Request.Context()
	deleteErr := controller.postService.DeletePost(
		ctx,
		DeletePostByIDParams{
			postID: req.ID,
			userID: userID,
		},
	)

	if deleteErr != nil {
		respondServiceError(c, op, deleteErr)
		return
	}

	c.JSON(200, gin.H{})
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
