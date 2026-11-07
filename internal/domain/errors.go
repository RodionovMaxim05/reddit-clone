package domain

import "errors"

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidToken      = errors.New("invalid token")
)
