package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

func RequestAuthHeaderCheck(c echo.Context) error {
	authorizationString := c.Request().Header.Get("Authorization")

	authErr := authStringCheck(authorizationString)
	return authErr
}

// Check that an auth string is populated and formatted correctly
//    `hint: pass in the value of c.Request().Header.Get("Authorization")`
func authStringCheck(authorizationString string) error {
	if authorizationString == "" {
		return fmt.Errorf("missing auth header")
	}

	authParts := strings.Split(authorizationString, " ")
	if len(authParts) != 2 {
		return fmt.Errorf("malformed auth header - must be of the form BEARER <token>")
	}
	if authParts[0] != "Bearer" {
		return fmt.Errorf("malformed auth header - must have `Bearer` prefix")
	}

	estuaryAuthKey, _ := regexp.MatchString("^(EST).*(ARY)$", authParts[1])
	deltaAuthKey, _ := regexp.MatchString("^(DEL).*(TA)$", authParts[1])

	if !estuaryAuthKey && !deltaAuthKey {
		return fmt.Errorf("malformed auth header - must be DELTA or ESTUARY key")
	}
	return nil
}
