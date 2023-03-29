package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

var CLIConnectFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "ddm-api-info",
		Usage:   "DDM API connection info",
		EnvVars: []string{"DDM_API_INFO"},
		Hidden:  true,
	},
	&cli.StringFlag{
		Name:    "delta-auth",
		Usage:   "delta auth token",
		EnvVars: []string{"DELTA_AUTH"},
		Hidden:  true,
	},
}

type CmdProcessor struct {
	ddmUrl     string
	ddmAuthKey string
}

func NewCmdProcessor(c *cli.Context) (*CmdProcessor, error) {
	ddmUrl := getFlagOrEnvVar(c, "ddm-api-info", "DDM_API_INFO", "http://localhost:1314")
	ddmAuthKey := getFlagOrEnvVar(c, "delta-auth", "DELTA_AUTH", "")

	if ddmAuthKey == "" {
		return nil, fmt.Errorf("DELTA_AUTH env variable or --delta-auth flag is required")
	}

	err := healthCheck(ddmUrl, ddmAuthKey)

	if err != nil {
		return nil, fmt.Errorf("unable to communicate with ddm daemon: %s", err)
	}

	return &CmdProcessor{
		ddmUrl:     ddmUrl,
		ddmAuthKey: ddmAuthKey,
	}, nil
}

// If the flag is set, use it. If not, check the environment variable. If that's not set, use the default value
func getFlagOrEnvVar(c *cli.Context, flagName, envVarName, defaultValue string) string {
	value := c.String(flagName)
	if value == "" {
		value = os.Getenv(envVarName)
		if value == "" {
			value = defaultValue
		}
	}
	return value
}

// Verify that DDM API is reachable
func healthCheck(ddmUrl string, ddmAuthKey string) error {
	req, err := http.NewRequest(http.MethodGet, ddmUrl+"/api/v1/health", nil)
	if err != nil {
		return fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ddmAuthKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make http request %s", err)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		return fmt.Errorf(string(body))
	}

	return err
}

func (c *CmdProcessor) MakeRequest(method string, url string, raw []byte) ([]byte, func() error, error) {
	req, err := http.NewRequest(method, c.ddmUrl+url, bytes.NewBuffer(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.ddmAuthKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not make http request %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return body, resp.Body.Close, nil
}
