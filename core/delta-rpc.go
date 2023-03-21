package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DeltaAPI struct {
	NodeUUID         string
	url              string
	ServiceAuthToken string
}

func NewDeltaAPI(url string, authToken string) (*DeltaAPI, error) {

	hcError := healthCheck(url)
	if hcError != nil {
		return nil, hcError
	}

	dapi := &DeltaAPI{
		url:              url,
		ServiceAuthToken: authToken,
	}

	err := dapi.populateNodeUuid()
	if err != nil {
		return nil, err
	}

	return dapi, nil
}

// Verify that Delta API is reachable
func healthCheck(baseUrl string) error {
	_, err := http.Get(baseUrl + "/api/v1/node/info")

	return err
}

// Retrieves delta node UUID and sets it on the DeltaAPI struct
func (d *DeltaAPI) populateNodeUuid() error {
	body, closer, err := d.getRequest("/open/node/uuids", "")
	defer closer()

	if err != nil {
		return fmt.Errorf("could not get node uuids: %s", err)
	}

	result, err := UnmarshalNodeUUIDsResponse(body)
	if err != nil {
		return fmt.Errorf("could not unmarshal node uuids response %s : %s", err, string(body))
	}

	d.NodeUUID = result[0].InstanceUUID
	return nil
}

// Register a wallet with Delta based on private key & type (i.e, from a private key file)
func (d *DeltaAPI) AddWalletByPrivateKey(wallet RegisterWalletRequest, authString string) (*RegisterWalletResponse, error) {
	w, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from wallet json: %s", err)
	}

	body, closer, err := d.postRequest("/admin/wallet/register", w, authString)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalRegisterWalletResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal add wallet response %s : %s", err, string(body))
	}

	return &result, nil
}

// Register a wallet with Delta based on hex key (i.e, from lotus wallet export)
func (d *DeltaAPI) AddWalletByHexKey(wallet RegisterWalletHexRequest, authString string) (*RegisterWalletResponse, error) {
	w, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from wallet json: %s", err)
	}

	body, closer, err := d.postRequest("/admin/wallet/register-hex", w, authString)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalRegisterWalletResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal add wallet response %s : %s", err, string(body))
	}

	return &result, nil
}

// Queries delta for wallet balance information
func (d *DeltaAPI) GetWalletBalance(walletAdr string, authString string) (*GetWalletBalanceResponse, error) {
	body, closer, err := d.getRequest("/admin/wallet/balance/"+walletAdr, authString)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalGetWalletBalanceResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal wallet balance response %s : %s", err, string(body))
	}

	return &result, nil
}

// Requests offline deals to be made from Delta
func (d *DeltaAPI) MakeOfflineDeals(deals OfflineDealRequest, authString string) (*OfflineDealResponse, error) {
	ds, err := json.Marshal(deals)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deals json: %s", err)
	}

	log.Debugf("delta deals request: %s", string(ds))

	body, closer, err := d.postRequest("/api/v1/deal/piece-commitments", ds, authString)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalOfflineDealResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal offline deal response %s : %s", err, string(body))
	}

	return &result, nil
}

func (d *DeltaAPI) GetDealStatus(deltaIds []int64) (*DealStatsResponse, error) {
	dids, err := json.Marshal(deltaIds)
	if err != nil {
		return nil, fmt.Errorf("could not marshal from deal ids json: %s", err)
	}

	body, closer, err := d.postRequest("/api/v1/stats/contents", dids, d.ServiceAuthToken)
	if err != nil {
		return nil, err
	}
	defer closer()

	result, err := UnmarshalDealStatsResponse(body)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal deal stats response %s : %s", err, string(body))
	}

	return &result, nil

}

func (d *DeltaAPI) postRequest(url string, raw []byte, authKey string) ([]byte, func() error, error) {
	if authKey == "" {
		return nil, nil, fmt.Errorf("auth token must be provided")
	}

	req, err := http.NewRequest("POST", d.url+url, bytes.NewBuffer(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not make http request %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("error in delta call %d : %s", resp.StatusCode, body)
	}

	if err != nil {
		return nil, nil, err
	}

	return body, resp.Body.Close, nil
}

func (d *DeltaAPI) getRequest(url string, authKey string) ([]byte, func() error, error) {
	if authKey == "" {
		return nil, nil, fmt.Errorf("auth token must be provided")
	}

	req, err := http.NewRequest("GET", d.url+url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct http request %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not make http request %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("error in delta call %d : %s", resp.StatusCode, body)
	}

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
	DeltaContentID       uint64          `json:"content_id,omitempty"` // TODO: rename to delta_id
	Cid                  string          `json:"cid"`
	Wallet               Wallet          `json:"wallet"`
	Miner                string          `json:"miner"` //TODO: rename to provider
	PieceCommitment      PieceCommitment `json:"piece_commitment"`
	ConnectionMode       string          `json:"connection_mode"`
	Size                 uint64          `json:"size"`
	RemoveUnsealedCopies bool            `json:"remove_unsealed_copies"`
	SkipIpniAnnounce     bool            `json:"skip_ipni_announce"`
	DurationInDays       uint64          `json:"duration_in_days,omitempty"`
	StartEpochAtDays     uint64          `json:"start_epoch_at_days,omitempty"`
}

type PieceCommitment struct {
	PieceCid        string `json:"piece_cid"`
	PaddedPieceSize uint64 `json:"padded_piece_size"`
}

type RegisterWalletRequest struct {
	Type       string `json:"key_type"`
	PrivateKey string `json:"private_key"`
}

type RegisterWalletHexRequest struct {
	HexKey string `json:"hex_key"`
}

type RegisterWalletResponse struct {
	Message    string `json:"message"`
	WalletAddr string `json:"wallet_addr"`
	WalletUuid string `json:"wallet_uuid"`
}

func UnmarshalRegisterWalletResponse(data []byte) (RegisterWalletResponse, error) {
	var r RegisterWalletResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RegisterWalletResponse) Marshal() ([]byte, error) {
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

func UnmarshalGetWalletBalanceResponse(data []byte) (GetWalletBalanceResponse, error) {
	var r GetWalletBalanceResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GetWalletBalanceResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type GetWalletBalanceResponse struct {
	Balance Balance `json:"balance"`
	Message string  `json:"message"`
}

type Balance struct {
	Account               string `json:"account"`
	Balance               uint64 `json:"balance"`
	MarketAvailable       uint64 `json:"market_available"`
	MarketEscrow          uint64 `json:"market_escrow"`
	MarketLocked          uint64 `json:"market_locked"`
	VerifiedClientBalance uint64 `json:"verified_client_balance"`
	WalletBalance         uint64 `json:"wallet_balance"`
}

func UnmarshalNodeUUIDsResponse(data []byte) (NodeUUIDsResponse, error) {
	var r NodeUUIDsResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

type NodeUUIDsResponse = []NodeUUID

type NodeUUID struct {
	Id           string `json:"id"`
	InstanceUUID string `json:"instance_uuid"`
	CreatedAt    string `json:"created_at"`
}
