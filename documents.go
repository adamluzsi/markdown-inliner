package markdowninliner

import (
	"context"
	"os"
)

type Document struct {
	FilePath string `ext:"id"`
	FileMode os.FileMode
}

type DocumentStore interface {
	Create(context.Context, *Document) error
	FindByID(ctx context.Context, ptr *Document, filePath string) (bool, error)
	DeleteByID(ctx context.Context, filePath string) error
}
