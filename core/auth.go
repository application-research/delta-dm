package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

var AUTH_KEY = "AUTH_KEY"

type AuthServer struct {
	authServerUrl string
}

func NewAuthServer(authServerUrl string) *AuthServer {
	return &AuthServer{authServerUrl: authServerUrl}
}

func (as *AuthServer) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authKey, err := extractAuthKey(c.Request().Header.Get("Authorization"))

		if err != nil {
			return c.JSON(401, err.Error())
		}

		res, err := as.checkAuthToken(*authKey)
		if err != nil {
			return c.JSON(401, err.Error())
		}

		if !res.Validated {
			return c.JSON(401, res.Details)
		}

		c.Set(AUTH_KEY, *authKey)

		return next(c)

	}
}

// Check that an auth string is populated in header and formatted correctly, then return it
//    `hint: pass in the value of c.Request().Header.Get("Authorization")`
func extractAuthKey(authorizationString string) (*string, error) {
	if authorizationString == "" {
		return nil, fmt.Errorf("missing auth header")
	}

	authParts := strings.Split(authorizationString, " ")
	if len(authParts) != 2 {
		return nil, fmt.Errorf("malformed auth header - must be of the form BEARER <token>")
	}
	if authParts[0] != "Bearer" {
		return nil, fmt.Errorf("malformed auth header - must have `Bearer` prefix")
	}

	estuaryAuthKey, _ := regexp.MatchString("^(EST).*(ARY)$", authParts[1])
	deltaAuthKey, _ := regexp.MatchString("^(DEL).*(TA)$", authParts[1])

	if !estuaryAuthKey && !deltaAuthKey {
		return nil, fmt.Errorf("malformed auth header - must be DELTA or ESTUARY key")
	}
	return &authParts[1], nil
}

// Makes a request to the auth server to check if a token is valid
func (as *AuthServer) checkAuthToken(token string) (*AuthResult, error) {
	rqBody := strings.NewReader(fmt.Sprintf(`{"token": "%s"}`, token))
	resp, err := http.Post(as.authServerUrl+"/check-api-key", "application/json", rqBody)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error making auth call %d : %s", resp.StatusCode, body)
	}
	if err != nil {
		return nil, fmt.Errorf("error reading auth response: %s", err)
	}

	var ar AuthResponse

	err = json.Unmarshal(body, &ar)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling auth response: %s", err)
	}

	return &ar.Result, nil
}

type AuthResponse struct {
	Result AuthResult `json:"result"`
}

type AuthResult struct {
	Validated bool   `json:"validated"`
	Details   string `json:"details"`
}
