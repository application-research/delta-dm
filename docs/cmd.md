# DDM CLI

# daemon
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
`> ./delta-dm wallet import --hex "7b2254797065...."`

#### Or directly from Lotus export to DDM Import (hex)
`> ./delta-dm wallet import --hex $(lotus wallet export f1mmb3lx7lnzkwsvhridvpugnuzo4mq2xjmawvnfi)`
