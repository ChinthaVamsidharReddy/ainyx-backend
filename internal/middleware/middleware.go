package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

// RequestID injects a unique request-id into every response header.
// If the client already sent one it is forwarded as-is, otherwise a new UUID is minted.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(requestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(requestIDHeader, id)
		// Store in locals so other middleware / handlers can reference it.
		c.Locals("requestID", id)
		return c.Next()
	}
}

// Logger logs the method, path, status code, latency, and request-id for every request.
func Logger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Run the actual handler chain.
		err := c.Next()

		log.Info("request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", c.Locals("requestID").(string)),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
