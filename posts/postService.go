package posts

import (
	"context"
	"log/slog"

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
	log *slog.Logger
}

func NewPostService(repo IPostRepository, log *slog.Logger) *PostService {
	return &PostService{repo: repo, log: log}
}

func (service *PostService) IncrementViews(ctx context.Context, postId uint)(models.Post, error) {
	op := "PostService.IncrementViews"

	log := service.log.With(
		slog.String("op", op),
		slog.Int("id", int(postId)),
	)

	log.Info("Trying to increment the post views")

	return service.repo.IncrementViews(ctx, postId)
}

func (service *PostService) CreatePost(ctx context.Context, postDTO CreatePostDto) (models.Post, error) {
	op := "PostService.CreatePost"
	log := service.log.With(
		slog.String("op", op),
		slog.String("postBody", postDTO.Body),
		slog.String("postTitle", postDTO.Title),
		slog.Int("postAuthorID", int(postDTO.AuthorId)),
	)

	log.Info("Trying to create a post")

	return service.repo.CreatePost(ctx, postDTO)
}

func (service *PostService) GetPosts(ctx context.Context) ([]models.Post, error) {
	op := "PostService.GetPosts"
	log := service.log.With(
		slog.String("op", op),
	)

	log.Info("Trying to Get posts")

	return service.repo.GetPosts(ctx)
}

func (service *PostService) GetPostById(ctx context.Context,id uint) (models.Post, error) {
	op := "PostService.GetPostById"
	log := service.log.With(
		slog.String("op", op),
		slog.Int("postId", int(id)),
	)

	log.Info("Trying to find a post by id")

	return service.repo.GetPostById(ctx,id)
}

// TODO если у нас есть возможность делать, например тело поста - пустой строкой, то тут надо чётко различать, когда пользователь не трогает тело
// и не хочет его редактировать и когда он пустую строку отправил
func (service *PostService) UpdatePost(ctx context.Context, id uint, post UpdatePostDto) (models.Post, error) {
	op := "PostService.UpdatePost"
	log := service.log.With(
		slog.String("op", op),
		slog.String("postBody", *post.Body),
		slog.String("postTitle", *post.Title),
	)

	log.Info("Trying to update a post")

	return service.repo.UpdatePost(ctx,id, post)
}

func (service *PostService) DeletePost(ctx context.Context, id uint) (error) {
	op := "PostService.DeletePost"
	log := service.log.With(
		slog.String("op", op),
		slog.Int("id", int(id)),
	)

	log.Info("Trying to delete the post")

	return service.repo.DeletePostById(ctx, id)
}
