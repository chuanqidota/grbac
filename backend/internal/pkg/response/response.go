package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
)

// Response is the standard JSON envelope returned by all API endpoints.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData wraps paginated results.
type PageData struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

// OK sends a success response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.Success.Code,
		Message: errors.Success.Message,
		Data:    data,
	})
}

// OKMessage sends a success response without data payload.
func OKMessage(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.Success.Code,
		Message: errors.Success.Message,
		Data:    nil,
	})
}

// OKPage sends a paginated success response.
func OKPage(c *gin.Context, total int64, list interface{}) {
	OK(c, PageData{
		Total: total,
		List:  list,
	})
}

// Fail sends an error response defined by an AppError.
func Fail(c *gin.Context, appErr *errors.AppError) {
	c.JSON(http.StatusOK, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// FailWithMessage sends an error response with a custom message override.
func FailWithMessage(c *gin.Context, appErr *errors.AppError, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    appErr.Code,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized sends a 401 HTTP response.
func Unauthorized(c *gin.Context, appErr *errors.AppError) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// Forbidden sends a 403 HTTP response.
func Forbidden(c *gin.Context, appErr *errors.AppError) {
	c.JSON(http.StatusForbidden, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// InternalError sends the generic internal error response.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errors.ErrInternal.Code,
		Message: errors.ErrInternal.Message,
		Data:    nil,
	})
}

// RespondError sends the appropriate error response based on the error type.
func RespondError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		Fail(c, appErr)
		return
	}
	InternalError(c)
}
