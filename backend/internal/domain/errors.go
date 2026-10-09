package domain

import "errors"

// ErrNotFound indicates that the requested entity does not exist.
// Repository implementations must return this (via errors.Is) for missing rows,
// and must not wrap unrelated database failures as not-found.
var ErrNotFound = errors.New("not found")

// ErrConflict indicates a unique constraint violation or equivalent conflict.
var ErrConflict = errors.New("conflict")
