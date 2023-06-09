# Provider Self-Service

This functionality allows a provider to request a deal to be sent to them using their token, at the `/api/v1/self-service/` endpoint. 

Currently, support is for a specific CID (PieceCID/CommP). 

After adding a provider to DDM, query `GET /api/v1/providers`, and get the `key` for the provider. This is their private authentication key, so it should be stored/transmitted safely. 

## Requesting a deal

### By CID
The provider can make a request as follows:

```bash
curl --request GET \
  --url 'http://your-delta-dm-address-here/api/v1/self-service/by-cid/bagaCID?start_epoch_delay=3' \
  --header 'X-DELTA-AUTH: b3cc8a99-155a-4fff-8974-999ec313e5cc'
```

Where
- `bagaCID` is the Piece CID to be replicated (example: `baga6ea4seaqd5nbcbhx5yzpoqtcdwkn5eawl2e63gui7jp5qpiwtil43z6eysdq`)
- `start_epoch_delay` is the number of epochs to wait before starting the deal (optional, default: 3)
- Header `X-DELTA-AUTH` is the provider's `key`, as described above


Calling this endpoint will cause DDM to issue a deal for that content to the provider.

If  successful, the following will be returned:

```
Status: 200 (OK)
{
	"cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhx",
	"content_location": "http://google.com/carfile.car"
}
```

If it fails, a 500 error will be returned, with the error message in the body. For example:

```
{
	"error": {
		"code": 500,
		"reason": "Internal Server Error",
		"details": "content 'baga6ea4seaqh26gjoj72ruwjhfshu76byab4tkt6kr53xicfnplu3rjiazxtski' is already replicated to provider 'f012345'"
	}
}
```

### By Dataset
The provider can request any CID to be dealt from a given dataset as follows:
```bash
curl --request GET \
  --url 'http://your-delta-dm-address-here/api/v1/self-service/by-dataset/dataset-name?start_epoch_delay=3' \
  --header 'X-DELTA-AUTH: b3cc8a99-155a-4fff-8974-999ec313e5cc'
```

**Reasons for failure may include:**
- Content being requested is already replicated to the provider
- Content being request has already reched its `replication_quota` for the dataset
- Key is invalid
- No wallet is associated with the dataset


A deal for Piece CID that has failed previously can be re-requested; it will re-attempt the deal.