# Delta Large Dataset Manager

A tool to manage onboarding of large datasets to the Filecoin network. 

# Instructions

- Set `DB_NAME` env var to postgres connection string ex) `host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai`



## Data Onboarding Flow


### Set Up
1. Add the dataset into Delta-LDM -> specify the name(slug), replication quota, deal length, and associated wallet for datacap
2. Add Storage Providers to Delta-LDM, by actor ID / identifier
3. Associate content (CAR files) with the dataset. Specify each by CID and provide sizes (padded and raw)

### Deal Making
1. Call `/deal` endpoint to make deals with providers- specify either a # of deals or amount (TiB) to replicate



#### Wishlist
- Ability to specify region / IP address of providers so that deals can be made in a geo-aware way (only replicate a certain amount to a given region)
- Never deal the same CAR to the same provider! 
- Show progress bar of how much datacap is being used for each dataset/wallet

### Reporting

Dataset level 
-> see how **distributed** the data is. See which SPs 
-> match what Validation-bot does 
-> see % replicated 

SP Level
-> Total list of data/CIDs replicated, which datasets