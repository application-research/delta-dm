package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DeltaAPI struct {
	url       string
	authToken string
}

func NewDeltaAPI(url string, authToken string) (*DeltaAPI, error) {

	hcError := healthCheck(url)
	if hcError != nil {
		return nil, hcError
	}

	return &DeltaAPI{
		url:       url,
		authToken: authToken,
	}, nil
}

// Verify that Delta API is reachable
func healthCheck(baseUrl string) error {
	_, err := http.Get(baseUrl + "/api/v1/node/info")

	return err
}

type DealRequest struct {
}

// Requests offline deals to be made from Delta
func (d *DeltaAPI) MakeOfflineDeals(deals OfflineDealRequest) (*OfflineDealResponse, error) {
	fmt.Printf("%+v\n", deals)
	ds, err := json.Marshal(deals)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deals json: %s", err)
	}

	req, err := http.NewRequest("POST", d.url+"/api/v1/deal/commitment-pieces", bytes.NewBuffer(ds))
	if err != nil {
		return nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make http request %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read from response body %s", err)
	}

	fmt.Printf("%+v\n", string(body))

	result, err := UnmarshalOfflineDealResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal offline deal response %s", err)
	}

	return &result, nil
}

type OfflineDealRequest []Deal

func UnmarshalOfflineDealRequest(data []byte) (OfflineDealRequest, error) {
	var r OfflineDealRequest
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *OfflineDealRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type OfflineDealResponse []OfflineDealResponseElement

func UnmarshalOfflineDealResponse(data []byte) (OfflineDealResponse, error) {
	var r OfflineDealResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *OfflineDealResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type OfflineDealResponseElement struct {
	Status            string `json:"status"`
	Message           string `json:"message"`
	ContentID         int64  `json:"content_id"`
	PieceCommitmentID int64  `json:"piece_commitment_id"`
	Meta              Deal   `json:"meta"`
}

type Deal struct { // AKA meta
	DeltaContentID int64  `json:"content_id,omitempty"`
	Cid            string `json:"cid"`
	Wallet         Wallet `json:"wallet"`
	Miner          string `json:"miner"` //TODO: rename to provider
	Commp          Commp  `json:"commp"`
	ConnectionMode string `json:"connection_mode"`
	Size           int64  `json:"size"`
}

type Commp struct {
	Piece             string `json:"piece"`
	PaddedPieceSize   int64  `json:"padded_piece_size"`
	UnpaddedPieceSize int64  `json:"unpadded_piece_size,omitempty"`
}

type Wallet struct {
}
