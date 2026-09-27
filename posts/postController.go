package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/models"
)

// Go-сообщество склоняется к тому, что интерфейс должен объявляться там, где он потребляется, а не там, где реализуется.
type IPostService interface {
	CreatePost(post CreatePostDto) models.Post
	GetPosts() []models.Post
	GetPostById(id string) models.Post
	UpdatePost(id string, post CreatePostDto) models.Post
	DeletePost(id string) bool
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
	posts := controller.postService.GetPosts()

	c.JSON(200, gin.H{
		"posts": posts,
	})
}

func (controller *PostController) getPostById(c *gin.Context) {
	id := c.Param("id")

	post := controller.postService.GetPostById(id)

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) createPost(c *gin.Context) {
	var body CreatePostDto

	c.Bind(&body)

	post := controller.postService.CreatePost(body)

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

	post := controller.postService.UpdatePost(id, body)

	c.JSON(200, gin.H{
		"post": post,
	})
}

func (controller *PostController) deletePost(c *gin.Context) {
	// Get id from URL
	id := c.Param("id")

	success := controller.postService.DeletePost(id)

	c.JSON(200, gin.H{
		"Success": success,
	})
}
