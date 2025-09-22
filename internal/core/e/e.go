package e

import "errors"

var ErrInvalidPassword = errors.New("invalid password")
var ErrInvalidLogin = errors.New("invalid login")
var ErrUserAlreadyExist = errors.New("user already exist")
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidToken = errors.New("invalid token")

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
