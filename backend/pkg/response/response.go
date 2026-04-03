package response

import "github.com/gin-gonic/gin"

// Success sends a JSON response with the given status code and data.
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"data": data})
}

// Error sends a JSON error response with the given status code and message.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
