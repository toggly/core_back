package api

import (
	"github.com/Toggly/core/internal/pkg/storage"
)

// Engine type
type Engine struct {
	Storage *storage.DataStorage
}

// ForOwner returns owner api
func (e *Engine) ForOwner(owner string) *OwnerAPI {
	return &OwnerAPI{Owner: owner, Storage: e.Storage}
}

// OwnerAPI type
type OwnerAPI struct {
	Owner   string
	Storage *storage.DataStorage
}

// Projects returns project api
func (o *OwnerAPI) Projects() *ProjectAPI {
	return &ProjectAPI{*o}
}
