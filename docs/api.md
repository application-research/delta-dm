
# API Methods

All endpoints are prefixed with `/api/v1`.

For example, `http://localhost:1314/api/v1/datasets`

All endpoints (with the exception of `/self-service`) require the `Authorization: Bearer <XXX>` header present on the request. It must match the `Delta API Key` that is passed into `delta-dm daemon` in order to be permitted to make requests.

## /health

### GET /health
- Checks to ensure delta-dm and delta are running
- Returns back the Delta instance ID, which can be used for further troubleshooting
- Returns versions and commit hashes of DDM and Delta

#### Request Params
<nil>

#### Request Body 
<nil>

#### Response
> 200: Success

```jsonc
{
	"uuid": "504fd7de-729d-44f1-a70c-e9d8cc8c59ba",
	"ddm_info": {
		"commit": "6f12184",
		"version": "v0.0.0"
	},
	"delta_info": {
		"commit": "e8296f720eab063ebc230b66951fb152248b02fc",
		"version": "v0.0.0"
	}
}
```

## /datasets

### POST /datasets
- Add a dataset to be tracked 
- Note: the dataset name must contain only lowercase letters, numbers and hypens, and must be less than 255 characters in length. It must start and end with a letter, and double-dashes are not allowed.

#### Request Params
<nil>

#### Request Body
```jsonc
{
	"name": "delta-test",
	"replication_quota": 6,
	"deal_duration": 540,
}
```

#### Response
> 200: Success
> 500: Fail

### PUT /datasets/:dataset
- Update a dataset
- Note: the dataset name must contain only lowercase letters, numbers and hypens, and must be less than 255 characters in length. It must start and end with a letter, and double-dashes are not allowed.

#### Params
```
:dataset // Dataset ID
```

#### Body
```jsonc
{
	"name": "delta-test",
	"replication_quota": 6,
	"deal_duration": 540,
}
```

#### Response
> 200: Success
> 500: Fail

### GET /datasets
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
		"wallets": [
			{
			"address": "f1tuoahmuwfhnxpugqigxliu4muasggezw2efuczq",
			"dataset_name": "delta-test",
			"type": "secp256k1"
			}
		],
		"contents": null,
		"count_replicated": 21, // # of successful replications/storage deals 
		"count_total": 210, // total # of contents for this dataset
		"bytes_replicated": [
			198110211431, // Raw bytes (the content itself)
			377957122048 // Padded bytes (i.e, filecoin piece)
		],
		"bytes_total": [
			1801001922192, // Raw bytes (the content itself)
			3435973836800 // Padded bytes (i.e, filecoin piece)
		],
		"replication_profiles": [
			{
				"provider_actor_id": "f012345",
				"dataset_id": 1,
				"unsealed": false,
				"indexed": false
			}
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
		"contents": null,
		"count_replicated": 14, // # of successful replications/storage deals 
		"count_total": 440, // total # of contents for this dataset
		"bytes_replicated": [
			198110211431, // Raw bytes (the content itself)
			377957122048 // Padded bytes (i.e, filecoin piece)
		],
		"bytes_total": [
			1801001922192, // Raw bytes (the content itself)
			3435973836800 // Padded bytes (i.e, filecoin piece)
		]
		"replication_profiles": [
			{
				"provider_actor_id": "f012345",
				"dataset_id": 2,
				"unsealed": false,
				"indexed": false
			}
		]
	}
]

```


## /contents

### POST /contents
- Add content (CAR files) to collections

#### Request Body

##### delta-dm format
```jsonc
[
  {
    "payload_cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhy",
    "commp": "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmlq",
    "padded_size": 34359738368,
    "size": 18010019221,
    "collection": "collection-1",
    "content_location": "http://location.of.content.com/bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhy.car"
  },
  {
    "payload_cid": "bafybeib5nunwd6nmhe3x3mfzmfhrddegsrxxk6lq4lszploeplktzkxhzu",
    "commp": "baga6ea4seaqcqnnwp7n5ra5ltnvwkd3xk3jxujtxg4bqrueangl3t5cyn5p6soq",
    "padded_size": 34359738368,
    "size": 18010019221,
    "collection": "collection-2",
    "content_location": "http://location.of.content.com/bafybeib5nunwd6nmhe3x3mfzmfhrddegsrxxk6lq4lszploeplktzkxhzu.car"
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

### POST /contents/:dataset
- Add content (CAR files) to the dataset
- Accepts three types of input - standard (delta-dm) format, singularity format, or CSV- as defined below
- The :dataset parameter is the ID (uint) of the dataset to add the content to

#### Request Params
```jsonc
/dataset // ID of dataset to add content to
?import_type=<type> // singularity or csv. omit for standard format.
```

#### Request Body

##### delta-dm format
```jsonc
[
  {
    "payload_cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhy",
    "commp": "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmlq",
    "padded_size": 34359738368,
    "size": 18010019221,
	"content_location": "http://location.of.content.com/bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhy.car" // Optional
  },
  {
    "payload_cid": "bafybeib5nunwd6nmhe3x3mfzmfhrddegsrxxk6lq4lszploeplktzkxhzu",
    "commp": "baga6ea4seaqcqnnwp7n5ra5ltnvwkd3xk3jxujtxg4bqrueangl3t5cyn5p6soq",
    "padded_size": 34359738368,
    "size": 18010019221, 
	"content_location": "http://location.of.content.com/bafybeib5nunwd6nmhe3x3mfzmfhrddegsrxxk6lq4lszploeplktzkxhzu.car" // Optional
  },
 ...
]
```

##### singularity format
```jsonc
[
	{
		"carSize": 24692724205,
		"dataCid": "bafybeien5jzez2bsn3eluylu2mcdubnnz7bqbx6umw7ijxwl7ghauaaxjq",
		"pieceCid": "baga6ea4seaqjtm4sapz4vlxur6m37griffry266er5jwrwvpoodfg7mfl33ukcy",
		"pieceSize": 34359738368
	},
	{
		"carSize": 20709814175,
		"dataCid": "bafybeif2bu5bdqc6bkcpzg3h24vnavkva7lhjmemd6cwzrjehrtose7pfy",
		"pieceCid": "baga6ea4seaqhf2ymr6ahkxe3i2txmnqbmltzyf65nwcdvq2hvwmcx4eu4wzl4fi",
		"pieceSize": 34359738368
	}
]
```

##### CSV format
```csv
commP,payloadCid,size,paddedSize
baga6ea4seaqjtm4sapz4vlxur6m37griffry266er5jwrwvpoodfg7mfl33ukcy,bafybeien5jzez2bsn3eluylu2mcdubnnz7bqbx6umw7ijxwl7ghauaaxjq,24692724205,34359738368
baga6ea4seaqhf2ymr6ahkxe3i2txmnqbmltzyf65nwcdvq2hvwmcx4eu4wzl4fi,bafybeif2bu5bdqc6bkcpzg3h24vnavkva7lhjmemd6cwzrjehrtose7pfy,20709814175,34359738368
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

### GET /contents/:dataset
- Get list of contents in a dataset

#### Request Params
```jsonc
/dataset // dataset ID to get contents for
```



#### Request Body
<nil>

#### Response
> 200: Success
> 500: Fail

```json
[
	{
		"commp": "baga6ea4seaqlxodkgpb5j34cq2bamnhcn73wdma763d3ylxk5wemldpjrpxnkmy",
		"payload_cid": "bafybeifyaefzfalorttcqfcvago2rbide3mnm72geau6xxdl6iewc5leki",
		"size": 26619574156,
		"padded_size": 34359738368,
		"dataset_id": 1,
		"num_replications": 0
	},
	{
		"commp": "baga6ea4seaqaqoogvy2fkicdzm5xbmpcn4vsffapc54tfl4nbrlbfczkqsuxooi",
		"payload_cid": "bafybeiaupshs7vgsgs5e4y6n7tqkz4ghuyt3teqmqqad6ee5drlbg6dcfq",
		"size": 24389555373,
		"padded_size": 34359738368,
		"dataset_id": 2,
		"num_replications": 0
	}
]
```

## /providers
### POST /providers
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

### PUT /providers/:provider
- Update a storage provider

#### Params
```
:provider // SP actor ID
```

#### Body
```jsonc
{
	actor_name: "Friendly name" // optional - friendly sp name 
	allow_self_service: "on" // allow self-service replications ("on" or "off")
}
```

#### Response
> 200: Success
> 500: Fail


### GET /providers
- Gets list of storage providers

#### Params
<nil>

#### Body
```jsonc
[
{
		"key": "b3cc8a99-155a-4fff-8974-999ec313e5cc",
		"actor_id": "f0123456",
		"actor_name": "friendly sp",
		"allow_self_service": false,
		"bytes_replicated": {
			"raw": 234130249877,
			"padded": 446676598784
		},
		"count_replicated": 12,
		"replication_profiles": [
			{
				"provider_actor_id": "f0123456",
				"dataset_id": 1,
				"unsealed": false,
				"indexed": false
			},
			{
				"provider_actor_id": "f0123456",
				"dataset_id": 2,
				"unsealed": false,
				"indexed": false
			}
		]
	},
	{
		"key": "29c0c1ce-6b13-434c-8b94-49ba5a21b7a9",
		"actor_id": "f0998272",
		"actor_name": "test sp",
		"allow_self_service": true,
		"replication_profiles": [],
		"bytes_replicated": {
			"raw": 0,
			"padded": 0
		},
		"count_replicated": 0,
	},
]
```

#### Response
> 200: Success
> 500: Fail

## /replications

### POST /replications
- Create replications (deals)

#### Params
<none>

```jsonc
{
  provider_id: "f01234", // required! ID of the SP to create deals with
  dataset_id: 1, // optional - if unspecified, will select content from any dataset
  num_deals: 10, // Number of deals to make
	delay_start_days: 3 // Optional - delay start of deals by this many days. Default is 3. Must be between 1 and 14.
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

### GET /replications
- Get Replications

#### Params
Note: When multiple parameters are specified, they are "AND"ed together. For example, if you specify `?statuses=success&providers=f012345`, then you will only get replications that are successful and were created with the provider `f012345`.

Specifying a `proposal_cid` or `piece_cid` will override all other parameters.

```json
?statuses=success,failure,pending // can specify multiple (comma delimited), returns any that match
?datasets=bird-sounds,university-dataset-test // can specify multiple (comma delimited), returns any that match
?self_service=true
?providers=f012345,f045678 // can specify multiple (comma delimited), returns any that match
?deal_time_start=1579343980 // unix timestamp (in seconds)
?deal_time_end=1679343980// unix timestamp (in seconds)
?proposal_cid=bafyreib5sip7i4aflvxx3wpze4sdunsuo3ad7hfl3zu6n4mfontzxhviga // only one may be specified
?piece_cid=baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmlq // only one may be specified
?message=illegal // searches all replications where the message contains this text
?limit=100 // max number of replications to return (default=100)
?offset=0 // offset to start returning replications from (default=0)
```

#### Response
> 200: Success

Note the response contains two properties. `totalCount` is the total number of replications given the filter parameters (ignoring limit/offset), and `data` contains the actual replication data.

```json
{
	"data": [
		{
			"ID": 2256,
			"CreatedAt": "2023-05-31T15:34:16.052673274-07:00",
			"UpdatedAt": "2023-05-31T15:34:28.037248473-07:00",
			"DeletedAt": null,
			"content": {
				"commp": "baga6ea4seaqbcp2ujtp4v2ldidqe3saohzdfk2ssmg2ksad4f7fwghsrcxbruka",
				"payload_cid": "bafybeihfm75r5p3jd7365w4p4pkanjfnwni4cfubjvkxlzqksmtqzjj6w4",
				"size": 18215910897,
				"padded_size": 34359738368,
				"dataset_id": 1,
				"num_replications": 4
			},
			"deal_time": "2023-05-31T15:34:16.052586454-07:00",
			"delta_content_id": 2403,
			"deal_uuid": "7a08ecb8-fea4-4e28-b3c9-c3216ca7f182",
			"on_chain_deal_id": 0,
			"proposal_cid": "bafyreidlrtmbjtfp2uniw5tsc76r7bn5e5f5kmoph26d446ipuxwci2kma",
			"provider_actor_id": "f01963614",
			"content_commp": "baga6ea4seaqbcp2ujtp4v2ldidqe3saohzdfk2ssmg2ksad4f7fwghsrcxbruka",
			"status": "SUCCESS",
			"self_service": {
				"is_self_service": true,
				"last_update": "0001-01-01T00:00:00Z",
				"status": "",
				"message": ""
			}
		},
		{
			"ID": 2255,
			"CreatedAt": "2023-05-31T15:32:20.998182439-07:00",
			"UpdatedAt": "2023-05-31T15:32:27.494745191-07:00",
			"DeletedAt": null,
			"content": {
				"commp": "baga6ea4seaqfcjignlqka2qdox5m4re2jvhztddeswpmfdrmbol4ttgejdymaiy",
				"payload_cid": "bafybeifwnmgg5ielqxcfwhy2nqxfqeguoruozdt2cgzuztxfbfeiekvgrm",
				"size": 18143697080,
				"padded_size": 34359738368,
				"dataset_id": 1,
				"num_replications": 4
			},
			"deal_time": "2023-05-31T15:32:20.998080359-07:00",
			"delta_content_id": 2402,
			"deal_uuid": "3411e5e3-f169-4d4f-a841-b12ffa4e56fb",
			"on_chain_deal_id": 0,
			"proposal_cid": "bafyreib3e5dzbut5edfj5qfiv7ex3q2nni7qpm3sbd23cwycl2crlj4txe",
			"provider_actor_id": "f01963614",
			"content_commp": "baga6ea4seaqfcjignlqka2qdox5m4re2jvhztddeswpmfdrmbol4ttgejdymaiy",
			"status": "SUCCESS",
			"self_service": {
				"is_self_service": true,
				"last_update": "0001-01-01T00:00:00Z",
				"status": "",
				"message": ""
			}
		},
		{
			"ID": 2255,
			"CreatedAt": "2023-05-31T15:32:20.998182439-07:00",
			"UpdatedAt": "2023-05-31T15:32:27.494745191-07:00",
			"DeletedAt": null,
			"content": {
				"commp": "baga6ea4seaqfcjignlqka2qdox5m4re2jvhztddeswpmfdrmbol4ttgejdymaiy",
				"payload_cid": "bafybeifwnmgg5ielqxcfwhy2nqxfqeguoruozdt2cgzuztxfbfeiekvgrm",
				"size": 18143697080,
				"padded_size": 34359738368,
				"dataset_id": 1,
				"num_replications": 4
			},
			"deal_time": "2023-05-31T15:32:20.998080359-07:00",
			"delta_content_id": 2402,
			"deal_uuid": "3411e5e3-f169-4d4f-a841-b12ffa4e56fb",
			"on_chain_deal_id": 0,
			"proposal_cid": "bafyreib3e5dzbut5edfj5qfiv7ex3q2nni7qpm3sbd23cwycl2crlj4txe",
			"provider_actor_id": "f01963614",
			"content_commp": "baga6ea4seaqfcjignlqka2qdox5m4re2jvhztddeswpmfdrmbol4ttgejdymaiy",
			"status": "FAILURE",
			"delta_message": "deal proposal rejected: failed validation: invalid deal end epoch 4236142: cannot be more than 1555200 past current epoch 2660782",
			"self_service": {
				"is_self_service": false,
				"last_update": "0001-01-01T00:00:00Z",
				"status": "",
				"message": ""
			}
		},
	],
	"totalCount": 4,
}
```

## /wallets
### POST /wallets
- Add a wallet

#### Params
```json
?hex // OPTIONAL: if true, expects wallet input in Hex format (see Hex wallet import below)
```

#### Body
**Private Key wallet import**
```jsonc
{
  Type: "secp256k1" // Wallet type
  PrivateKey: "XXX" // Private key from wallet file
}
```

**Hex wallet import**
```jsonc
{
	hex_key: "7b2254797065..." // hex representation of wallet file i.e from `lotus wallet export <addr>` command
}
```

#### Response
> 200: Success
> 500: Fail

### GET /wallets
- Get all wallets

#### Params
<none>

#### Response 
> 200: Success

```json
[
	{
		"address": "f1tuoahmuwfhnxpugqigxliu4muasggezw2efuczq",
		"datasets": [
			{
				...
				"ID": 1,
				"name": "delta-test",
			}
		],
		"type": "secp256k1",
		"balance": {
			"balance_filecoin": 775398756064282, // fil balance (in attofil)
			"balance_datacap": 0 // storage power (in bytes)
		}
	}
]
```

### DELETE /wallets
- Delete a wallet

#### Params
```json
/:wallet // address of wallet to delete
```

#### Response 
> 200: Success

```json
"successfully deleted wallet" 
```

### POST /wallets/associate
- Associate a wallet with datasets

#### Params
<none>

#### Body
```jsonc
{
	"address": "f1mmb3lx7lnzkwsvhridvpugnuzo4mq2xjmawvnfi",
	"datasets": [1, 2] // ids of datasets to associate
}
```

#### Response 
> 200: Success

```json
"successfully associated wallet with datasets" 
```


## /self-service
### GET /self-service/by-cid

This endpoint requires the Provider's self-service key is present in the header in the form: 

```sh
X-DELTA-AUTH: b3cc8a99-155a-4fff-8974-999ec313e5cc
```

For more details, see the [Self-Service API](/docs/self-service.md) documentation.

#### Params
```s
/:cid # CID of content to replicate 
?start_epoch_delay # delay, in number of days, before deal starts (default: 3)
?end_epoch_advance # delay, in number of days, to advance end epoch (default: 0)
```

#### Body
<none>

#### Response
> 200: Success
```sh
{
	"cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhx",
	"content_location": "http://google.com/carfile.car"
}
```

### GET /self-service/by-dataset

This endpoint requires the Provider's self-service key is present in the header in the form: 

```sh
X-DELTA-AUTH: b3cc8a99-155a-4fff-8974-999ec313e5cc
```

For more details, see the [Self-Service API](/docs/self-service.md) documentation.

#### Params
```s
/:dataset # name of dataset to replicate for
?start_epoch_delay # delay, in number of days, before deal starts (default: 3)
?end_epoch_advance # delay, in number of days, to advance end epoch (default: 0)
```

#### Body
<none>

#### Response
> 200: Success
```json
{
	"cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhx",
	"content_location": "http://google.com/carfile.car"
}
```

### GET /self-service/eligible_pieces

Returns a list of contents that is downloadable by the client, which can then have deals requested for it.
This endpoint requires the Provider's self-service key is present in the header in the form: 

```sh
X-DELTA-AUTH: b3cc8a99-155a-4fff-8974-999ec313e5cc
```

#### Params
```s
?limit # max number of records to return (default: 500)
```

#### Body
<none>

#### Response
> 200: Success
```json
[
	{
		"payload_cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhe",
		"piece_cid": "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmle",
		"size": 18010019221,
		"padded_size": 34359738368,
		"content_location": "http://google.com/carfile"
	},
	[
	{
		"payload_cid": "bafybeidylyizmuhqny6dj5vblzokmrmgyq5tocssps3nw3g22dnlty7bhe",
		"piece_cid": "baga6ea4seaqblmkqfesvijszk34r3j6oairnl4fhi2ehamt7f3knn3gwkyylmle",
		"size": 18010019221,
		"padded_size": 34359738368,
		"content_location": "http://google.com/carfile"
	}
]
]
```

## /replication-profiles

### GET /replication-profiles
- Get all replication profiles

#### Params
<none>

#### Body
<none>

#### Response
```json
[
	{
		"provider_actor_id": "f012345",
		"dataset_id": 1,
		"unsealed": false,
		"indexed": false
	},
	{
		"provider_actor_id": "f012345",
		"dataset_id": 2,
		"unsealed": true,
		"indexed": false
	}
]
```

### PUT /replication-profiles
- Update a replication profile

#### Params
<none>

#### Body
```json
{
	"provider_actor_id": "f012345", 
	"dataset_id": 1,
	"unsealed": true,
	"indexed": true
}
```

- Note: `provider_actor_id` and `dataset_id` cannot be changed with the PUT request - they are used to identify the profile to update.

#### Response
> 200: Success

```json	
{
	"provider_actor_id": "f012345",
	"dataset_id": 1,
	"unsealed": true,
	"indexed": true
}
```

### DELETE /replication-profiles
- Delete a replication profile

#### Params
<none>

#### Body
```json
{
	"provider_actor_id": "f012345", 
	"dataset_id": 2,
}
```

#### Response
> 200: Success

```json
"replication profile with ProviderActorID f012345 and DatasetID 2 deleted successfully"
```
