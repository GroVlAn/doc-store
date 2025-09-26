package e

import "errors"

var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidLogin     = errors.New("invalid login")
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrNoDocuments      = errors.New("documents not found")

	ErrEmptyBody = errors.New("empty data")
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

type ErrInvalidToken struct {
	Msg string
	Err error
}

func (eit *ErrInvalidToken) Error() string {
	return eit.Msg
}

func (eit *ErrInvalidToken) Unwrap() error {
	return eit.Err
}
