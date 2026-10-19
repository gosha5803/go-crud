package posts

import (
	"context"

	"github.com/gosha5803/go-crud/models"
)

type IPostRepository interface {
	CreatePost(context.Context, CreatePostDto) (models.Post, error)
	GetPostById(context.Context,uint) (models.Post, error)
	GetPosts(context.Context) ( []models.Post, error)
	UpdatePost(context.Context,uint, UpdatePostDto) (models.Post, error)
	DeletePostById(context.Context,uint) error
	IncrementViews( context.Context,  uint)(models.Post, error)
}

type PostService struct {
	repo IPostRepository
}

func NewPostService(repo IPostRepository) *PostService {
	return &PostService{repo: repo}
}

func (service *PostService) IncrementViews(ctx context.Context, postId uint)(models.Post, error) {
	return service.repo.IncrementViews(ctx, postId)
}

func (service *PostService) CreatePost(ctx context.Context,postDTO CreatePostDto) (models.Post, error) {
	return service.repo.CreatePost(ctx, postDTO)
}

func (service *PostService) GetPosts(ctx context.Context) ([]models.Post, error) {
	return service.repo.GetPosts(ctx)
}

func (service *PostService) GetPostById(ctx context.Context,id uint) (models.Post, error) {
	return service.repo.GetPostById(ctx,id)
}

// TODO если у нас есть возможность делать, например тело поста - пустой строкой, то тут надо чётко различать, когда пользователь не трогает тело
// и не хочет его редактировать и когда он пустую строку отправил
func (service *PostService) UpdatePost(ctx context.Context,id uint, post UpdatePostDto) (models.Post, error) {
	return service.repo.UpdatePost(ctx,id, post)
}

func (service *PostService) DeletePost(ctx context.Context,id uint) (error) {
	return service.repo.DeletePostById(ctx, id)
}
