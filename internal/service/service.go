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

type userRepo interface {
	CreateUser(ctx context.Context, user core.User) error
	User(ctx context.Context, login string) (core.User, error)
}

type tokenRepo interface {
	CreateToken(ctx context.Context, token core.AccessToken) error
	Token(ctx context.Context, token string) (core.AccessToken, error)
	DeleteToken(ctx context.Context, token string) error
}

type Deps struct {
	UserRepo       userRepo
	TokenRepo      tokenRepo
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

	if err := s.verifyPassword(ctx, user); err != nil {
		return "", err
	}

	accessToken, err := s.createAccessToken(user)
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

func (s *Service) checkUserByToken(ctx context.Context, tokenDetails jwt.MapClaims) error {
	userID, ok := tokenDetails["user_id"].(string)
	if !ok {
		return e.ErrInvalidToken
	}

	_, err := s.UserRepo.User(ctx, userID)
	if err != nil {
		return e.ErrUserNotFound
	}

	return nil
}

func (s *Service) checkExpiredToken(ctx context.Context, token string, tokenDetails jwt.MapClaims) error {
	exp := time.Unix(tokenDetails["exp"].(int64), 0)
	now := time.Now()

	if now.After(exp) {
		return s.deleteTokenWithError(ctx, token)
	}

	return nil
}

func (s *Service) checkExistToken(ctx context.Context, token string) error {
	accessToken, err := s.TokenRepo.Token(ctx, token)
	if err != nil {
		return fmt.Errorf("getting token: %w", err)
	}
	if len(accessToken.ID) == 0 {
		return e.ErrInvalidToken
	}

	return nil
}

func (s *Service) parseToken(token string) (jwt.MapClaims, error) {
	tokenClaims := jwt.MapClaims{}

	jwtToken, err := jwt.ParseWithClaims(
		token,
		tokenClaims,
		func(jwtToken *jwt.Token) (interface{}, error) {
			return s.SecretKey, nil
		},
	)
	if err != nil {
		return jwt.MapClaims{}, fmt.Errorf("parsing access token: %w", err)
	}

	tokenDetails, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, e.ErrInvalidToken
	}

	return tokenDetails, nil
}

func (s *Service) deleteTokenWithError(ctx context.Context, token string) error {
	if err := s.TokenRepo.DeleteToken(ctx, token); err != nil {
		return fmt.Errorf("deleting access token: %w", err)
	}

	return e.ErrInvalidToken
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

func (s *Service) verifyPassword(ctx context.Context, user core.User) error {
	userFromDB, err := s.UserRepo.User(ctx, user.Login)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password))
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
