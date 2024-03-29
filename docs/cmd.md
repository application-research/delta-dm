# DDM CLI

> Square brackets (ex, `[--dataset <dataset-name>]`) indicate optional arguments.

# daemon

In order for other commands to work, you must have a running daemon to connect to, and execute the commands against.
## Run the DDM Daemon
`> ./delta-dm daemon`
*Note*: you must have `DELTA_API="http://url-to-delta"` in your environment, or it will default to `http://localhost:1414`

# Command Line - Interacting with DDM
*Note* Please ensure you have `DELTA_AUTH=DEL-XXX-TA` auth key in your environment before running any of these commands below.

## wallet
### Import a wallet
#### By keyfile
`> ./delta-dm wallet import --file <path-to-file>`

Example:
```bash
./delta-dm wallet import --file ~/.lotus/keystore/O5QWY3DFOQWWMMLNNVRDG3DYG5WG46TLO5ZXM2DSNFSHM4DVM5XHK6TPGRWXCMTYNJWWC53WNZTGS 
```

#### By wallet hex
`> ./delta-dm wallet import --hex <wallet hex>`

Example: Import directly from lotus wallet export (hex)
`> ./delta-dm wallet import --hex $(lotus wallet export f1mmb3lx7lnzkwsvhridvpugnuzo4mq2xjmawvnfi)`


### Delete a wallet
`> ./delta-dm wallet delete <address>`

Example:
```bash
./delta-dm wallet delete f1mmb3lx7lnzkwsvhridvpugnuzo4mq2xjmawvnfi
```

### Associate wallet with dataset
`> ./delta-dm wallet associate --address <address> --datasets <dataset-name-1>,<dataset-name-2>`

Example:
```bash
./delta-dm wallet associate --address f1mmb3lx7lnzkwsvhridvpugnuzo4mq2xjmawvnfi --datasets delta-test,delta-test-2
```

### List wallets
`> ./delta-dm wallet list [--dataset <dataset-name>]`


## provider
### Add a provider
`> ./delta-dm provider add --id <sp-actor-id> [--name <friendly-name>]`

Example:
```bash
./delta-dm provider add --id f01000 --name "My Provider"
```

### Modify a provider
`> ./delta-dm provider modify --id <sp-actor-id> [--name <friendly-name>] [--allowed-datasets <datasets>] [--allow-self-service <on|off>] `

Example:
```bash
./delta-dm provider modify --id f01000 --name "My Provider" --allowed-datasets delta-test,delta-test-2 --allow-self-service on
```

### List providers
`> ./delta-dm provider list`

## dataset
### Add a dataset
`> ./delta-dm dataset add --name <dataset-name> [--replication-quota <quota>] [--duration <deal-duration-days>]`

Example:
```bash
./delta-dm dataset add --name delta-test --replication-quota 6 --duration 540
```

### List datasets
`> ./delta-dm dataset list`

## replication
### Create a replication
`> ./delta-dm replication create --provider <sp-actor-id> -num <num-deals-to-make> [--dataset <dataset-id>] [--delay-start <delay-start-days>]`

Example:
```bash
./delta-dm replication create --provider f01000 --num 3 --dataset 1 --delay-start 3
```

## content
### Import content to a dataset
`> ./delta-dm content import --dataset <dataset-id> [--json <path-to-json-file>] [--csv <path-to-csv-file>] [--singularity <path-to-singularity-export-json-file>]`

One of `--json`, `--csv`, or `--singularity` must be provided.

For the expected file format, see the [api docs](api.md##/contents)

Example:
```bash
./delta-dm content import --dataset 1 --json ./content.json
```

### List content in a dataset
`> ./delta-dm content list --dataset <dataset-id>`


## replication profiles
- Note: `replication-profile`/`rp` commands take a `dataset id`, you can run `dataset list` to get the id for a dataset.
### Add a replication profile
`> ./delta-dm rp add --spid <sp-id> --dataset <dataset-id> [--unsealed] [--indexed]`

Example:
```bash
./delta-dm rp add --spid f01000 --dataset 1 --unsealed --indexed
```

### Modify a replication profile
`> ./delta-dm rp modify --spid <sp-id> --dataset <dataset-id> [--unsealed] [--indexed]`

### Delete a replication profile
`> ./delta-dm rp delete --spid <sp-id> --dataset <dataset-id>`

### List replication profiles
`> ./delta-dm rp list`