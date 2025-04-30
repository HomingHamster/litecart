package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	jwtMiddleware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/shurco/litecart/internal/models"
	"github.com/shurco/litecart/internal/queries"
	"github.com/shurco/litecart/pkg/webutil"
)

// JWTProtected is ...
func JWTProtected() func(*fiber.Ctx) error {
	config := jwtMiddleware.Config{
		KeyFunc:      customKeyFunc(),
		ContextKey:   "jwt",
		ErrorHandler: jwtError,
		TokenLookup:  "cookie:token",
	}

	return jwtMiddleware.New(config)
}

func jwtError(c *fiber.Ctx, err error) error {
	path := strings.Split(c.Path(), "/")[1]
	if path == "api" {
		if err.Error() == "Missing or malformed token" {
			return webutil.Response(c, fiber.StatusBadRequest, "Bad request", err.Error())
		}
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", err.Error())
	}

	return c.Redirect("/_/signin")
}

func customKeyFunc() jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		// Set a timeout of 5 secs to prevent indefinite blocking
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := queries.DB()
		settingJWT, err := queries.GetSettingByGroup[models.JWT](ctx, db)
		// Handles database errors when retrieving the JWT secret
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				// Database time out
				return nil, fmt.Errorf("database took too long to respond")
			}
			// Database error
			return nil, fmt.Errorf("database error: %w", err)
		}
		//Add secret validation
		if settingJWT.Secret == "" {
			return nil, fmt.Errorf("JWT secret is empty or not configured")
		}

		return []byte(settingJWT.Secret), nil
	}
}
