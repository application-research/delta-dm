# [WIP] ΔDM (Delta Dataset Manager)

> This tool is still in very active development! Things may change/break as we add functionality and streamline the user experience.

A tool to manage deal replication tracking for onboarding datasets to the Filecoin network via **import storage deals**. This provides a solution to quickly make deals for massive amounts of data, where the transfer is better handled out-of-band. 

## Data Flow

### Dataset
The top-level logical grouping of data in DDM is the **dataset**. Datasets are identified by a name (aka "slug"), along with a replication quota, deal length, and a wallet to make the deals from.
Datasets are added independently from the content making them up. 

### Content
Once a dataset has been created, content may be added to it. A content represents a .CAR file - archive of data that will be shipped to the SP and loaded into their operation. Content is identified by its **PieceCID** (CommP), has two sizes (raw file size, Padded Piece Size), and also contains a CID of the actual data (Payload CID).

### Providers
DDM tracks deals to Storage Providers in the network. Add a list of storage providers to DDM before making deals to begin tracking them.


### Replication
Once a **Dataset**, **Content**, and **Providers** have been specified, the DDM `replication` API can be called to issue a number of import deals out to the providers. 


# Instructions

- Set `DB_NAME` env var to postgres connection string ex) `testdb`. Can also be specified as a cli argument.

## Example Usage

```bash
deltadm daemon --db testdb
```

## API
See api docs in /api/api.md.

## Importing CIDs from Singularity
- First, create datasets and wallets in Delta (follow above guide)
- Use the script in `scripts/singularity-import.sh` to import CIDs from Singularity into DDM. This will add the CIDs to the dataset.