
# API Methods


## /datasets

### POST /
- Add a dataset to be tracked 

Permission: `admin`

#### Params
<nil>

#### Body
```json
{
  name: "dataset-name" // Name / slug that identifies the dataset. Must be unique!
  replications: 6 // Max. Number of replications that can be made
  wallet: "f1xxx" // Wallet to be used for datacap or payment
}
```

#### Response
> 200: Success
> 500: Fail

### GET /
- Returns a list of all datasets

Permission: `admin`

#### Params: 
<nil>

#### Body: 
<nil> 

#### Response
> 200 : Success
```json
[
  {
    id: 1
    name: "dataset-name"
    replications: 6
    wallet: "f1xxx"
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
```json
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

Permission: `admin`

#### Params
<nil>

#### Body
```json
{
  id: "f01234" // unique! SP identifier
}
```

#### Response
> 200: Success
> 500: Fail


### GET /
- Gets list of storage providers

Permission: `admin`

#### Params
<nil>

#### Body
```json
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

Permission: `admin`

#### Params
<nil>

#### Body 
```json
{
  provider: "f01234", // required! ID of the SP to create deals with
  dataset: "test-dataset", // optional - if unspecified, will select content from any dataset
  numDeals: 10, // optional - if unspecified, then numTib must be specified. Number of deals to make
  numTib: 2 // optional - if unspecified, then numDeals must be specified. Amount of TiB of deals to make
}
```

#### Response
> 200: Success
```json
[
  "bafy123", // Proposal CIDs
  "bafy456",
]
```