package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var AUTH_KEY = "AUTH_KEY"

type AuthServer struct {
	authServerUrl string
	db            *gorm.DB
}

func NewAuthServer(authServerUrl string, db *gorm.DB) *AuthServer {
	return &AuthServer{authServerUrl: authServerUrl, db: db}
}

func (as *AuthServer) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authKey, err := extractAuthKey(c.Request().Header.Get("Authorization"))

		if err != nil {
			return c.JSON(401, err.Error())
		}

		res, err := as.checkEstuaryAuthToken(*authKey)
		if err != nil {
			return c.JSON(401, err.Error())
		}

		if !res.Validated {
			return c.JSON(401, res.Details)
		}

		valid, err := as.checkLocalAuthToken(*authKey)
		if err != nil {
			return c.JSON(401, err.Error())
		} else if !valid {
			return c.JSON(401, "auth key is not registered in ddm")
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

	if !estuaryAuthKey {
		return nil, fmt.Errorf("malformed auth header - must be ESTUARY key")
	}
	return &authParts[1], nil
}

// Makes a request to the auth server to check if a token is valid
func (as *AuthServer) checkEstuaryAuthToken(token string) (*AuthResult, error) {
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

// Check the local DB to see if a token is valid
func (as *AuthServer) checkLocalAuthToken(token string) (bool, error) {
	var at Auth
	res := as.db.Model(&Auth{}).Where("auth_token = ?", token).First(&at)

	if res.Error != nil {
		return false, res.Error
	}

	if at.AuthToken != token {
		return false, nil
	}

	return true, nil
}

type AuthResponse struct {
	Result AuthResult `json:"result"`
}

type AuthResult struct {
	Validated bool   `json:"validated"`
	Details   string `json:"details"`
}
