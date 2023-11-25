package inertia

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/theartefak/inertia-echo/utils"
)

// MiddlewareConfig holds configuration options for the Inertia middleware.
type MiddlewareConfig struct {
	Inertia *Inertia               // Reference to the Inertia instance
	Skipper middleware.Skipper    // Skipper function to conditionally skip the middleware
}

// Middleware creates a default Inertia middleware for the given Echo instance.
func Middleware(echo *echo.Echo) echo.MiddlewareFunc {
	return MiddlewareWithConfig(MiddlewareConfig{
		Inertia: NewInertia(echo),
	})
}

// MiddlewareWithConfig creates an Inertia middleware with a specific configuration.
func MiddlewareWithConfig(config MiddlewareConfig) echo.MiddlewareFunc {
	// Ensure that an Inertia reference is provided in the configuration.
	if config.Inertia == nil {
		log.Fatal("[Inertia] Please provide an Inertia reference with your config!")
	}

	// Return the actual middleware function.
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// The middleware function to be executed.
		return func(c echo.Context) error {
			// Skip, if configured and true
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			// Run Inertia post
			if err := next(c); err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			// Check if the request has the Inertia header
			if req.Header.Get(HeaderPrefix) == "" {
				return nil
			}

			// Adjust status code for certain HTTP methods and response status
			if exists, _ := utils.InArray(req.Method, []string{"PUT", "PATCH", "DELETE"}); exists && res.Status == http.StatusFound {
				res.Status = http.StatusSeeOther
			}

			return nil
		}
	}
}
