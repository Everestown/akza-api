package httputil

import (
	"strconv"

	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// ParseID extracts an int64 ID from the "id" path parameter.
// On error it writes a 422 response and returns false.
func ParseID(c *gin.Context) (int64, bool) {
	return ParseParam(c, "id")
}

// ParseParam extracts an int64 from any named path parameter.
func ParseParam(c *gin.Context, param string) (int64, bool) {
	v := c.Param(param)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		middleware.Err(c, apperror.Validation("invalid "+param+": must be integer"))
		return 0, false
	}
	return id, true
}
