# OpsSC Agent

## Overview

The OpsSC agent is an agent program for operating nodes (e.g., peers) for each organization according to the directions from the OpsSC chaincodes.
Each agent listens to chaincode events issued by OpsSC chaincodes and, automatically executes operations to components for each organization based on the events and then submits the result of the operations to the OpsSC chaincodes.

In the current implementation, the agent program prepared for each organization automatically executes:
- Bootstrapping
  - When the agent is launched or receives a chaincode event that notifies channel configuration updates
    - joins their peers described in the connection profile to the ops channel
    - deploys initial OpsSCs embedded as local files on their peers
    - joins their peers described in the connection profile to to all the application channels which are managed by the OpsSC
    - deploys all the existing chaincodes (only the latest version of each) which are managed by the OpsSC on their peers
- Chaincode operations
  - When the agent receives a chaincode event that a chaincode update proposal is voted by a majority of the channel organizations
    - downloads the source code of the chaincode from the remote repository specified in the proposal
    - packages and installs the downloaded source code
    - approves the chaincode definition with the above package based on the content of the proposal
    - submits the result of the above as acknowledge to the OpsSC chaincode
  - When the agent receives a chaincode event that a chaincode update proposal is acknowledged a by all the channel organizations
    - commits the chaincode definition based on the content of the proposal (if only selected as the executor)
    - submits the result of the commit to the OpsSC chaincode
- Channel operations
  - When the agent receives a chaincode event that a channel update proposal is voted by a majority of the channel organizations
    - updates or creates the channel based on the content of the proposal (if only selected as the executor)
    - submits the result of the commit to the OpsSC chaincode

In the current implementation, the agent calls `peer` and `fabric-configtx-cli` commands or `_lifecycle` system chaincode to do the above.

## Prerequisites

- Linux
- Node.js >= 10.15
- Docker
- Docker Compose

## Environment variables

The current OpsSC agent is set via environment variables.

The following environment variables must be set:

| Category           | Variable Name           | Default Value                                | Description                                                                                                |
| ------------------ | ----------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| Hyperledger Fabric | `ADMIN_MSPID`           | `Org1MSP`                                    | MSP ID for the organization to be operated                                                                 |
| Hyperledger Fabric | `ADMIN_CERT`            | `/opt/fabric/msp/signcerts`                  | Certificate for the client identity to interact with the OpsSC chaincodes and execute peer commands        |
| Hyperledger Fabric | `ADMIN_KEY`             | `/opt/fabric/msp/keystore`                   | Private key for the client identity to interact with the OpsSC chaincodes and execute peer commands        |
| Hyperledger Fabric | `MSP_CONFIG_PATH`       | `/opt/fabric/msp`                            | MSP config path for the client identity to interact with the OpsSC chaincodes and execute peer commands    |
| Hyperledger Fabric | `DISCOVER_AS_LOCALHOST` | `false`                                      | Whether to discover as localhost                                                                           |
| Hyperledger Fabric | `CONNECTION_PROFILE`    | `/opt/fabric/config/connection-profile.yaml` | Connection profile path for the organization (NOTE: should be written all peers owned by the organization) |
| OpsSC              | `CHANNEL_NAME`          | `ops-channel`                                | Channel name for the OpsSC                                                                                 |
| OpsSC              | `CC_OPS_CC_NAME`        | `chaincode_ops`                              | Chaincode name of the chaincode OpsSC                                                                      |
| OpsSC              | `CH_OPS_CC_NAME`        | `channel_ops`                                | Chaincode name of the channel OpsSC                                                                        |
| Chaincode Ops      | `GIT_USER`              | None                                         | Git user to access to the chaincode repository (If not set, access without credentials)                    |
| Chaincode Ops      | `GIT_PASSWORD`          | None                                         | Git password to access to the chaincode repository (If not set, access without credentials)                |
| Chaincode Ops      | `GOPATH`                | None                                         | GOPATH                                                                                                     |
| WebSocket Client   | `WS_ENABLED`            | `false`                                      | Whether to enable WebSocket client to send messages to the server                                          |
| WebSocket Client   | `WS_URL`                | `ws://localhost:5000`                        | URL of the WebSocket server to connect to                                                                  |
| Logging            | `LOG_LEVEL`             | `info`                                       | Log level                                                                                                  |


## API specification

The current OpsSC agent provides the following API:

| Category | Title        | URL        | Method | URL Params | Data Params | Content in Success Response |
| -------- | ------------ | ---------- | ------ | ---------- | ----------- | --------------------------- |
| Utility  | Health check | `/healthz` | `GET`  | None       | None        | `OK`                        |

## How to run

The current OpsSC API agent is expected to run primarily as a Docker container.

*NOTE:* If run locally, it may overwrite your repositories under your GOPATH to download the proposal chaincode source code form the remote repository.

### Build the Docker image

Build the docker image by running the following commands:
```sh
cd fabric-opssc/opssc-agent
./scripts/build.sh

$ docker images # Command to confirm the images are created
REPOSITORY                                                     TAG                              IMAGE ID            CREATED             SIZE
(...)
fabric-opssc/opssc-agent                                       latest                           44e30c583566        44 hours ago        1.49GB
(...)
```

The docker image definition for the OpsSC agent is [here](../Dockerfile-for-agent).

### Run the Docker container

Refer to [an example of the docker-compose file](../sample-environments/fabric-samples/test-network/docker/docker-compose-opssc-agents.yaml).