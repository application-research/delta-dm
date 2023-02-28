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

func (d *DeltaAPI) AddWallet(wallet AddWalletRequest) (*AddWalletResponse, error) {
	w, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from wallet json: %s", err)
	}

	body, closer, err := d.postRequest("/admin/wallet/register", w)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalAddWalletResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal add wallet response %s", err)
	}

	return &result, nil
}

// Requests offline deals to be made from Delta
func (d *DeltaAPI) MakeOfflineDeals(deals OfflineDealRequest) (*OfflineDealResponse, error) {
	ds, err := json.Marshal(deals)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deals json: %s", err)
	}

	body, closer, err := d.postRequest("/api/v1/deal/piece-commitments", ds)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalOfflineDealResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal offline deal response %s", err)
	}

	return &result, nil
}

func (d *DeltaAPI) postRequest(url string, raw []byte) ([]byte, func() error, error) {
	req, err := http.NewRequest("POST", d.url+url, bytes.NewBuffer(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.authToken)
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
	RequestMeta       Deal   `json:"request_meta"`
}

type Deal struct { // AKA meta
	DeltaContentID       int64           `json:"content_id,omitempty"` // TODO: rename to delta_id
	Cid                  string          `json:"cid"`
	Wallet               Wallet          `json:"wallet"`
	Miner                string          `json:"miner"` //TODO: rename to provider
	PieceCommitment      PieceCommitment `json:"piece_commitment"`
	ConnectionMode       string          `json:"connection_mode"`
	Size                 int64           `json:"size"`
	RemoveUnsealedCopies bool            `json:"remove_unsealed_copies"`
	SkipIpniAnnounce     bool            `json:"skip_ipni_announce"`
}

type PieceCommitment struct {
	PieceCid        string `json:"piece_cid"`
	PaddedPieceSize int64  `json:"padded_piece_size"`
}

type AddWalletRequest struct {
	Type       string `json:"type"`
	PrivateKey string `json:"private_key"`
}

type AddWalletResponse struct {
	Message    string `json:"message"`
	WalletAddr string `json:"wallet_addr"`
	WalletUuid string `json:"wallet_uuid"`
}

func UnmarshalAddWalletResponse(data []byte) (AddWalletResponse, error) {
	var r AddWalletResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AddWalletResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
