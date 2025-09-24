package service

import (
	"context"
	"time"

	"github.com/GroVlAn/doc-store/internal/core"
)

type userRepo interface {
	CreateUser(ctx context.Context, user core.User) error
	User(ctx context.Context, login string) (core.User, error)
}

type tokenRepo interface {
	CreateToken(ctx context.Context, token core.AccessToken) error
	Token(ctx context.Context, token string) (core.AccessToken, error)
	DeleteToken(ctx context.Context, token string) error
}

type documentRepo interface {
	CreateDocument(ctx context.Context, document core.Document) error
	Document(ctx context.Context, userID string, documentID string) (core.Document, error)
	DocumentsList(ctx context.Context, df core.DocumentFilter) ([]core.Document, error)
	DeleteDocument(ctx context.Context, userID string, documentID string) error
}

type fileRepo interface {
	SaveFile(userID string, fileName string, file []byte) error
	File(userID string, fileName string) (string, error)
	DeleteFile(userID string, fileName string) error
	FileExist(userID string, fileName string) bool
}

type Deps struct {
	UserRepo       userRepo
	TokenRepo      tokenRepo
	DocumentRepo   documentRepo
	FileRepo       fileRepo
	DefaultTimeout time.Duration
	HashCost       int
	TokenEndTTL    time.Duration
	SecretKey      string
}

type Service struct {
	Deps
}

func New(deps Deps) *Service {
	return &Service{
		Deps: deps,
	}
}
