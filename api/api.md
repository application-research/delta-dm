
# API Methods

All endpoints are prefixed with `/api/v1`.

For example, `http://localhost:1314/api/v1/datasets`

## /datasets

### POST /
- Add a dataset to be tracked 

#### Request Params
<nil>

#### Request Body
```jsonc
{
	"name": "delta-test",
	"replication_quota": 6,
	"deal_duration": 540,
	"unsealed": false,
	"indexed": true
}
```

#### Response
> 200: Success
> 500: Fail

### GET /
- Returns a list of all datasets

#### Request Params:
<nil>

#### Request Body
<nil> 

#### Response
> 200 : Success
```jsonc
[
	{
		"ID": 1,
		"name": "delta-test",
		"replication_quota": 6,
		"delay_start_epoch": 7,
		"deal_duration": 540,
		"wallet": {
			"address": "f1tuoahmuwfhnxpugqigxliu4muasggezw2efuczq",
			"dataset_name": "delta-test",
			"type": "secp256k1"
		},
		"unsealed": false,
		"indexed": true,
		"contents": null,
		"bytes_replicated": [
			198110211431, // Raw bytes (the content itself)
			377957122048 // Padded bytes (i.e, filecoin piece)
		],
		"bytes_total": [
			1801001922192, // Raw bytes (the content itself)
			3435973836800 // Padded bytes (i.e, filecoin piece)
		]
	},
	{
		"ID": 2,
		"name": "delta-test-2",
		"replication_quota": 6,
		"delay_start_epoch": 7,
		"deal_duration": 540,
		"wallet": {
			"address": "f1tuoahmuwfhnxpugqigxliu4muasggezw2eaaaa",
			"dataset_name": "delta-test-2",
			"type": "secp256k1"
		},
		"unsealed": false,
		"indexed": true,
		"contents": null,
		"bytes_replicated": [
			198110211431, // Raw bytes (the content itself)
			377957122048 // Padded bytes (i.e, filecoin piece)
		],
		"bytes_total": [
			1801001922192, // Raw bytes (the content itself)
			3435973836800 // Padded bytes (i.e, filecoin piece)
		]
	}
]

```

### POST /content/:dataset
- Add content (CAR files) to the dataset

#### Request Params
<nil>

#### Request Body
```jsonc
[
  {
    "payload_cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhy",
    "commp": "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmlq",
    "padded_size": 34359738368,
    "size": 18010019221
  },
  {
    "payload_cid": "bafybeib5nunwd6nmhe3x3mfzmfhrddegsrxxk6lq4lszploeplktzkxhzu",
    "commp": "baga6ea4seaqcqnnwp7n5ra5ltnvwkd3xk3jxujtxg4bqrueangl3t5cyn5p6soq",
    "padded_size": 34359738368,
    "size": 18010019221
  },
 ...
]
```

#### Response Body
```jsonc
{
	"success": [
    "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmlq",
		"baga6ea4seaqcqnnwp7n5ra5ltnvwkd3xk3jxujtxg4bqrueangl3t5cyn5p6soq"
    ..
  ],
	"fail": []
}
```

		
## /providers
### POST /
- Add a storage provider

#### Params
<nil>

#### Body
```jsonc
{
  id: "f01234" // unique! SP identifier
}
```

#### Response
> 200: Success
> 500: Fail


### GET /
- Gets list of storage providers

#### Params
<nil>

#### Body
```jsonc
[
	{
		"key": "b3cc8a99-155a-4fff-8974-999ec313e5cc",
		"actor_id": "f01963614",
		"bytes_replicated": {
			"raw": 216120230655,
			"padded": 412316860416
		},
	},
	{
		"key": "29c0c1ce-6b13-434c-8b94-49ba5a21b7a9",
		"actor_id": "f01886797",
		"bytes_replicated": {
			"raw": 0,
			"padded": 0
		},
	}
]
```

#### Response
> 200: Success
> 500: Fail

## /replication

### POST / 
- Create replications (deals)

> This endpoint requires the Delta API key in the `Authorization: Bearer XXX` header

#### Params
<none>

```jsonc
{
  provider: "f01234", // required! ID of the SP to create deals with
  dataset: "test-dataset", // optional - if unspecified, will select content from any dataset
  numDeals: 10, // optional - if unspecified, then numTib must be specified. Number of deals to make
}
```

#### Response
> 200: Success
```jsonc
[
  "bafy123", // Proposal CIDs
  "bafy456",
]
```


## /wallet
### POST /
- Add a wallet

> This endpoint requires the Delta API key in the `Authorization: Bearer XXX` header

#### Params
```json
/dataset-name // name that identifies the dataset. Must already exist (add it using /datasets POST)
```

#### Body
```jsonc
{
  Type: "secp256k1" // Wallet type
  PrivateKey: "XXX" // Private key from wallet file
}
```

#### Response
> 200: Success
> 500: Fail