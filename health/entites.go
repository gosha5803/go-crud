package health

import "database/sql"

type DBHealthResDTO struct {
	sql.DBStats
}