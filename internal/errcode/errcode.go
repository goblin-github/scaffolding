package errcode

// AppError 业务错误。Code 给前端 switch-case，Reason 给日志检索和 i18n 翻译，
// Message 是默认（英文）的人类可读信息，会被 i18n 覆盖。
type AppError struct {
	Code    int
	Reason  string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// New 便于 service 层包装动态错误信息（底层默认 message 会被覆盖的场景）。
func New(code int, reason, message string) *AppError {
	return &AppError{Code: code, Reason: reason, Message: message}
}

// ── 通用错误 ──────────────────────────────────────────────

var (
	ErrInvalidParams = &AppError{Code: 400, Reason: "INVALID_PARAMS", Message: "Invalid request parameters"}
	ErrUnauthorized  = &AppError{Code: 401, Reason: "UNAUTHORIZED", Message: "Not logged in or session expired"}
	ErrForbidden     = &AppError{Code: 403, Reason: "FORBIDDEN", Message: "No permission"}
	ErrNotFound      = &AppError{Code: 404, Reason: "NOT_FOUND", Message: "Resource not found"}
	ErrInternal      = &AppError{Code: 500, Reason: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrDatabase      = &AppError{Code: 503, Reason: "SYSTEM_BUSY", Message: "System busy, please try again later"}
)

// ── 业务错误示例 ──────────────────────────────────────────

var (
	ErrTooManyTags = &AppError{Code: 5001, Reason: "TOO_MANY_TAGS", Message: "Too many tags"}
)
