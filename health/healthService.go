package health

import (
	"fmt"

	"gorm.io/gorm"
)

type HealthService struct {
	DB *gorm.DB
}

func NewHealthService(DB *gorm.DB) *HealthService {
	return &HealthService{DB: DB}
}

func (s *HealthService) GetDBStats() (DBHealthResDTO, error) {
	op := "HealthService.GetDBStats"
	plainSQL, err := s.DB.DB()

	if err != nil {
		return DBHealthResDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	
	pingErr := plainSQL.Ping()
	
	if pingErr != nil {
		return DBHealthResDTO{}, fmt.Errorf("%s: %w", op, pingErr)
	}

	stats := plainSQL.Stats()
	
	return DBHealthResDTO{DBStats: stats}, nil
}