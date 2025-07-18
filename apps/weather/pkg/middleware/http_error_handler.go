package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

type HttpResponse struct {
	Timestamp   string `json:"timestamp"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func HttpErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		errs := c.Errors
		if len(errs) == 0 {
			return
		}

		err := errs[0].Err

		var (
			code        = "internal_error"
			status      = http.StatusInternalServerError
			description = "Internal server error"
		)

		if apiErr, ok := pkgErrors.IsApiError(err); ok {
			code = apiErr.Base.Error()
			description = apiErr.Message

			if httpCode, found := ToHTTPStatus(apiErr.Base); found {
				status = httpCode
			}
		}
		resp := HttpResponse{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Code:        code,
			Description: description,
		}

		c.AbortWithStatusJSON(status, resp)
	}
}

func ToHTTPStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, internalErrors.ErrNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, internalErrors.ErrConflict):
		return http.StatusConflict, true
	case errors.Is(err, internalErrors.ErrInvalidInput):
		return http.StatusBadRequest, true
	case errors.Is(err, internalErrors.ErrServiceUnavailable):
		return http.StatusServiceUnavailable, true
	case errors.Is(err, internalErrors.ErrInternal):
		return http.StatusInternalServerError, true
	default:
		return http.StatusInternalServerError, false
	}
}
