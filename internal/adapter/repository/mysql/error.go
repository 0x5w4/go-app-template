package mysqlrepository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrCodeConflict   = errors.New("client code conflict")
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrForeignKey     = errors.New("foreign key constraint violation")
	ErrDataTooLong    = errors.New("data too long")
	ErrDataInvalid    = errors.New("data is invalid")
	ErrDataNull       = errors.New("data is null")
	ErrIDNull         = errors.New("id is null")
	ErrNotNull        = errors.New("null constraint violation")
	ErrNotFound       = errors.New("no rows in result set")
)

var duplicateKeyRegex = regexp.MustCompile(`for key '(.+?)'`)

func transformDBIdentifier(tableName, identifier string) string {
	processedName := strings.ToLower(identifier)
	normalizedTableName := strings.ToLower(tableName)

	prefixes := []string{"uq_", "uk_", "idx_", "fk_"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(processedName, prefix) {
			processedName = strings.TrimPrefix(processedName, prefix)
			break
		}
	}

	if normalizedTableName != "" {
		tableNameWithUnderscore := normalizedTableName + "_"
		processedName = strings.TrimPrefix(processedName, tableNameWithUnderscore)
	}

	suffixesToTrim := []string{"_active"}
	for _, suffix := range suffixesToTrim {
		if strings.HasSuffix(processedName, suffix) {
			processedName = strings.TrimSuffix(processedName, suffix)
			break
		}
	}

	if tableName == "clients" && strings.HasPrefix(processedName, "company_") {
		// Special case for client code to avoid confusion with 'company' field
		processedName = strings.TrimPrefix(processedName, "company_")
	}

	return processedName
}

func handleDBError(err error, tableName, operationDesc string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(ErrNotFound, operationDesc)
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		var baseErrorToWrap error

		var specificMessage string

		switch mysqlErr.Number {
		case 1062:
			baseErrorToWrap = ErrDuplicateEntry

			matches := duplicateKeyRegex.FindStringSubmatch(mysqlErr.Message)
			if len(matches) > 1 {
				rawKeyName := matches[1]
				fieldName := transformDBIdentifier(tableName, rawKeyName)
				specificMessage = fmt.Sprintf("Duplicate entry for field '%s'", fieldName)
			}
		case 1451, 1452, 1216, 1217:
			wrappedForeignKeyErr := errors.Wrap(ErrForeignKey, mysqlErr.Message)
			return errors.Wrap(wrappedForeignKeyErr, operationDesc)
		case 1048:
			baseErrorToWrap = ErrNotNull

			parts := strings.SplitN(mysqlErr.Message, "'", 3)
			if len(parts) >= 2 {
				columnName := parts[1]
				specificMessage = fmt.Sprintf("Field '%s' cannot be null", columnName)
			}
		case 1406:
			baseErrorToWrap = ErrDataTooLong

			parts := strings.SplitN(mysqlErr.Message, "'", 3)
			if len(parts) >= 2 {
				columnName := parts[1]
				specificMessage = fmt.Sprintf("Data too long for field '%s'", columnName)
			}
		case 1064, 1054:
			return errors.Wrap(err, operationDesc+": query error")
		default:
			return errors.Wrap(err, operationDesc+": mysql error occurred")
		}

		if baseErrorToWrap != nil {
			detailedError := baseErrorToWrap
			if specificMessage != "" {
				detailedError = errors.Wrap(baseErrorToWrap, specificMessage)
			}

			detailedError = errors.Wrap(detailedError, mysqlErr.Message)

			return errors.Wrap(detailedError, operationDesc)
		}

		return errors.Wrap(err, operationDesc+": unhandled MySQL error")
	}

	return errors.Wrap(err, operationDesc)
}
