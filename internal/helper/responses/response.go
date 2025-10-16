package responses

import "github.com/gin-gonic/gin"

// JSONError sends a structured error response
func JSONError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"status":  "error",
		"message": message,
	})
}

// JSONSuccess sends a structured success response
func JSONSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}
