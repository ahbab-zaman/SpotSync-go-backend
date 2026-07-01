package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RoleMiddleware(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok || role != requiredRole {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"message": "Forbidden: insufficient permissions",
				})
			}
			return next(c)
		}
	}
}
