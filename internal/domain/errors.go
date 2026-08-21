package domain
import "errors"
var (
	ErrNotFound  = errors.New("resource not found")
	ErrInvalid   = errors.New("invalid input")
	ErrConflict  = errors.New("resource conflict")
	ErrForbidden = errors.New("operation forbidden")
)
