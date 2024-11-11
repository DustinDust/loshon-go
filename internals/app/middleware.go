package app

import (
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/labstack/echo/v4"
)

func (app App) ClerkAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionToken := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		}
		usr, err := user.Get(c.Request().Context(), claims.Subject)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		}
		c.Set("user", usr)
		return next(c)
	}
}

func (app App) OptionalClerkAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionToken := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			return next(c)
		}
		usr, err := user.Get(c.Request().Context(), claims.Subject)
		if err != nil {
			return next(c)
		}
		c.Set("user", usr)
		return next(c)
	}
}
