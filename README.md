# airdrop-svc

## Description

Service to manage airdrops for passport ZK-proofs

## Install

```
git clone github.com/rarimo/airdrop-svc
cd airdrop-svc
go build main.go
export KV_VIPER_FILE=./config.yaml
./main migrate up
./main run service
```

## API documentation

[Online docs](https://rarimo.github.io/airdrop-svc/) are available.

All endpoints from docs MUST be publicly accessible.

### Local build

We do use openapi:json standard for API. We use swagger for documenting our API.

To open online documentation, go to [swagger editor](http://localhost:8080/swagger-editor/) here is how you can start it
```
  cd docs
  npm install
  npm run start
```
To build documentation use `npm run build` command,
that will create open-api documentation in `web_deploy` folder.

To generate resources for Go models run `./generate.sh` script in root folder.
use `./generate.sh --help` to see all available options.

Note: if you are using Gitlab for building project `docs/spec/paths` folder must not be
empty, otherwise only `Build and Publish` job will be passed.  

## Running from Source

* Run dependencies, based on config example
* Set up environment value with config file path `KV_VIPER_FILE=./config.yaml`
* Provide valid config file
* Launch the service with `migrate up` command to create database schema
* Launch the service with `run service` command

### Database
For services, we do use ***PostgresSQL*** database. 
You can [install it locally](https://www.postgresql.org/download/) or use [docker image](https://hub.docker.com/_/postgres/).
