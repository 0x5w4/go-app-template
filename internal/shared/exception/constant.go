package exception

import (
	"github.com/cockroachdb/errors"
)

const (
	TypeInternalError        ErrorType = "Internal Error"
	TypeServiceUnavailable   ErrorType = "Service Unavailable"
	TypeTimeout              ErrorType = "Timeout"
	TypeBadRequest           ErrorType = "Bad Request"
	TypeValidationError      ErrorType = "Validation Error"
	TypeUnauthorized         ErrorType = "Unauthorized"
	TypeForbidden            ErrorType = "Forbidden"
	TypeTokenInvalid         ErrorType = "Token Invalid"
	TypeTokenExpired         ErrorType = "Token Expired"
	TypePermissionDenied     ErrorType = "Permission Denied"
	TypeNotFound             ErrorType = "Not Found"
	TypeMethodNotAllowed     ErrorType = "Method Not Allowed"
	TypeConflict             ErrorType = "Conflict"
	TypeUnsupportedMediaType ErrorType = "Unsupported Media Type"
	TypeRateLimitExceeded    ErrorType = "Rate Limit Exceeded"
	TypeQueryError           ErrorType = "Query Error"
	TypeConnectionError      ErrorType = "Connection Error"
	TypeAuthenticationError  ErrorType = "Authentication Error"
	TypeResourceError        ErrorType = "Resource Error"
	TypeConstraintError      ErrorType = "Constraint Error"
)

const (
	CodeInternalError         = "INTERNAL_ERROR"
	CodeValidationFailed      = "VALIDATION_FAILED"
	CodeNotFound              = "NOT_FOUND"
	CodeConflict              = "CONFLICT"
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeForbidden             = "FORBIDDEN"
	CodeBadRequest            = "BAD_REQUEST"
	CodeTimeout               = "TIMEOUT"
	CodeServiceUnavailable    = "SERVICE_UNAVAILABLE"
	CodeUserNotFound          = "USER_NOT_FOUND"
	CodeUserAlreadyExists     = "USER_ALREADY_EXISTS"
	CodeUserInvalidLogin      = "USER_INVALID_LOGIN"
	CodeResourceNotFound      = "RESOURCE_NOT_FOUND"
	CodeDuplicateResource     = "DUPLICATE_RESOURCE"
	CodeTokenInvalid          = "TOKEN_INVALID"
	CodeTokenExpired          = "TOKEN_EXPIRED"
	CodeTokenBlacklisted      = "TOKEN_BLACKLISTED"
	CodeAuthHeaderMissing     = "AUTH_HEADER_MISSING"
	CodeAuthHeaderInvalid     = "AUTH_HEADER_INVALID"
	CodeAuthUnsupported       = "AUTH_UNSUPPORTED"
	CodeDBConstraintViolation = "DB_CONSTRAINT_VIOLATION"
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
	ErrTimeout        = errors.New("operation timed out")
	ErrConnection     = errors.New("connection error")
	ErrTxFailed       = errors.New("transaction failed")
)

var (
	ErrAuthHeaderMissing    = New(TypePermissionDenied, CodeAuthHeaderMissing, "Authorization header not provided")
	ErrAuthHeaderInvalid    = New(TypePermissionDenied, CodeAuthHeaderInvalid, "Invalid authorization header format")
	ErrAuthUnsupported      = New(TypePermissionDenied, CodeAuthUnsupported, "Unsupported authorization type")
	ErrAuthTokenInvalid     = New(TypeBadRequest, CodeTokenInvalid, "Invalid or expired token")
	ErrAuthTokenBlacklisted = New(TypePermissionDenied, CodeTokenBlacklisted, "Token has been logged out")
)
