package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Aneesh-382005/Orbital/internal/auth"
	"github.com/Aneesh-382005/Orbital/internal/metrics"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		labels := []string{c.Request.Method, c.FullPath(), status}
		metrics.HTTPRequestsTotal.WithLabelValues(labels...).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	}
}

const userIDKey = "userID"

func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := jwtService.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userID, ok := c.Get(userIDKey)
	if !ok {
		return ""
	}
	s, ok := userID.(string)
	if !ok {
		return ""
	}
	return s
}