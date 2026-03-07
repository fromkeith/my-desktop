package middleware

// influenced by https://gist.github.com/SubCoder1/3a700149b2e7bb179a9123c6283030ff

import (
	"github.com/gin-gonic/gin"
)

// Mandatory Headers which should be set in the Response header for SSE to work.
func StreamHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
