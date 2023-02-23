# ΔDM (Delta Dataset Manager)

A tool to manage deal replication tracking for onboarding datasets to the Filecoin network via **import storage deals**. This provides a solution to quickly make deals for massive amounts of data, where the transfer is better handled out-of-band. 

## Data Flow

### Dataset
The top-level logical grouping of data in LDM is the **dataset**. Datasets are identified by a name (aka "slug"), along with a replication quota, deal length, and a wallet to make the deals from.
Datasets are added independently from the content making them up. 

### Content
Once a dataset has been created, content may be added to it. A content represents a .CAR file - archive of data that will be shipped to the SP and loaded into their operation. Content is identified by its **PieceCID** (CommP), has two sizes (raw file size, Padded Piece Size), and also contains a CID of the actual data (Payload CID).

### Providers
LDM tracks deals to Storage Providers in the network. Add a list of storage providers to LDM before making deals to begin tracking them.


### Replication
Once a **Dataset**, **Content**, and **Providers** have been specified, the LDM `replication` API can be called to issue a number of import deals out to the providers. 


# Instructions

- Set `DB_NAME` env var to postgres connection string ex) `testdb`. Can also be specified as a cli argument.

## Example Usage

```bash
deltadm daemon --db testdb
```

## API
See api docs in /api/api.md.

## Wishlist
- Ability to specify region / IP address of providers so that deals can be made in a geo-aware way (only replicate a certain amount to a given region)
- Show progress bar of how much datacap is being used for each dataset/wallet

### Reporting

Dataset level 
-> see how **distributed** the data is. See which SPs 
-> match what Validation-bot does 
-> see % replicated 

SP Level
-> Total list of data/CIDs replicated, which datasets