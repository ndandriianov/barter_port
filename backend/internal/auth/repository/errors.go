package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// UniqueError represents a uniqueness constraint violation error.
type UniqueError struct {
	Constraint string // Constraint is the name of the database constraint that was violated.
	Err        error  // Err is the original error.
}

// Error returns a string representation of the UniqueError.
func (e *UniqueError) Error() string {
	return fmt.Sprintf("uniqueness constraint: %s was violated", e.Constraint)
}

// Unwrap returns the original error wrapped by UniqueError.
func (e *UniqueError) Unwrap() error {
	return e.Err
}

// TryCastToUniqueViolation checks if the given error is a uniqueness constraint violation.
// If a violation is detected, it returns a UniqueError with the provided field and value for better readability.
// Otherwise, it returns the original error.
func TryCastToUniqueViolation(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return &UniqueError{
			Constraint: pgErr.ConstraintName,
			Err:        err,
		}
	}
	return err
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
