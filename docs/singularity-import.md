# Importing datasets from Singularity

If you have generated CAR files via Singularity, you can import them into Delta for dealmaking.

To do this, ensure your Singularity database is running and you have the `DELTA_AUTH` environment variable set to your Delta auth token

## Prerequisites
- `mongoexport` - install it for your distro at https://www.mongodb.com/try/download/database-tools

## Importing CIDs from Singularity
- First, create datasets and wallets in Delta (see [docs/howto.md](docs/howto.md))
- Edit the script `singularity-import.sh` - replace `localhost:27017` with your singularity database URL, if it's not the default
  - Tip: You can find this information in `~/.singularity/config.toml`
- Run the script in `scripts/singularity-import.sh` to import CIDs from Singularity into DDM. This will add the CIDs to the dataset.
- Note that the script may take a few minutes to run

**Script Usage**
```bash
./singularity-import.sh <SINGULARITY datasetName> <DELTA datasetName> <deltaToken>
```

Where
- SINGULARITY datasetName - the name of the dataset in Singularity (aka the slug)
- DELTA datasetName - the name of the dataset in Delta
- deltaToken - your delta auth token

**Example**
```bash
./singularity-import.sh common-crawl common-crawl EST-XXX-ARY
```