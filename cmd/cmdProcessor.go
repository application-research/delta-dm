package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CmdProcessor struct {
	ddmUrl     string
	ddmAuthKey string
}

func NewCmdProcessor(ddmUrl string, ddmAuthKey string) (*CmdProcessor, error) {
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
