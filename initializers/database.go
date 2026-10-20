package initializers

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error

	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		// TODO валидация .env. Создание конфиг сервиса
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB := GetPlainPool()

	// TODO конфиг сервис
	// В тестах, можно поднять БД, с малым количеством соединений и посмотреть как растут ожидающие воркеры в очереди
	sqlDB.SetMaxOpenConns(15)
	sqlDB.SetMaxIdleConns(8)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	sqlDB.SetConnMaxLifetime(20 * time.Minute)

}

func GetPlainPool() *sql.DB {
		sqlDB, sqlSwitchErr := DB.DB()

	if sqlSwitchErr != nil {
		panic("failed to set connections settings")
	}

	return sqlDB
}
