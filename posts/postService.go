package posts

import (
	"errors"
	"fmt"

	"github.com/gosha5803/go-crud/models"
	"gorm.io/gorm"
)

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{DB: db}
}

// TODO реализовать методы
// Инстанциировать классы в модуле что ли)

func (service *PostService) CreatePost(postDTO CreatePostDto) (models.Post, error) {
	newPost := models.Post{
		Title: postDTO.Title,
		Body:  postDTO.Body,
	}

	result := service.DB.Create(&newPost)

	if result.Error != nil {
		err := fmt.Errorf("service: createPost: %w", result.Error)
		return models.Post{}, err
	}

	return newPost, nil
}

func (service *PostService) GetPosts() ([]models.Post, error) {
	var posts []models.Post

	// Мутация и назначение через указатель
	if result := service.DB.Find(&posts); result.Error != nil {
		err := fmt.Errorf("service: getPosts: %w", result.Error)

		return []models.Post{}, err
	}

	return posts, nil
}

func (service *PostService) GetPostById(id string) (models.Post, error) {
	var post models.Post

	result := service.DB.First(&post, "id = ?", id)

	if result.Error != nil {
		// Неужели я только благодоря своей сентинел ошибке могу идентифицировать ошибку, когда пост не найден?
		// В простом формате без парсинга строк как будто да
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Post{}, fmt.Errorf("get post %s: %w", id, ErrPostNotFound)
		}

		return models.Post{}, fmt.Errorf("get post: %s. %w", id, result.Error)
	}

	return post, nil
}

func (service *PostService) UpdatePost(id string, post UpdatePostDto) (models.Post, error) {
	var existing models.Post
	// Get post to update
	result := service.DB.First(&existing, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Post{}, fmt.Errorf("update post %s: %w", id, ErrPostNotFound)
		}

		return models.Post{}, fmt.Errorf("update post: %s, %w", id, result.Error)
	}

	updateResult := service.DB.Model(&existing).Updates(models.Post{
		Title: post.Title,
		Body:  post.Body,
	})

	if updateResult.Error != nil {
		return models.Post{}, fmt.Errorf("update post: %s, %w", id, result.Error)
	}

	return existing, nil
}

func (service *PostService) DeletePost(id string) (bool, error) {
	result := service.DB.Delete(&models.Post{}, "id = ?", id)

	if result.Error != nil {
		err := fmt.Errorf("delete post %s: %w", id, result.Error)
		return false, err
	}

	if result.RowsAffected != 1 {
		err := fmt.Errorf("delete post %s: %w", id, ErrPostNotFound)
		return false, err
	}

	return true, nil
}
