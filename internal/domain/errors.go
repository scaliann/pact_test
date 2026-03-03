package domain

import "errors"

var (
	// Returned when the session does not exist.
	ErrSessionNotFound = errors.New("session not found")
	// Returned when input arguments are invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// Returned when the session is not ready to perform requested action.
	ErrSessionNotReady = errors.New("session is not authorized yet")
	// Returned when action requires non-authorized state, but session is already authorized.
	ErrSessionAuthorized = errors.New("session is already authorized")
)
