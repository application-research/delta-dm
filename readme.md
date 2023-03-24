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

- Set `DELTA_AUTH` environment variable to Delta API key. It can also be provided as the CLI flag `--delta-auth`
- DDM will default to a [delta](https://github.com/application-research/delta) instance running at `localhost:1414`. It must be running or DDM will not start. Override the url by providing specifying the `DELTA_API` environment variable, or CLI flag `--delta-api`
- DDM will use the [Estuary Auth server](https://github.com/application-research/estuary-auth/) by default. It can be overridden by specifying the `AUTH_URL` environment variable, or CLI flag `--auth-url`

## Usage
DDM runs as a daemon, which is a webserver. Start it up with the `daemon` command.
```bash
./deltadm daemon
```

Once running, you can interact with DDM through the API, CLI, or via the [Delta Web frontend](https://github.com/application-research/delta-nextjs-client/)

## API
See api docs in [/docs/api.md](/docs/api.md).

## Command-Line Interface
See cli docs in [/docs/cmd.md](/docs/cmd.md).

## Provider Self-service
See docs in [/docs/self-service.md](/docs/self-service.md).

## Importing CIDs from Singularity
See docs in [/docs/singularity-import.md](/docs/singularity-import.md).