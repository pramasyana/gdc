package database

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// mysqlDuplicateEntry is the MySQL error number for a unique constraint
// violation.
const mysqlDuplicateEntry = 1062

// Translate converts GORM's record-not-found error into the kernel sentinel so
// repositories return transport-agnostic errors to usecases.
func Translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return kernel.ErrNotFound
	}
	return err
}

// IsDuplicate reports whether err is a MySQL unique-constraint violation.
func IsDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == mysqlDuplicateEntry
	}
	return false
}
