package uaclient

import "errors"

var (
	ErrConnection    = errors.New("connection error")
	ErrAuthSecurity  = errors.New("authentication/security error")
	ErrNodeNotFound  = errors.New("node not found")
	ErrBadStatusCode = errors.New("bad OPC UA status code")
	ErrWriteRejected = errors.New("write rejected")
	ErrValidation    = errors.New("validation error")
)
