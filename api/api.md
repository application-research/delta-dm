
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