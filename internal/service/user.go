package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/GroVlAn/doc-store/internal/core/e"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(user core.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	if err := s.verifyNewUser(ctx, user); err != nil {
		return err
	}

	if err := s.createUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *Service) Auth(user core.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	userFromDB, err := s.UserRepo.User(ctx, user.Login)
	if err != nil {
		return "", fmt.Errorf("getting user: %w", err)
	}

	if err := s.verifyPassword(userFromDB, user); err != nil {
		return "", err
	}

	accessToken, err := s.createAccessToken(userFromDB)
	if err != nil {
		return "", err
	}

	if err = s.saveAccessToken(ctx, accessToken); err != nil {
		return "", err
	}

	return accessToken.Token, nil
}

func (s *Service) VerifyAccessToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	if err := s.checkExistToken(ctx, token); err != nil {
		return err
	}

	tokenDetails, err := s.parseToken(token)
	if err != nil {
		return err
	}

	if err := s.checkExpiredToken(ctx, token, tokenDetails); err != nil {
		return err
	}

	if err := s.checkUserByToken(ctx, tokenDetails); err != nil {
		return e.ErrUserNotFound
	}

	return nil
}

func (s *Service) Logout(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.DefaultTimeout)
	defer cancel()

	err := s.TokenRepo.DeleteToken(ctx, token)
	if err != nil {
		return fmt.Errorf("deleting token: %w", err)
	}

	return nil
}

func (s *Service) checkUserByToken(ctx context.Context, tokenDetails jwt.MapClaims) error {
	login, ok := tokenDetails["login"].(string)
	if !ok {
		return &e.ErrInvalidToken{Msg: "invalid token"}
	}

	_, err := s.UserRepo.User(ctx, login)
	if err != nil {
		return e.ErrUserNotFound
	}

	return nil
}

func (s *Service) checkExpiredToken(ctx context.Context, token string, tokenDetails jwt.MapClaims) error {
	exp := time.Unix(int64(tokenDetails["exp"].(float64)), 0)
	now := time.Now()

	if now.After(exp) {
		return s.deleteTokenWithError(ctx, token)
	}

	return nil
}

func (s *Service) checkExistToken(ctx context.Context, token string) error {
	if _, err := s.TokenRepo.Token(ctx, token); err != nil {
		return &e.ErrInvalidToken{Msg: "invalid token", Err: fmt.Errorf("getting token: %w", err)}
	}

	return nil
}

func (s *Service) parseToken(token string) (jwt.MapClaims, error) {
	tokenClaims := jwt.MapClaims{}

	jwtToken, err := jwt.ParseWithClaims(
		token,
		tokenClaims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(s.SecretKey), nil
		},
	)
	if err != nil {
		return jwt.MapClaims{}, &e.ErrInvalidToken{Msg: "invalid token", Err: fmt.Errorf("parsing access token: %w", err)}
	}

	tokenDetails, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, &e.ErrInvalidToken{Msg: "invalid token"}
	}

	return tokenDetails, nil
}

func (s *Service) deleteTokenWithError(ctx context.Context, token string) error {
	if err := s.TokenRepo.DeleteToken(ctx, token); err != nil {
		return fmt.Errorf("deleting access token: %w", err)
	}

	return &e.ErrInvalidToken{Msg: "invalid token"}
}

func (s *Service) verifyNewUser(ctx context.Context, user core.User) error {
	userFromDB, err := s.UserRepo.User(ctx, user.Login)
	if err != nil && err != e.ErrUserNotFound {
		return fmt.Errorf("getting user: %w", err)
	}
	if len(userFromDB.ID) > 0 {
		return e.ErrUserAlreadyExist
	}

	if len(user.Login) == 0 {
		return e.ErrInvalidLogin
	}

	valid := s.validatePassword(user.Password)
	if !valid {
		return e.ErrInvalidPassword
	}

	return nil
}

func (s *Service) createUser(ctx context.Context, user core.User) error {
	user.ID = uuid.NewString()
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), s.HashCost)
	if err != nil {
		return fmt.Errorf("generating password: %w", err)
	}
	user.Password = string(password)

	err = s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("creating new user: %w", err)
	}

	return nil
}

func (s *Service) validatePassword(pswd string) bool {
	if len(pswd) < 8 {
		return false
	}

	var (
		isNumber = false
		isLower  = false
		isUpper  = false
		isSymbol = false
	)

	for _, ch := range pswd {
		switch {
		case unicode.IsNumber(ch) && !isNumber:
			isNumber = true
		case unicode.IsLower(ch) && !isLower:
			isLower = true
		case unicode.IsUpper(ch) && !isUpper:
			isUpper = true
		case (unicode.IsPunct(ch) || unicode.IsSymbol(ch)) && !isSymbol:
			isSymbol = true
		}
	}

	return isNumber && isLower && isUpper && isSymbol
}

func (s *Service) verifyPassword(userFromDB, user core.User) error {
	err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return e.ErrInvalidPassword
	}
	if err != nil {
		return fmt.Errorf("comparing hash adn password: %w", err)
	}

	return nil
}

func (s *Service) createAccessToken(user core.User) (core.AccessToken, error) {
	accessToken := core.AccessToken{}
	accessToken.StartTTL = time.Now()
	accessToken.EndTTl = accessToken.StartTTL.Add(s.TokenEndTTL)

	payload := jwt.MapClaims{
		"user_id": user.ID,
		"login":   user.Login,
		"exp":     accessToken.EndTTl.Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := jwtToken.SignedString([]byte(s.SecretKey))
	if err != nil {
		return core.AccessToken{}, fmt.Errorf("creating access token: %w", err)
	}

	accessToken.ID = uuid.NewString()
	accessToken.Token = t
	accessToken.UserID = user.ID

	return accessToken, nil
}

func (s *Service) saveAccessToken(ctx context.Context, accessToken core.AccessToken) error {
	err := s.TokenRepo.CreateToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("creating new token: %w", err)
	}

	return nil
}
