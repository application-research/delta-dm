package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type PtolemyAPI struct {
	url string
}

func NewPtolemyAPI(url string) (*PtolemyAPI, error) {
	ni, hcError := healthCheck(url)
	if hcError != nil {
		return nil, hcError
	}

	dapi := &DeltaAPI{
		url:              url,
		ServiceAuthToken: authToken,
		DeltaDeploymentInfo: DeploymentInfo{
			Commit:  ni.Commit,
			Version: ni.Version,
		},
	}

	err := dapi.populateNodeUuid()
	if err != nil {
		return nil, err
	}

	return dapi, nil
}

// Verify that Delta API is reachable
func healthCheck(baseUrl string) (*NodeInfoResponse, error) {
	resp, err := http.Get(baseUrl + "/open/node/info")
	if err != nil {
		return nil, fmt.Errorf("could not reach delta api: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %s", err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("error in delta call %d : %s", resp.StatusCode, body)
	}

	defer resp.Body.Close()

	result, err := UnmarshalNodeInfoResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal node info response %s : %s", err, string(body))
	}

	return &result, nil
}
