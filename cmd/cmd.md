# DDM CLI

# daemon
## Run the DDM Daemon
`> ./delta-dm daemon`
*Note*: you must have `DELTA_API="http://url-to-delta"` in your environment, or it will default to `http://localhost:1414`

# Command Line - Interacting with DDM
*Note* Please ensure you have `DELTA_AUTH=DEL-XXX-TA` auth key in your environment before running any of these commands below.

## wallet
### Import a wallet
By keyfile
`> ./delta-dm wallet import --dataset dataset-name --file <path-to-file>`

By raw json
`> ./delta-dm wallet import --dataset dataset-name --json {"Type":"secp256k1","PrivateKey":"XXX"}`


the `dataset` flag is optional. If unspecified, the newly added wallet will not be associated with any dataset