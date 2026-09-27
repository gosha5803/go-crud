package main

import (
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/models"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectToDB()
}

// Этот скрипт запускается отдельно, не используется во время рантайма,
// а выполняется в момент, кгода надо применить миграции
func main() {
	// Тут обращается к переменной,
	// но перед этим он задачёт ей значение через вызов
	// initializers.ConnectToDB()
	// В файле database мы этой переменной назначили возвращаемое значение connectToDb().
	initializers.DB.AutoMigrate(&models.Post{})
}
