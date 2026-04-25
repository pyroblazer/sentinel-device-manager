package middleware

import "errors"

// Sentinel errors for authentication and authorization failures.
var (
	ErrInvalidToken     = errors.New("invalid token format")
	ErrInvalidSignature = errors.New("token signature verification failed")
	ErrExpiredToken     = errors.New("token has expired")
	ErrNoClaims         = errors.New("no claims in context")
)
