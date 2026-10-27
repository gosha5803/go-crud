package posts

import (
	"context"
	"errors"
	"fmt"

	"github.com/gosha5803/go-crud/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostRepository struct {
	DB *gorm.DB
}

func NewPostRepository(DB *gorm.DB) *PostRepository {
	return &PostRepository{DB: DB}
}

func (r *PostRepository) CreatePost(
	ctx context.Context, 
	postBody, 
	postTitle string,
	userId uint,
) (models.Post, error) {
	op := "PostRepository.CreatePost"

	newPost := models.Post{
		Title: postTitle,
		Body:  postBody,
		AuthorID: userId,
	}

	result := r.DB.
		WithContext(ctx).
// Указатель чтобы горм прокинул сгенерированные даты, айди и т.п.
		Create(&newPost)

	if result.Error != nil {
		err := fmt.Errorf("%s: %w", op, result.Error)
		return models.Post{}, err
	}

	return newPost, nil
}

func (r *PostRepository) GetPosts(ctx context.Context) ([]models.Post, error) {
	op := "PostRepository.GetPosts"

	var posts []models.Post
	// TODO пагинация, лимит, офсет, сортировка
	if err := r.DB.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return posts, nil
}

// "ID": 10,
//         "CreatedAt": "2026-10-20T17:57:38.154636+03:00",
//         "UpdatedAt": "2026-10-20T17:57:38.154636+03:00",
//         "DeletedAt": null,
//         "title": "Первый пост пользователя (6)",
//         "body": "Тело поста 1 (id = 9) пользователя (6)",
//         "AuthorID": 6,
//         "views": 0,
// TODO при чтении постов, не все поля надо возвращать
// Наверное для этого нужно DTO слой добавит для овтетов)
func (r *PostRepository) GetPostById(ctx context.Context, postID uint) (models.Post, error) {
	op := "PostRepository.GetPostById"
	var post models.Post

	if err := r.DB.
		WithContext(ctx).
		// Сканирует в указатель
		First(&post, "id = ?", postID).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Post{}, fmt.Errorf("%s %d: %w", op, postID, ErrPostNotFound)
		}

		return models.Post{}, fmt.Errorf("%s %d: %w", op, postID, err)
	}

	return post, nil

}

func (r *PostRepository) DeletePostById(
	ctx context.Context,
	params DeletePostByIDParams,
) error {
	op := "PostRepository.DeletePostById"
	
	result := r.DB.
	WithContext(ctx).
	Where("id = ? AND author_id = ?", params.postID, params.userID).
	Delete(&models.Post{}) 
	
	if result.Error != nil {
		return fmt.Errorf("%s %d: %w", op, params.postID, result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s %d: %w", op, params.postID, ErrPostNotFound)
	}

	return nil
}

func (r *PostRepository) UpdatePost(
	ctx context.Context,
	params UpdatePostParams,
) (models.Post, error) {
	op := "PostRepository.UpdatePost"
	var post models.Post

	if err := r.DB.
	WithContext(ctx).
	Where("id = ? AND author_id = ?", params.postID, params.userID).
	First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Post{}, fmt.Errorf("%s %d: %w", op, params.postID, ErrPostNotFound)
		}

		return models.Post{}, fmt.Errorf("%s: %d, %w", op, params.postID, err)
	}

	updates := map[string]any{}

	if params.postBody != nil {updates["body"] = *params.postBody}
	if params.postTitle!= nil {updates["title"] = *params.postTitle}

	if len(updates) == 0 {
		return post, nil
	}

	result := r.DB.
	WithContext(ctx).
	Model(&post).
	Updates(updates)
	
	if result.Error != nil {
		return models.Post{}, fmt.Errorf("%s: %d, %w", op, params.postID, result.Error)
	}

	if result.RowsAffected == 0 {
		return models.Post{}, fmt.Errorf("%s %d: %w", op, params.postID, ErrPostNotFound)
	}

	return post, nil
}

func (r *PostRepository) IncrementViews(ctx context.Context, postID uint)(models.Post, error) {
	op := "PostRepository.IncrementViews"
	var post models.Post

	// DB - доступ к пулу
	result := r.DB.
	// передача контекста для отмены запроса ожидающего соединения по таймауту либо, если клиент ушёл, запрос abort'нулся
	WithContext(ctx).
	// указание таблицы, с которой GORM будет работать. Он достанет её из модели. 
	// Также в данном случае это одновременно и указание куда сканировать результат. Так как мы указали Clauses. 
	Model(&post).
	// в данном случае именно добавляет конструкцию Returning в postgres id, title, body, views
	// По умолчанию Update не возвращает обновлённую строку, а Clause говорит gorm вернуть обновлённую строку
	// и записывает эту строку в &post 
	Clauses(clause.Returning{}).
	// Условие лдя выбора строки для обновления
	// ? плейсхолдер - защита от SQL инъекций
	Where("id = ?", postID).
	// Update(col, val) обновляет 1 колонку
	// Updates(...) бновляет много колонок по мапе или структуре
	// UpdateColumn(col, val) также 1 колонку, но только её в обход Updated at
	// UpdateColumns(...) много колонок, но только в обход Updated at
	// Просмотр поста не является его редактированием, поэтому UpdateCollumn
	UpdateColumn(
		"views", 
		// Сырое SQl выражение, gorm подставит как есть
		gorm.Expr("views + 1"),
)

	if result.Error != nil {
		// Rows affected может быть ноль, при отмене контекста по истечении таймаута, и будет ошибка. как будто не найден пост
		return models.Post{}, fmt.Errorf("%s: %d %w", op, postID, result.Error)
	}

	// Not Found ошибка происходит только от first, take, last 
	if result.RowsAffected == 0 {
		return models.Post{}, fmt.Errorf("%s: %d %w", op, postID, ErrPostNotFound)
	}

	return post, nil

}