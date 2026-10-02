package main

import (
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
	"github.com/gosha5803/go-crud/user"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectToDB()
}

// stop 12 16
// Этот скрипт запускается отдельно, не используется во время рантайма,
// а выполняется в момент, кгода надо применить миграции
func main() {
	// Тут обращается к переменной,
	// но перед этим он задачёт ей значение через вызов
	// initializers.ConnectToDB()
	// В файле database мы этой переменной назначили возвращаемое значение connectToDb().
	initializers.DB.AutoMigrate(&models.Post{})
	// Мигрировал не ту модель
	// Миграции хранятся историей?
	initializers.DB.AutoMigrate(&user.User{})
}
