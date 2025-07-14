package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"

	"gateway/internal/errs"
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

		errsList := c.Errors
		if len(errsList) == 0 {
			return
		}

		err := errsList[0].Err

		code := "internal_error"
		statusCode := http.StatusInternalServerError
		description := "Internal server error"

		var httpErr errs.HTTPError
		if errors.As(err, &httpErr) {
			code = http.StatusText(httpErr.StatusCode)
			description = httpErr.Message
			statusCode = httpErr.StatusCode
		} else {
			if st, ok := status.FromError(err); ok {
				statusCode = runtime.HTTPStatusFromCode(st.Code())
				code = http.StatusText(statusCode)
				description = st.Message()
			}
		}

		resp := HttpResponse{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Path:        c.Request.URL.Path,
			Method:      c.Request.Method,
			Code:        code,
			Description: description,
		}

		c.AbortWithStatusJSON(statusCode, resp)
	}
}
