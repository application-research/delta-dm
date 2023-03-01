package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CmdProcessor struct {
	ddmUrl    string
	ddmApiKey string
}

func NewCmdProcessor(ddmUrl string, ddmApiKey string) (*CmdProcessor, error) {
	err := healthCheck(ddmUrl, ddmApiKey)

	if err != nil {
		return nil, err
	}

	return &CmdProcessor{
		ddmUrl:    ddmUrl,
		ddmApiKey: ddmApiKey,
	}, nil
}

// Verify that DDM API is reachable
func healthCheck(ddmUrl string, ddmApikey string) error {
	req, err := http.NewRequest("GET", ddmUrl+"/api/v1/health", nil)
	if err != nil {
		return fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ddmApikey)

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

func (c *CmdProcessor) ddmPostRequest(url string, raw []byte) ([]byte, func() error, error) {
	req, err := http.NewRequest("POST", c.ddmUrl+url, bytes.NewBuffer(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.ddmApiKey)
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
