package initializers

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB(dbConfig *DBConfig) {
	var err error

	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.DBPort,
		dbConfig.DBUser,
		dbConfig.DBPassword,
		dbConfig.DBName,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB := GetPlainPool()

	sqlDB.SetMaxOpenConns(dbConfig.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(dbConfig.DBMaxIdleConns)
	sqlDB.SetConnMaxIdleTime(dbConfig.DBConnIdleTime)
	sqlDB.SetConnMaxLifetime(dbConfig.DBConnIdleTime)
}

func GetPlainPool() *sql.DB {
	sqlDB, sqlSwitchErr := DB.DB()

	if sqlSwitchErr != nil {
		panic("failed to set connections settings")
	}

	return sqlDB
}
