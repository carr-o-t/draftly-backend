package constants

type ErrorCode string
type SuccessCode string

const (
	// success
	SuccessLogin    SuccessCode = "LOGIN_SUCCESS"
	SuccessCreated  SuccessCode = "USER_CREATED"
	
	// errors
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrRateLimit      ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrHashPassword   ErrorCode = "PASSWORD_HASH_FAILED"
	ErrCreateUser     ErrorCode = "USER_CREATION_FAILED"
	ErrInvalidCreds   ErrorCode = "INVALID_CREDENTIALS"
	ErrDBError        ErrorCode = "DATABASE_ERROR"
	ErrTokenGen       ErrorCode = "TOKEN_GENERATION_FAILED"
)
