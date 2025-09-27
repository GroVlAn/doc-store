package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/GroVlAn/doc-store/internal/core/e"
	"github.com/google/uuid"
)

func (s *Service) CreateDocument(document core.Document, file []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	tokenDetails, err := s.parseToken(document.Token)
	if err != nil {
		return &e.ErrInvalidToken{Msg: "invalid token", Err: err}
	}

	document.ID = uuid.NewString()
	document.Created = time.Now()
	document.Grant = append(document.Grant, tokenDetails["login"].(string))

	existDocument, err := s.DocumentRepo.DocumentByName(ctx, tokenDetails["login"].(string), document.Name)
	if err != nil && !errors.Is(err, e.ErrNoDocuments) {
		return fmt.Errorf("getting error by name: %w", err)
	}

	if len(existDocument.ID) > 0 {
		err = s.DocumentRepo.DeleteDocument(ctx, tokenDetails["login"].(string), existDocument.ID)
		if err != nil {
			return fmt.Errorf("deleting document: %w", err)
		}
		s.Cache.Delete(s.Cache.GenerateKey(core.AddrCacheDocument, tokenDetails["user_id"].(string), document.ID))
	}

	err = s.DocumentRepo.CreateDocument(ctx, document)
	if err != nil {
		return fmt.Errorf("creating document: %w", err)
	}

	if file == nil {
		return nil
	}
	err = s.createFile(tokenDetails["user_id"].(string), document.Name, file)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	s.Cache.Set(s.Cache.GenerateKey(core.AddrCacheDocument, tokenDetails["user_id"].(string), document.ID), document)

	return nil
}

func (s *Service) Document(token string, documentID string) (core.Document, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	tokenDetails, err := s.parseToken(token)
	if err != nil {
		return core.Document{}, "", &e.ErrInvalidToken{Msg: "invalid token", Err: err}
	}

	var document core.Document

	documentCache, ok := s.Cache.Get(s.Cache.GenerateKey(core.AddrCacheDocument, tokenDetails["user_id"].(string), documentID))
	if ok {
		document = documentCache.(core.Document)
	} else {
		document, err = s.DocumentRepo.Document(ctx, tokenDetails["login"].(string), documentID)
		if err != nil {
			return core.Document{}, "", e.ErrNoDocuments
		}
	}

	s.Cache.Set(s.Cache.GenerateKey(core.AddrCacheDocument, tokenDetails["user_id"].(string), documentID), document)

	file, err := s.FileRepo.File(tokenDetails["user_id"].(string), document.Name)
	if err != nil {
		return core.Document{}, "", fmt.Errorf("loading file: %w", err)
	}

	return document, file, nil
}

func (s *Service) DocumentsList(token string, filter core.DocumentFilter) ([]core.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	tokenDetails, err := s.parseToken(token)
	if err != nil {
		return nil, &e.ErrInvalidToken{Msg: "invalid token", Err: err}
	}

	filter.Login = tokenDetails["login"].(string)

	documentsList, err := s.DocumentRepo.DocumentsList(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("getting documents list: %w", err)
	}

	return documentsList, nil
}

func (s *Service) DeleteDocument(token string, documentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	tokenDetails, err := s.parseToken(token)
	if err != nil {
		return err
	}

	document, err := s.DocumentRepo.Document(ctx, tokenDetails["login"].(string), documentID)
	if err != nil {
		return e.ErrNoDocuments
	}

	if err := s.deleteFile(tokenDetails["user_id"].(string), document.Name); err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	err = s.DocumentRepo.DeleteDocument(ctx, tokenDetails["login"].(string), documentID)
	if err != nil {
		return fmt.Errorf("deleting document: %w", err)
	}

	s.Cache.Delete(s.Cache.GenerateKey(core.AddrCacheDocument, tokenDetails["user_id"].(string), documentID))

	return nil
}

func (s *Service) setLoginToFilter(ctx context.Context, userID string, filter *core.DocumentFilter) error {
	if len(filter.Login) == 0 {
		return nil
	}

	user, err := s.UserRepo.User(ctx, userID)
	if err != nil {
		return e.ErrUserNotFound
	}

	filter.Login = user.Login

	return nil
}

func (s *Service) createFile(userID string, fileName string, file []byte) error {
	if err := s.deleteExistedFile(userID, fileName); err != nil {
		return fmt.Errorf("deleting existed file: %w", err)
	}

	err := s.FileRepo.SaveFile(userID, fileName, file)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	return nil
}

func (s *Service) deleteExistedFile(userID string, fileName string) error {
	if !s.FileRepo.FileExist(userID, fileName) {
		return nil
	}

	return s.deleteFile(userID, fileName)
}

func (s *Service) deleteFile(userID string, fileName string) error {
	if err := s.FileRepo.DeleteFile(userID, fileName); err != nil {
		return err
	}

	return nil
}
