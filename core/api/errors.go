package api

import (
	"errors"
)

var (
	ErrInvalidColumnName     = errors.New("column name is invalid")
	ErrEmptyRequest          = errors.New("empty request")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrCatalogIdIsMissing    = errors.New("catalog_id is missing")
	ErrCategoryNameIsMissing = errors.New("category_name is missing")
	ErrControlIdIsMissing    = errors.New("control_id is missing")
	ErrControlNotAvailable   = errors.New("control not available")
	ErrAuditScopeNotFound    = errors.New("audit scope not found")
)
