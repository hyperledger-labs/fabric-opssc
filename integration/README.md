# Integration Tests

## Overview

The current integration tests are scenario tests are written in typescript and use Cucumber.
The integration tests internally run (stand up and tear down) `samples-environments/fabric-samples/test-network`.

## Prerequisites

- Linux
- Node.js >= 10.15
- Docker
- Docker Compose

## How to run

### Setup

Setup the binaries and the docker images according to [Preparations](../README.md#preparations) in the top page.

Install all dependencies:

```bash
$ cd fabric-opssc/integration
$ npm install
```

### Run all scenario tests

```bash
$ npm test
```

### Run the specified scenario test

Can run the specified scenario test by running `npm test -- --name <SCENARIO_NAME>`

The example is:
```bash
$ npm test -- --name "Chaincode ops on docker-based Fabric network by using OpsSC"
```

### Run the scenario tests with specifying the remote chaincode repository for asset-transfer-basic

In the scenario tests, asset-transfer-basic chaincode is used to test whether the chaincode and/or channel operations was successful.
The original code of this is in [fabric-samples](https://github.com/hyperledger/fabric-samples),
but by default it uses [the local files](../sample-environments/fabric-samples/asset-transfer-basic) to initialize
or clones this OpsSC remote repository for chaincode deployment by using OpsSC.

Can run the scenario tests including the chaincode deployment by using OpsSC with the following environment variables to specify to remote chaincode repository for asset-transfer-basic:
- `GIT_USER=<Git repo user name used in agents>`
- `GIT_PASSWORD=<Git repo user password used in agents>`
- `IT_REMOTE_CC_REPO=<Remote repo which has asset-transfer-basic>`
- `IT_REMOTE_BASIC_GO_CC_PATH=<Path to the source code in Go>`
- `IT_REMOTE_BASIC_GO_JS_PATH=<Path to the source code in JavaScript>`
- `IT_REMOTE_BASIC_GO_TS_PATH=<Path to the source code in TypeScript>`
- `IT_REMOTE_COMMIT_ID=<The commit ID>`

The example is:
```bash
GIT_USER= GIT_PASSWORD= IT_REMOTE_CC_REPO=github.com/hyperledger/fabric-samples IT_REMOTE_BASIC_GO_CC_PATH=asset-transfer-basic/chaincode-go IT_REMOTE_BASIC_JS_CC_PATH=asset-transfer-basic/chaincode-javascript IT_REMOTE_BASIC_TS_CC_PATH=asset-transfer-basic/chaincode-typescript IT_REMOTE_COMMIT_ID=master npm test
```

### (Optional.) Tear down the test network

Tear down the sample environment by using the following command:
```bash
$ ./teardownDockerEnv.sh
```
