package posts

import (
	"context"
	"log/slog"

	"github.com/gosha5803/go-crud/models"
)

type IPostRepository interface {
	CreatePost(context.Context, string, string, uint) (models.Post, error)
	GetPostById(context.Context,uint) (models.Post, error)
	GetPosts(context.Context) ( []models.Post, error)
	UpdatePost(UpdatePostParams) (models.Post, error)
	DeletePostById(DeletePostByIDParams) error
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

func (service *PostService) CreatePost(
	ctx context.Context, 
	postBody, 
	postTitle string,
	userId uint,
	) (models.Post, error) {
	op := "PostService.CreatePost"
	log := service.log.With(
		slog.String("op", op),
		slog.String("postBody", postBody),
		slog.String("postTitle", postTitle),
		slog.Int("postAuthorID", int(userId)),
	)

	log.Info("Trying to create a post")

	return service.repo.CreatePost(
		ctx, 
		postBody, 
		postTitle,
		userId,
	)
}

func (service *PostService) GetPosts(ctx context.Context) ([]models.Post, error) {
	op := "PostService.GetPosts"
	log := service.log.With(
		slog.String("op", op),
	)

	log.Info("Trying to Get posts")

	return service.repo.GetPosts(ctx)
}

func (service *PostService) GetPostById(ctx context.Context, id uint) (models.Post, error) {
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
func (service *PostService) UpdatePost(
	params UpdatePostParams,
) (models.Post, error) {
	op := "PostService.UpdatePost"
	log := service.log.With(
		slog.String("op", op),
		slog.Any("postID", params.postID),
		slog.Any("userID", params.userID),
	)

	log.Info("Trying to update a post")

	return service.repo.UpdatePost(params)
}

func (service *PostService) DeletePost(params DeletePostByIDParams) (error) {
	op := "PostService.DeletePost"
	log := service.log.With(
		slog.String("op", op),
		slog.Int("postID", int(params.postID)),
		slog.Int("userID", int(params.userID)),
	)

	log.Info("Trying to delete the post")

	return service.repo.DeletePostById(params)
}
