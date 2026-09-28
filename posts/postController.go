package posts

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/models"
)

// Go-сообщество склоняется к тому, что интерфейс должен объявляться там, где он потребляется, а не там, где реализуется.
type IPostService interface {
	CreatePost(post CreatePostDto) (models.Post, error)
	GetPosts() ([]models.Post, error)
	GetPostById(id string) (models.Post, error)
	UpdatePost(id string, post CreatePostDto) (models.Post, error)
	DeletePost(id string) (bool, error)
}

type PostController struct {
	g *gin.Engine
	// Вот тут контроллер знает о типе сервиса
	// По сути тип должны быть на уровне портов,
	// и неявно реализовываться адаптерами
	postService IPostService
}

func NewPostController(g *gin.Engine, service IPostService) *PostController {
	return &PostController{g: g, postService: service}
}

func (controller *PostController) AssignRoutes() {

	post := controller.g.Group("/post")
	{
		post.GET("", controller.getPosts)
		// В nest удобно было задавать контроллеры и группы методов их,
		// как тут использовать общий префикс - не понятно
		post.POST("", controller.createPost)
		post.GET("/:id", controller.getPostById)
		post.PATCH("/:id", controller.updatePost)
		post.DELETE("/:id", controller.deletePost)
		// controller.g.POST("/chat", controller.ChatHandler)
	}
}

func (controller *PostController) getPosts(c *gin.Context) {
	posts, err := controller.postService.GetPosts()

	if err != nil {
		log.Printf("controller: getPosts: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"posts": posts,
	})
}

func (controller *PostController) getPostById(c *gin.Context) {
	id := c.Param("id")

	post, err := controller.postService.GetPostById(id)

	if err != nil {
		log.Printf("controller: get post by id: %v", err)

		if errors.Is(err, ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) createPost(c *gin.Context) {
	var body CreatePostDto

	// c.Bind` сам пишет 400 и игнорирует ошибку в твоём коде
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("controller: create post: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	post, err := controller.postService.CreatePost(body)

	if err != nil {
		log.Printf("controller: create post: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}

	c.JSON(201, gin.H{
		"post": post,
	})
}

func (controller *PostController) updatePost(c *gin.Context) {
	// Get id from URL
	id := c.Param("id")

	var body CreatePostDto

	// Get body from req
	c.Bind(&body)

	post, err := controller.postService.UpdatePost(id, body)

	if err != nil {
		log.Printf("controller: update post: %v", err)

		if errors.Is(err, ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) deletePost(c *gin.Context) {
	// Get id from URL
	id := c.Param("id")

	success, err := controller.postService.DeletePost(id)

	if err != nil {
		log.Printf("controller: delete post: %v", err)

		if errors.Is(err, ErrCouldNotDeletePost) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"Success": success,
	})
}
