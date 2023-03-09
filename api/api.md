
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
  actor_id: "f01234", // unique! SP identifier
	actor_name: "Friendly name" // optional - friendly sp name 
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

### GET /
- Get Replications

#### Params
```json
?provider=f01963614 // optional - filter by sp
?dataset=delta-test // optional - filter by dataset
```

#### Response
> 200: Success

```json
[
		{
		"ID": 8,
		"CreatedAt": "2023-03-01T19:16:14.956554-08:00",
		"UpdatedAt": "2023-03-02T17:21:17.260965-08:00",
		"DeletedAt": null,
		"content": {
			"commp": "baga6ea4seaqnkkxerlx4vljfvmpffjzqmxuusz4l72lil5quf7c5b2mcg5zy2mi",
			"payload_cid": "bafybeifoxwwx5newgdwnvojyotleh3x3sy7ckndwa2ysioe4corv4ixgti",
			"size": 18010019221,
			"padded_size": 34359738368,
			"dataset_name": "delta-test",
			"num_replications": 1
		},
		"deal_time": "2023-03-01T19:16:14.956401-08:00",
		"delta_content_id": 27388,
		"proposal_cid": "bafyreidvtrgw2z6l3m6slalibrb67rqxcnscjiexrsccjwnwr7pwhat5zq",
		"provider_actor_id": "f01963614",
		"content_commp": "baga6ea4seaqnkkxerlx4vljfvmpffjzqmxuusz4l72lil5quf7c5b2mcg5zy2mi",
		"status": "SUCCESS"
	},
	{
		"ID": 9,
		"CreatedAt": "2023-03-01T19:16:14.962937-08:00",
		"UpdatedAt": "2023-03-02T17:21:17.261791-08:00",
		"DeletedAt": null,
		"content": {
			"commp": "baga6ea4seaqchqjaoycetgpptpiipeuygx7h7aeuml5mfrqp7kskrqsloi6pwia",
			"payload_cid": "bafybeiezpv62emncxbe4adoyipxhzcdy2eqzxx3rde6rdzuqxs57gdsp2q",
			"size": 18010019221,
			"padded_size": 34359738368,
			"dataset_name": "delta-test",
			"num_replications": 1
		},
		"deal_time": "2023-03-01T19:16:14.962877-08:00",
		"delta_content_id": 27389,
		"proposal_cid": "bafyreid23uwyqqdwvgaaivnqzktmnpgxgo4ruo3hlk7efjgzs6lwcq75wy",
		"provider_actor_id": "f01963614",
		"content_commp": "baga6ea4seaqchqjaoycetgpptpiipeuygx7h7aeuml5mfrqp7kskrqsloi6pwia",
		"status": "SUCCESS"
	},
	{
		"ID": 18,
		"CreatedAt": "2023-03-06T11:09:42.185496-08:00",
		"UpdatedAt": "2023-03-06T11:09:43.997136-08:00",
		"DeletedAt": null,
		"content": {
			"commp": "baga6ea4seaqk3b7prx2ulmdztwbg4r4jvccxcdjqqzi3jdb25lggsgytpkxjgoy",
			"payload_cid": "bafybeiakf666idv6zs4uksckfkjr76jmvrcuu4neidldxlfngo2vh6jvfe",
			"size": 18010019222,
			"padded_size": 34359738368,
			"dataset_name": "delta-test",
			"num_replications": 3
		},
		"deal_time": "2023-03-06T11:09:42.185318-08:00",
		"delta_content_id": 27874,
		"proposal_cid": "PENDING_1508341816105618720",
		"provider_actor_id": "f01963614",
		"content_commp": "baga6ea4seaqk3b7prx2ulmdztwbg4r4jvccxcdjqqzi3jdb25lggsgytpkxjgoy",
		"status": "FAILURE",
		"delta_message": "illegal base64 data at input byte 0"
	},
	{
		"ID": 19,
		"CreatedAt": "2023-03-06T11:11:02.724047-08:00",
		"UpdatedAt": "2023-03-06T11:11:15.867567-08:00",
		"DeletedAt": null,
		"content": {
			"commp": "baga6ea4seaqk3b7prx2ulmdztwbg4r4jvccxcdjqqzi3jdb25lggsgytpkxjgoy",
			"payload_cid": "bafybeiakf666idv6zs4uksckfkjr76jmvrcuu4neidldxlfngo2vh6jvfe",
			"size": 18010019222,
			"padded_size": 34359738368,
			"dataset_name": "delta-test",
			"num_replications": 3
		},
		"deal_time": "2023-03-06T11:11:02.723922-08:00",
		"delta_content_id": 27875,
		"proposal_cid": "bafyreic7n7josf5klvdxop46zjjfr6ju4o4ywqtom2wxhpagaswel3krd4",
		"provider_actor_id": "f01963614",
		"content_commp": "baga6ea4seaqk3b7prx2ulmdztwbg4r4jvccxcdjqqzi3jdb25lggsgytpkxjgoy",
		"status": "FAILURE",
		"delta_message": "deal proposal rejected: failed validation: invalid deal end epoch 4236142: cannot be more than 1555200 past current epoch 2660782"
	}
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