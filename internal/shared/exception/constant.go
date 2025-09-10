package exception



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
	CodeAuthHeaderMissing     = "AUTH_HEADER_MISSING"
	CodeAuthHeaderInvalid     = "AUTH_HEADER_INVALID"
	CodeAuthUnsupported       = "AUTH_UNSUPPORTED"
	CodeDBConstraintViolation = "DB_CONSTRAINT_VIOLATION"
)