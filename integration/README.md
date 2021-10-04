# Integration Tests

## Overview

The current integration tests are scenario tests are written in typescript and use Cucumber.
The integration tests internally run (stand up and tear down) `samples-environments/fabric-samples/test-network`.
Also, the tests download (replace) the Fabric binaries in `samples-environments/fabric-samples/bin` before each scenarios according to the test target Fabric version.

## Prerequisites

- Linux
- Node.js >= 10.15
- Docker
- Docker Compose

## How to run

### Setup

Setup the docker images according to [Preparations](../README.md#preparations) in the top page.

Install all dependencies:

```bash
$ cd fabric-opssc/integration
$ npm install
```

### Run all scenario tests

```bash
$ npm test
```

### Run the scenario tests with specifying the version of Fabric

The current OpsSC supports the latest versions of Fabric v2.3 and v2.2 series.
The integration tests uses v2.3 by default but they can also be run using v2.2 as follows:

```bash
$ curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.2.4 1.5.2 -s -b # Download Fabric images for v2.2

$ cd fabric-opssc
$ make docker FABRIC_TWO_DIGIT_VERSION=2.2 # Make images of OpsSC agent and API server for v2.2

$ cd integration
$ FABRIC_TWO_DIGIT_VERSION=2.2 npm test # Run tests using Fabric v2.2
```

### Run the specified scenario test

Can run the specified scenario test by running `npm test -- --name <SCENARIO_NAME>`

The example is:
```bash
$ npm test -- --name "Chaincode ops on docker-based Fabric network by using OpsSC"
```

### Run the scenario tests with specifying the remote chaincode repository for asset-transfer-*

In the scenario tests, asset-transfer-basic chaincode is used to test whether the chaincode and/or channel operations was successful.
The original code of this is in [fabric-samples](https://github.com/hyperledger/fabric-samples),
but by default it uses [the local files](../sample-environments/fabric-samples) to initialize
or clones this OpsSC remote repository for chaincode deployment by using OpsSC.

Can run the scenario tests including the chaincode deployment by using OpsSC with the following environment variables to specify to remote chaincode repository for asset-transfer-basic and asset-transfer-private:
- `GIT_USER=<Git repo user name used in agents>`
- `GIT_PASSWORD=<Git repo user password used in agents>`
- `IT_REMOTE_CC_REPO=<Remote repo which has asset-transfer-basic and asset-transfer-private>`
- `IT_REMOTE_BASIC_GO_CC_PATH=<Path to the source code of asset-transfer-basic in Go>`
- `IT_REMOTE_BASIC_GO_JS_PATH=<Path to the source code of asset-transfer-basic in JavaScript>`
- `IT_REMOTE_BASIC_GO_TS_PATH=<Path to the source code of asset-transfer-basic in TypeScript>`
- `IT_REMOTE_PRIVATE_GO_CC_PATH=<Path to the source code of asset-transfer-private in Go>`
- `IT_REMOTE_COMMIT_ID=<The commit ID>`

The example is:
```bash
GIT_USER= GIT_PASSWORD= IT_REMOTE_CC_REPO=github.com/hyperledger/fabric-samples IT_REMOTE_BASIC_GO_CC_PATH=asset-transfer-basic/chaincode-go IT_REMOTE_BASIC_JS_CC_PATH=asset-transfer-basic/chaincode-javascript IT_REMOTE_BASIC_TS_CC_PATH=asset-transfer-basic/chaincode-typescript IT_REMOTE_PRIVATE_GO_CC_PATH=asset-transfer-private/chaincode-go IT_REMOTE_COMMIT_ID=main npm test
```

### Preserve the test network after running the last scenario

By default, each scenario test cleans up the Fabric test network *before* and *after* it runs.
By using the environment variables `PRESERVE_TEST_NETWORK`, you can preserve the test network after running the last scenario.
This will be useful for debugging and cause analysis when the tests fail. Also, it may be useful for trial and error based on the test network.

The example is:
```bash
$ PRESERVE_TEST_NETWORK=true npm test

# You can see the preserved test network
$ docker ps
CONTAINER ID   IMAGE                                                                                                                                                                              COMMAND                  CREATED         STATUS         PORTS                              NAMES
11dd7766c30c   dev-peer0.org1.example.com-basic_0226_183125_1-691b4a17fec6ac6efb41da73b2882349145a2ab695723f398041e3c8b09ca151-179a17e1cc36054ef4d71369002e419c7c26d1e23ee3634c600971b9c20454ee   "chaincode -peer.add…"   2 minutes ago   Up 2 minutes                                      dev-peer0.org1.example.com-basic_0226_183125_1-691b4a17fec6ac6efb41da73b2882349145a2ab695723f398041e3c8b09ca151
(...)
```

Please note that only the test network of the last run scenario is preserved.
If you want to preserve the network after running the specific scenario, you also need to use `--name` option.

### (Optional.) Tear down the test network

Tear down the sample environment by using the following command:
```bash
$ ./teardownDockerEnv.sh
```
