package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	aichat "github.com/gosha5803/go-crud/aiChat"
	"github.com/gosha5803/go-crud/auth"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
	"github.com/gosha5803/go-crud/posts"
)

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

	// отмена этого контекста, идущего в модуль позволяет воркерам начать быстро разбирать оставшуюся очередь
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	aiChatModule, err := aichat.NewAiChatModule(appCtx, r)
	if err != nil {
		log.Printf("main: %v", err)
		return
	}
	authModule := auth.NewAuthModule(r, initializers.DB)

	runModules([]models.Module{
		authModule,
		aiChatModule,
		posts.NewPostModule(r, initializers.DB),
	})

	srv := &http.Server{
		Addr:    "4020",
		Handler: r,
	}

	go srv.ListenAndServe()

	<-appCtx.Done()

	// Это контекст, чтобы прям убить работу модуля
	// создаётся не от App, а от Background()
	// TODO немного странно, что какую-то мидисекунду, канал ещё не закрыт, а воркеры уже опустошают его в быстром моде. Нельзя ли завязать всё на 1 контекст?
	// Хотя почти всё на это и завязано, так как второй контекст сразу же создаётся и вызывается Close
	shutDownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	srv.Shutdown(shutDownCtx)
	authModule.Close(shutDownCtx)

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
