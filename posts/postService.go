package posts

import (
	"errors"

	"github.com/gosha5803/go-crud/models"
	"gorm.io/gorm"
)

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{DB: db}
}

// TODO реализовать методы
// Инстанциировать классы в модуле что ли)

func (service *PostService) CreatePost(postDTO CreatePostDto) models.Post {
	newPost := models.Post{
		Title: postDTO.Title,
		Body:  postDTO.Body,
	}

	result := service.DB.Create(&newPost)

	if result.Error != nil {
		errors.New("Error while creating a post")
	}

	return newPost
}

func (service *PostService) GetPosts() []models.Post {
	var posts []models.Post

	// Мутация и назначение через указатель
	service.DB.Find(&posts)

	return posts
}

func (service *PostService) GetPostById(id string) models.Post {
	var post models.Post

	service.DB.First(&post, id)

	return post
}

func (service *PostService) UpdatePost(id string, post CreatePostDto) models.Post {
	var existing models.Post
	// Get post to update
	service.DB.First(&existing, id)

	service.DB.Model(&existing).Updates(models.Post{
		Title: post.Title,
		Body:  post.Body,
	})

	return existing
}

func (service *PostService) DeletePost(id string) bool {
	service.DB.Delete(&models.Post{}, id)

	return true
}
