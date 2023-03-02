package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DeltaAPI struct {
	url              string
	serviceAuthToken string
}

func NewDeltaAPI(url string, authToken string) (*DeltaAPI, error) {

	hcError := healthCheck(url)
	if hcError != nil {
		return nil, hcError
	}

	return &DeltaAPI{
		url:              url,
		serviceAuthToken: authToken,
	}, nil
}

// Verify that Delta API is reachable
func healthCheck(baseUrl string) error {
	_, err := http.Get(baseUrl + "/api/v1/node/info")

	return err
}

func (d *DeltaAPI) AddWallet(wallet AddWalletRequest, authString string) (*AddWalletResponse, error) {
	w, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from wallet json: %s", err)
	}

	body, closer, err := d.postRequest("/admin/wallet/register", w, authString)
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
func (d *DeltaAPI) MakeOfflineDeals(deals OfflineDealRequest, authString string) (*OfflineDealResponse, error) {
	ds, err := json.Marshal(deals)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deals json: %s", err)
	}

	body, closer, err := d.postRequest("/api/v1/deal/piece-commitments", ds, authString)
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

func (d *DeltaAPI) GetDealStatus(deltaIds []string) (*DealStatsResponse, error) {
	dids, err := json.Marshal(deltaIds)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deal ids json: %s", err)
	}

	body, closer, err := d.postRequest("/api/v1/stats/contents", dids, d.serviceAuthToken)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalDealStatsResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal deal stats response %s", err)
	}

	return &result, nil

}

func (d *DeltaAPI) postRequest(url string, raw []byte, authString string) ([]byte, func() error, error) {
	if authString == "" {
		return nil, nil, fmt.Errorf("auth token must be provided")
	}

	req, err := http.NewRequest("POST", d.url+url, bytes.NewBuffer(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", authString)
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
	Duration             int64           `json:"duration,omitempty"`
	StartEpoch           int64           `json:"start_epoch,omitempty"`
}

type PieceCommitment struct {
	PieceCid        string `json:"piece_cid"`
	PaddedPieceSize int64  `json:"padded_piece_size"`
}

type AddWalletRequest struct {
	Type       string `json:"key_type"`
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

type DealStatsResponse []DealStatsResponseElement

func UnmarshalDealStatsResponse(data []byte) (DealStatsResponse, error) {
	var r DealStatsResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DealStatsResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type DealStatsResponseElement struct {
	Content          DealStats_Content           `json:"content"`
	DealProposals    []DealStats_DealProposal    `json:"deal_proposals"`
	Deals            []DealStats_Deal            `json:"deals"`
	PieceCommitments []DealStats_PieceCommitment `json:"piece_commitments"`
}

type DealStats_Content struct {
	ID                int64  `json:"ID"`
	Name              string `json:"name"`
	Size              int64  `json:"size"`
	Cid               string `json:"cid"`
	RequestingAPIKey  string `json:"requesting_api_key"`
	PieceCommitmentID int64  `json:"piece_commitment_id"`
	Status            string `json:"status"`
	ConnectionMode    string `json:"connection_mode"`
	LastMessage       string `json:"last_message"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type DealStats_DealProposal struct {
	ID        int64  `json:"ID"`
	Content   int64  `json:"content"`
	Unsigned  string `json:"unsigned"`
	Signed    string `json:"signed"`
	Meta      string `json:"meta"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type DealStats_Deal struct {
	ID                  int64  `json:"ID"`
	Content             int64  `json:"content"`
	PropCid             string `json:"propCid"`
	DealUUID            string `json:"dealUuid"`
	Miner               string `json:"miner"`
	DealID              int64  `json:"dealId"`
	Failed              bool   `json:"failed"`
	Verified            bool   `json:"verified"`
	Slashed             bool   `json:"slashed"`
	FailedAt            string `json:"failedAt"`
	DtChan              string `json:"dtChan"`
	TransferStarted     string `json:"transferStarted"`
	TransferFinished    string `json:"transferFinished"`
	OnChainAt           string `json:"onChainAt"`
	SealedAt            string `json:"sealedAt"`
	LastMessage         string `json:"lastMessage"`
	DealProtocolVersion string `json:"deal_protocol_version"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type DealStats_PieceCommitment struct {
	ID                 int64  `json:"ID"`
	Cid                string `json:"cid"`
	Piece              string `json:"piece"`
	Size               int64  `json:"size"`
	PaddedPieceSize    int64  `json:"padded_piece_size"`
	UnnpaddedPieceSize int64  `json:"unnpadded_piece_size"`
	Status             string `json:"status"`
	LastMessage        string `json:"last_message"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
