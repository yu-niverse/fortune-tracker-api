package validator

import (
	"Fortune_Tracker_API/internal/response"
	"Fortune_Tracker_API/pkg/logger"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// chech param ULID is in valid format or not
func ValidateULIDParam(c *gin.Context) {
	// ULID should be 36 characters long and contains only lowercase letters, numbers and hyphens
	if matched, _ := regexp.MatchString("^[a-z0-9-]{36}$", c.Param("ulid")); !matched {
		logger.Warn("[LEDGER] ULID is not in valid format")
		r := response.New()
		r.Message = "ULID is not in valid format"
		c.JSON(http.StatusBadRequest, r)
		c.Abort()
		return
	}

	// Check if ULID is valid -> continue
	c.Next()
}
