package service

import (
	"context"
	"goapptemp/internal/shared/exception"
	"regexp"
	"strings"

	mysql "goapptemp/internal/adapter/repository/mysql"

	"github.com/cockroachdb/errors"
)

func getDetailedRepoMessage(fullRepoError error, associatedBaseError error) string {
	currentErr := fullRepoError
	for currentErr != nil {
		unwrapped := errors.Unwrap(currentErr)
		if errors.Is(unwrapped, associatedBaseError) {
			errMsg := currentErr.Error()
			baseErrMsgSuffix := ": " + associatedBaseError.Error()

			if strings.HasSuffix(errMsg, baseErrMsgSuffix) {
				return strings.TrimSuffix(errMsg, baseErrMsgSuffix)
			}

			return errMsg
		}

		if errors.Is(currentErr, associatedBaseError) {
			return associatedBaseError.Error()
		}

		currentErr = unwrapped
	}

	return associatedBaseError.Error()
}

func TranslateRepoError(err error) error {
	if err == nil {
		return nil
	}

	var ex *exception.Exception
	if errors.As(err, &ex) {
		return ex
	}

	if errors.Is(err, mysql.ErrCodeConflict) {
		detailedMsg := getDetailedRepoMessage(err, mysql.ErrCodeConflict)
		return exception.Wrap(err, exception.TypeConflict, exception.CodeConflict, detailedMsg)
	}

	if errors.Is(err, mysql.ErrDuplicateEntry) {
		detailedMsg := getDetailedRepoMessage(err, mysql.ErrDuplicateEntry)
		err := exception.New(exception.TypeConflict, exception.CodeConflict, detailedMsg)
		ex, ok := exception.GetException(err)

		if !ok {
			return errors.Wrap(err, "internal inconsistency: failed to get exception from new exception")
		}

		if ex.Errors == nil {
			ex.Errors = make(exception.FieldErrors)
		}

		duplicateKeyRegex := regexp.MustCompile(`for field '(.+?)'`)

		matches := duplicateKeyRegex.FindStringSubmatch(detailedMsg)
		if len(matches) > 1 {
			rawKeyName := matches[1]
			ex.Errors[rawKeyName] = append(ex.Errors[rawKeyName], detailedMsg)
		}

		return ex
	}

	if errors.Is(err, mysql.ErrForeignKey) {
		detailedMsg := getDetailedRepoMessage(err, mysql.ErrForeignKey)
		return exception.Wrap(err, exception.TypeValidationError, exception.CodeDBConstraintViolation, "Related data constraint violation: "+detailedMsg)
	}

	if errors.Is(err, mysql.ErrDataTooLong) {
		detailedMsg := getDetailedRepoMessage(err, mysql.ErrDataTooLong)
		return exception.Wrap(err, exception.TypeValidationError, exception.CodeValidationFailed, detailedMsg)
	}

	if errors.Is(err, mysql.ErrNotNull) {
		detailedMsg := getDetailedRepoMessage(err, mysql.ErrNotNull)
		return exception.Wrap(err, exception.TypeValidationError, exception.CodeValidationFailed, detailedMsg)
	}

	if errors.Is(err, mysql.ErrNotFound) {
		return exception.Wrap(err, exception.TypeNotFound, exception.CodeNotFound, "Data not found")
	}

	if errors.Is(err, mysql.ErrDataNull) {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Input data cannot be null")
	}

	if errors.Is(err, mysql.ErrIDNull) {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Identifier (ID) cannot be null")
	}

	if errors.Is(err, context.Canceled) {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Request canceled by client")
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return exception.Wrap(err, exception.TypeTimeout, exception.CodeTimeout, "Operation timed out")
	}

	return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "An internal server error occurred")
}
