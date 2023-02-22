package core

import (
	"bytes"
	"encoding/json"
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
	ds, err := json.Marshal(deals)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", d.url+"/api/v1/deal/commitment-pieces", bytes.NewBuffer(ds))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result, err := UnmarshalOfflineDealResponse(body)
	if err != nil {
		return nil, err
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

type Deal struct {
	Cid            string `json:"cid"`
	Wallet         string `json:"wallet"`
	Commp          Commp  `json:"commp"`
	ConnectionMode string `json:"connection_mode"`
	Size           int64  `json:"size"`
}

type Commp struct {
	Piece             string `json:"piece"`
	PaddedPieceSize   int64  `json:"padded_piece_size"`
	UnpaddedPieceSize int64  `json:"unpadded_piece_size,omitempty"`
}
