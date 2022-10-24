# OpsSC API Server

## Overview

This REST API server provides an API for each organization's administrator to interact with (to invoke and/or query transactions to) the OpsSC chaincodes.
Mainly, it has API to request a channel or chaincode update proposal and to approve (vote for) the proposal.
The server does some internal processing to translate user requests through the API into the parameters for the OpsSC chaincodes
by sometimes executing `peer` and `fabric-configtx-cli` commands or `_lifecycle` system chaincode.

This also works as a WebSocket server to notify the chaincode events and activities of the OpsSC agents.
In a typical use case, an administrator for each organization communicates with the API server through a GUI like a Web admin portal.

## Prerequisites

- Linux
- Node.js >= 12.21
- Docker
- Docker Compose

## Environment variables

The current OpsSC API server is set via environment variables.

The following environment variables must be set:

| Category           | Variable Name             | Default Value                                | Description                                                                                             |
| ------------------ | ------------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Hyperledger Fabric | `ADMIN_MSPID`             | `Org1MSP`                                    | MSP ID for the organization to be operated                                                              |
| Hyperledger Fabric | `ADMIN_CERT`              | `/opt/fabric/msp/signcerts`                  | Certificate for the client identity to interact with the OpsSC chaincodes and execute peer commands     |
| Hyperledger Fabric | `ADMIN_KEY`               | `/opt/fabric/msp/keystore`                   | Private key for the client identity to interact with the OpsSC chaincodes and execute peer commands     |
| Hyperledger Fabric | `MSP_CONFIG_PATH`         | `/opt/fabric/msp`                            | MSP config path for the client identity to interact with the OpsSC chaincodes and execute peer commands |
| Hyperledger Fabric | `DISCOVER_AS_LOCALHOST`   | `false`                                      | Whether to discover as localhost                                                                        |
| Hyperledger Fabric | `CONNECTION_PROFILE`      | `/opt/fabric/config/connection-profile.yaml` | Connection profile path for the organization                                                            |
| OpsSC              | `CHANNEL_NAME`            | `ops-channel`                                | Channel name for the OpsSC                                                                              |
| OpsSC              | `CC_OPS_CC_NAME`          | `chaincode-ops`                              | Chaincode name of the chaincode OpsSC                                                                   |
| OpsSC              | `CH_OPS_CC_NAME`          | `channel-ops`                                | Chaincode name of the channel OpsSC                                                                     |
| API Server         | `CLIENT_SERVICE_PORT`     | `5000`                                       | Port number used by the API server                                                                      |
| API Server         | `API_CH_PROPOSAL_ENABLED` | `true`                                       | Whether to enable the Channel Update Proposal APIs                                                      |
| API Server         | `API_UTIL_ENABLED`        | `true`                                       | Whether to enable the Utility APIs                                                                      |
| WebSocket          | `WS_ENABLED`              | `false`                                      | Whether to enable WebSocket server to receive messages from agents or the API server itself             |
| Logging            | `LOG_LEVEL`               | `info`                                       | Log level                                                                                               |

## API specification

The current OpsSC API server provides the following API to enable administrators to operate chaincodes and channels with communicating the other organizations:

| Category  | Title                                                                  | URL                                           | Method |
| --------- | ---------------------------------------------------------------------- | --------------------------------------------- | ------ |
| Chaincode | Get all update proposals                                               | `/api/v1/chaincode/proposals`                 | `GET`  |
| Chaincode | Request a new update proposal                                          | `/api/v1/chaincode/proposals/:id`             | `POST` |
| Chaincode | Get the proposal with the given ID                                     | `/api/v1/chaincode/proposals/:id`             | `GET`  |
| Chaincode | Vote for/against the proposal                                          | `/api/v1/chaincode/proposals/:id/vote`        | `POST` |
| Chaincode | Withdraw the proposal                                                  | `/api/v1/chaincode/proposals/:id/withdraw`    | `POST` |
| Chaincode | Get the task histories with the given proposal                         | `/api/v1/chaincode/:id/histories`             | `GET`  |
| Chaincode | Get the list of installed chaincodes                                   | `/api/v1/chaincode/getInstalledChaincodes`    | `GET`  |
| Chaincode | Get the chaincode definition with the given channel and chaincode name | `/api/v1/chaincode/queryChaincodeDefinition`  | `GET`  |
| Chaincode | Get the list of chaincode definitions with the given channel name      | `/api/v1/chaincode/queryChaincodeDefinitions` | `GET`  |
| Channel   | Get all update proposals                                               | `/api/v1/channel/proposals`                   | `GET`  |
| Channel   | Request a new update proposal                                          | `/api/v1/channel/proposals/:id`               | `POST` |
| Channel   | Get the proposal with the given ID                                     | `/api/v1/channel/proposals/:id`               | `GET`  |
| Channel   | Vote for the proposal                                                  | `/api/v1/channel/proposals/:id/vote`          | `POST` |
| Channel   | Get the channel information with the given channel name                | `/api/v1/channel/getChannel`                  | `GET`  |
| Channel   | Get the list of all channels                                           | `/api/v1/channel/getChannels`                 | `GET`  |
| Channel   | Get the system config block                                            | `/api/v1/channel/systemConfigBlock`           | `GET`  |
| Utility   | Invoke a chaincode (for test)                                          | `/api/v1/utils/invokeTransaction`             | `POST` |
| Utility   | Query a chaincode (for test)                                           | `/api/v1/utils/queryTransaction`              | `GET`  |
| Common    | Get the organization information to be operated                        | `/api/v1/organization`                        | `GET`  |
| Common    | Version check                                                          | `/api/v1/version`                             | `GET`  |
| Common    | Health check                                                           | `/healthz`                                    | `GET`  |

Refer to [API specification](./APISpecification.md) for the details

## How to run

The current OpsSC API server is expected to run primarily as a Docker container.

### Build the Docker image

Build the docker image for API Server by running the following commands:

```sh
$ cd fabric-opssc
$ make docker-opssc-api-server

$ docker images # Command to confirm the images are created
REPOSITORY                                                     TAG                              IMAGE ID            CREATED             SIZE
(...)
fabric-opssc/opssc-api-server                                  latest                           154c4a550823        44 hours ago        1.43GB
(...)
```

The docker image definition for the OpsSC API server is [here](../Dockerfile-for-api-server).

### Run the Docker container

Refer to [an example of the docker-compose file](../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-api-servers.yaml).
