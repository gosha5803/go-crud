package main

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	aichat "github.com/gosha5803/go-crud/aiChat"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
	"github.com/gosha5803/go-crud/posts"
)

// init функция выполняется до main
func init() {
	initializers.LoadEnv()
	initializers.ConnectToDB()
}

func runModules(modules []models.Module) {
	for _, module := range modules {
		module.Init()
	}
}

func main() {

	r := gin.Default()

	r.Use(cors.Default())

	ctx := context.Background()

	runModules([]models.Module{posts.NewPostModule(r, initializers.DB), aichat.NewAiChatModule(ctx, r)})

	r.Run()

}

/*
	🔴 P0 — Критичные баги (ломают работу)
Обработка ошибок в сервисе
□ CreatePost — errors.New(...) создаётся, но не возвращается. Нужно изменить сигнатуру на (models.Post, error).
□ UpdatePost — возвращает старые данные, потому что GORM не синхронизирует структуру после Updates. Либо перечитывать запись, либо использовать map.
□ UpdatePost — Updates со структурой игнорирует пустые поля. Если нужно очистить Body, используйте map[string]interface{}.
□ DeletePost — всегда возвращает true, даже если записи не было. Проверяйте result.RowsAffected и result.Error.
□ GetPostById — не обрабатывает ErrRecordNotFound. Возвращает пустую структуру вместо ошибки.
□ GetPosts — не проверяет result.Error.
□ GetPostById / UpdatePost — используйте First(&post, "id = ?", id) вместо First(&post, id), чтобы GORM не трактовал строку как условие.
Валидация входа в контроллере
□ createPost / updatePost — c.Bind(&body) игнорирует ошибку. Заменить на c.ShouldBindJSON(&body) с проверкой и возвратом 400 Bad Request.
□ Добавить валидацию DTO через теги binding:"required" и проверку ShouldBindJSON.
Синхронизация интерфейса и реализации
□ IPostController содержит неэкспортированные методы — *PostController не реализует его за пределами пакета. Оставить только AssignRoutes().
□ IPostService — сигнатуры изменятся после добавления error. Обновить интерфейс и все реализации.
*/
// Валидация .env
