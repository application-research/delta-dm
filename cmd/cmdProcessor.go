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
	ddmUrl := c.String("ddm-api-info")
	if ddmUrl == "" {
		ddmUrl = os.Getenv("DDM_API_INFO")
		if ddmUrl == "" {
			ddmUrl = "http://localhost:1314"
		}
	}

	ddmAuthKey := c.String("delta-auth")
	if ddmAuthKey == "" {
		ddmAuthKey = os.Getenv("DELTA_AUTH")
		fmt.Printf("from cli param: %s", ddmAuthKey)
		if ddmAuthKey == "" {
			return nil, fmt.Errorf("DELTA_AUTH env variable or --delta-auth flag is required")
		}
	}

	err := healthCheck(ddmUrl, ddmAuthKey)

	if err != nil {
		return nil, err
	}

	return &CmdProcessor{
		ddmUrl:     ddmUrl,
		ddmAuthKey: ddmAuthKey,
	}, nil
}

// Verify that DDM API is reachable
func healthCheck(ddmUrl string, ddmAuthKey string) error {
	req, err := http.NewRequest("GET", ddmUrl+"/api/v1/health", nil)
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

func (c *CmdProcessor) ddmRequest(method string, url string, raw []byte) ([]byte, func() error, error) {
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
