
# API Methods

## /datasets

### POST /
- Add a dataset to be tracked 

#### Params
<nil>

#### Body
```jsonc
{
  name: "dataset-name" // Name / slug that identifies the dataset. Must be unique!
  replication_quota: 6 // Max. Number of replications that can be made
  duration: 540 // Deal length, in days
  wallet: "f1xxx" // Wallet to be used for datacap 
  unsealed: true // Whether unsealed copies should be created or not
  indexed: true // Whether dataset should be announced to indexer or not
}
```

#### Response
> 200: Success
> 500: Fail

### GET /
- Returns a list of all datasets

#### Params: 
```json
?provider="f0123456" // REQUIRED - ID of provider to replicate with
?dataset="dataset-name" // OPTIONAL - Name of dataset to replicate
```

#### Body: 
<nil> 

#### Response
> 200 : Success
```jsonc
[
  {
    id: 1
    name: "dataset-name"
    replications: 6
    wallet: "f1xxx"
    duration: 540 // Deal length, in days
    size: 38451017 // total size of CARfiles
    cids: 10000 // Number of CIDs / files comprising the dataset
  }
]

```

### POST /content/:dataset
- Add content (CAR files) to the dataset

### Params:
dataset: identifier (name) of dataset

### Body: 
```jsonc
[
 {
  cid: "Q1234",
  size: 1024,
  padded_piece_size: 4096
 },
 ...
]
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
{
  id: "f01234" // unique! SP identifier
  replicated_bytes: 58712698 // Number of bytes replicated to SP
  replicated_deals: 12332 // Number of deals made with SP
}
```

#### Response
> 200: Success
> 500: Fail

## /deal

### POST / 
- Create deals

> This endpoint requires the Delta API key in the `Authorization: Bearer XXX` header

#### Params
<nil>

#### Body 
```jsonc
{
  provider: "f01234", // required! ID of the SP to create deals with
  dataset: "test-dataset", // optional - if unspecified, will select content from any dataset
  numDeals: 10, // optional - if unspecified, then numTib must be specified. Number of deals to make
  numTib: 2 // optional - if unspecified, then numDeals must be specified. Amount of TiB of deals to make
  pricePerDeal : 0 // optional - amount of fil per deal. If unspecifed, makes verified deal with datacap
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
/"dataset-name" // name that identifies the dataset. Must already exist (add it using /datasets POST)
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