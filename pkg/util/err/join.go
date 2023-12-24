package errutil

import (
	"errors"
	"strings"
)

type joinedErrors []error

func Join(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	if n == 1 {
		for _, err := range errs {
			if err != nil {
				return err
			}
		}
	}
	joinedErrors := make(joinedErrors, 0, n)
	for _, err := range errs {
		if err != nil {
			joinedErrors = append(joinedErrors, err)
		}
	}
	return joinedErrors
}

func (e joinedErrors) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}

	sb := strings.Builder{}
	sb.WriteString(e[0].Error())
	for _, err := range e[1:] {
		sb.WriteString(", ")
		sb.WriteString(err.Error())
	}
	return sb.String()
}

//goland:noinspection GoStandardMethods
func (e joinedErrors) Unwrap() []error {
	return e
}

// Is returns true if any of the errors in the joinedErrors is target (according to errors.Is() logic).
// It's needed make joinedErrors compatible with errors.Is()
func (e joinedErrors) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As makes joinedErrors compatible with errors.As()
func (e joinedErrors) As(target any) bool {
	for _, err := range e {
		if //goland:noinspection GoErrorsAs
		errors.As(err, target) {
			return true
		}
	}
	return false
}
