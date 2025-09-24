package e

import "errors"

var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidLogin     = errors.New("invalid login")
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidToken     = errors.New("invalid token")
	ErrNoDocuments      = errors.New("documents not found")
)

type ErrInsert struct {
	Msg string
	Err error
}

func (ei *ErrInsert) Error() string {
	return ei.Msg
}

func (ei *ErrInsert) Unwrap() error {
	return ei.Err
}

type ErrFind struct {
	Msg string
	Err error
}

func (ef *ErrFind) Error() string {
	return ef.Msg
}

func (ef *ErrFind) Unwrap() error {
	return ef.Err
}

type ErrDelete struct {
	Msg string
	Err error
}

func (ed *ErrDelete) Error() string {
	return ed.Msg
}

func (ed *ErrDelete) Unwrap() error {
	return ed.Err
}
