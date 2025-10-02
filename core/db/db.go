package db

import (
	"fmt"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	// TODO(lebogg): Eventually, make this unexported
	DB *gorm.DB
}

func NewStorage() (*Storage, error) {
	g, err := gorm.Open(sqlite.Open(":memory:?_pragma=foreign_keys(1)"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, fmt.Errorf("could not open in-memory sqlite database: %w", err)
	}

	return &Storage{DB: g}, nil
}

func (s *Storage) Create(r any) (err error) {
	err = s.DB.Create(r).Error

	if err != nil && (strings.Contains(err.Error(), "constraint failed: UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "duplicate key value violates unique constraint")) {
		return fmt.Errorf("could not create record: unique constraint failed")
	}

	if err != nil && strings.Contains(err.Error(), "constraint failed") {
		return fmt.Errorf("could not create record: constraint failed")
	}

	return
}

// List retrieves records from the database with optional ordering, offset, limit, and conditions.
func (s *Storage) List(r any, orderBy string, asc bool, offset int, limit int, conds ...any) error {
	var query = s.DB
	// Set default direction to "ascending"
	var orderDirection = "asc"

	if limit != -1 {
		query = s.DB.Limit(limit)
	}

	// Set direction to "descending"
	if !asc {
		orderDirection = "desc"
	}
	orderStmt := orderBy + " " + orderDirection
	// No explicit ordering
	if orderBy == "" {
		orderStmt = ""
	}

	return query.Order(orderStmt).Find(r, conds...).Error
}
