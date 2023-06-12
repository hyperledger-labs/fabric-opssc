# Operations Smart Contract (OpsSC) for Hyperledger Fabric v2.x

## Overview

_Operations Smart Contract (OpsSC)_ is smart contract-based system operations for blockchain-based systems.
This enables decentralized system operations over multiple organizations effectively.
This repository provides a system operation tool using _the OpsSC for Hyperledger Fabric v2.x_.

In the OpsSC:

- A system operational flow is defined as a smart contract.
- The administrator of each organization interacts with the smart contract to share configuration parameters and execute operational workflow with other organizations.
- The agent for each organization automatically executes the operations (e.g., `peer` commands / administrative CLI commands) for the node according to (the event issued by) the smart contract.

This enables execution of cross-domain operations without a single point of trust (SPOT) and sharing credentials by blockchain consensus establishment.
Also it enables to execute unified procedures with unified configuration parameters based on the smart contract.

In Fabric v2.x, individual operational tasks (e.g., `peer` commands) has been refined, and SPOT is eliminated (e.g., introduced the new chaincode lifecycle).
The OpsSC for Fabric v2.x aims to enhance negotiation and (decentralized) automation capabilities to enable more efficient (typical) end-to-end operational workflows using these individual tasks and more.

As the first step for applying this to Fabric v2.x, we have developed a purpose-specific OpsSC which is essential for managing the Fabric network.

- Chaincode operations: Streamline chaincode deployment with chaincode new lifecycle introduced from v2.x.
- Channel operations: Streamline channel configuration updates across multiple organizations (e.g., creating a channel, adding an organization, adding an orderer etc.)

### Materials

Please refer to the following files to get the detail of OpsSC for Hyperledger Fabric v2.x.

- [1] [Toward Fully-Decentralized System With Hyperledger Fabric - Requirement Analysis of Decentralized System Configuration/Operations and Potential Use of a Decentralized Operations Tool "OpsSC"](https://sched.co/14H4o), Hyperledger Global Forum 2022.
- [2] [Operations Smart Contract (OpsSC) for Hyperledger Fabric v2.x](https://static.sched.com/hosted_files/hgf2021/a3/HyperledgerGlobalForum_OpsSC_pub.pdf), Hyperledger Global Forum 2021.
- [3] [Demo movies](https://github.com/satota2/fabric-opssc-materials#demo-movies)
- [4] [OpsSC: Decentralized Blockchain Network Operation Workflow for Hyperledger Fabric](https://ieeexplore.ieee.org/document/9680558), IEEE Blockchain 2021, pp.287-294, 2021.
- [5] [Operations Smart Contract to Realize Decentralized System Operations Workflow for Consortium Blockchain](https://doi.org/10.1587/transcom.2021TMP0008), IEICE Transactions on Communications (Special Section: Towards Management for Future Communications and Services in Conjunction with Main Topics of APNOMS2021), Vol.E105-B, No.11, 2022.
  - [ArXiv version](https://arxiv.org/ftp/arxiv/papers/2205/2205.04657.pdf)

### Other related materials

The other materials (the OpsSC concept etc.):

- [6] [Presentation slide for Labs proposal](https://github.com/satota2/fabric-opssc-materials/blob/main/materials/OpsSC_for_Hyperledger_Fabric_v2.x_pub.pdf), Hyperledger Fabric Contributor Meeting, 2020.
- [7] [Practical Tools for Enterprise Uses of Hyperledger Fabric (Audit and System Operations)](https://static.sched.com/hosted_files/hgf20/c4/Practical_Tools_for_Enterprise_Uses_of_Fabric.pub.pdf), Hyperledger Global Forum 2020.
  - OpsSC is presented in pp.19-24 and pp.31-38.
- [8] [Smart-Contract Based System Operations for Permissioned Blockchain](https://ieeexplore.ieee.org/document/8328745), BSC 2018, pp.1-6.
- [9] [Design and Evaluation of Smart-Contract-based System Operations for Permissioned Blockchain-based Systems](https://arxiv.org/ftp/arxiv/papers/1901/1901.11249.pdf), arXiv:1901.11249, pp.1-11, 2019.

## Feedback and contribution welcome!

- [Hyperledger Discord Chat](https://discord.com/invite/hyperledger)
  - [fabric-opssc channel](https://discord.com/channels/905194001349627914/942805501681496094)
- [GitHub Issues](https://github.com/hyperledger-labs/fabric-opssc/issues)

## Prerequisites

- Linux
- Go >= 1.14
- Node.js >= 16 (Tested by using v16.19.0)
- Docker (assuming the current stable release)
- Docker Compose (assuming the current stable release, OpsSC is tested with version 1.28.6)
- Git
- Curl (for trying the sample environment)
- jq (for trying the sample environment)

## Components

The current implementation of the OpsSC for Hyperledger Fabric v2.x consists of the following components:

- [OpsSC chaincode](./chaincode): is a chaincode that provides functions to control operational workflows and stores requests for executing them and the operational histories as states. This also issues chaincode events including the operational instructions to administrators and agents. Currently, there are two chaincodes, one for operating chaincodes and one for operating channels.

- [OpsSC API server](./opssc-api-server): is a REST API server provides an API for each organization's administrator to interact with (to invoke and/or query transactions to) the OpsSC chaincodes. In a typical use case, an administrator for an organization first requests a proposal of an operation (e.g., deploying a chaincode or updating a chanel) and then administrators for other organizations approve (vote for) the proposal.

- [OpsSC agent](./opssc-agent): is an agent program for operating nodes (e.g., peers) for each organization according to the directions from the OpsSC chaincodes. Each agent listens to chaincode events issued by OpsSC chaincodes and, automatically executes operations to components for each organization based on the events and then submits the result of the operations to the OpsSC chaincodes.

- [Fabric ConfigTX CLI](./configtx-cli): is a tiny channel / organization management CLI tool for Hyperledger Fabric v2.x. This tool outputs a new config transaction that controls channels and organizations based on some inputs. This tool internally uses [Config Transaction Library](https://github.com/hyperledger/fabric-config). This tool is internally used by the OpsSC agents and API servers.

Currently, the OpsSC chaincodes and the Fabric ConfigTX CLI are implemented in Go. On the other hand, the OpsSC agent and API server are implemented in TypeScript.
The main reason for using two languages is that the Fabric SDK Go is not yet GA at the time of development.

## Folder structures

```
(fabric-opssc)
|- chaincode/                  ... Source code for the OpsSC chaincodes (in Go)
|   |- chaincode-ops/          ... Source code for the OpsSC chaincode for operating chaincode (in Go)
|   |- channel-ops/            ... Source code for the OpsSC chaincode for operating channel (in Go)
|- common/                     ... Source code for common features for the OpsSC agent and API server (in TypeScript)
|- configtx-cli/               ... Source code for the Fabric ConfigTX CLI (in Go)
|- docs/                       ... Documents for OpsSC (e.g., Design memo)
|- integration/                ... Integration tests for the OpsSC (in TypeScript)
|- opssc-agent/                ... Source code for the OpsSC agent (in TypeScript)
|- opssc-api-server/           ... Source code for the OpsSC API server (in TypeScript)
|- sample-environments/        ... Sample environments for trying to run the OpsSC
|   |- fabric-samples/
|       |- test-network        ... Docker-based sample environment (This is based on test-network in fabric-samples. This is internally used by the integration tests.)
|   |- k8s-support/            ... K8s-based sample environment (This is based on test-network-k8s in fabric-samples.)
|- Dockerfile-for-agent        ... Docker image definition for the OpsSC agent
|- Dockerfile-for-api-server   ... Docker image definition for the OpsSC API server
|- Makefile                    ... Make file to build docker images and run the integration tests etc.
```

## Assumed Hyperledger Fabric environment

The current implementation assumes the following Fabric network:

- Hyperledger Fabric v2.5.0 or later (Tested by using v2.5.2)
  - Also it works in v2.2.2 or later (Tested by using v2.2.12)
- Fabric network configuration
  - Using Fabric CAs
  - Using Raft orderers
  - Each organization has their one or more peers and one ore more orderers
    - An orderer for each organization is required to operate channels including the system channel
    - An peer for orderer organization is required to interact with the OpsSC to execute operations on orderers
  - Having a channel for the OpsSC chaincodes (referred to "ops channel") and all organizations joins the channel
    - The OpsSC chaincodes should be deployed to all organizations
    - The OpsSC chaincodes on the ops channel is used for managing all channels and all chaincodes on the channels
    - This channel configuration is to simplify the management of the OpsSC chaincodes
  - Each organization has one or more agents and one or more API servers for that organization itself
    - The agent and API server need to use a private key and certificate for the client identity to execute admin commands to all nodes owned by that organization

## Try the OpsSC in the sample environment

This repository includes a sample environment for running the OpsSC based on [test-network in fabric-samples](https://github.com/hyperledger/fabric-samples/tree/main/test-network).
Running the OpsSC sample environment gives you a rough idea of how the OpsSC works and how to use it in a Fabric network.

_NOTE:_ This sample will collide with the original test-network in fabric-samples and destroy the environment. So, please tear down the existing test-network environment before trying the sample.

The following shows how to set up this sample environment first. After that, as a typical example of decentralized operation using the OpsSC,
it will explain a procedure for creating a new channel, deploying a new chaincode on the channel and adding a new organization using the OpsSC.

### System configuration of the test-network

The original test-network provides scripts to run a simple Fabric test network and to create channels and deploy chaincodes in centralized manner.
After deploying the test network, you can try to do decentralized operations over multiple organizations by using the OpsSC.

The test network customized for the OpsSC has the following initial configuration:

- two peer organizations (`Org1MSP` and `Org2MSP`) with one peer each
- a Raft orderer service, where each peer organization has an ordering node
- each organization has an OpsSC API server instance and an OpsSC agent instance

The customized test network has the following differences from the original version:

- Only works in an environment with Fabric CA
- Added an orderer to each peer organization (Org1MSP and Org2MSP)
- Removed the orderer organization (OrdererOrg)
- Added docker-compose files for running an OpsSC API server for each organization
- Added docker-compose files for running an OpsSC agent for each organization
- Prepared docker-compose files for organizations that will be added later (Org3MSP and Org4MSP)
- Added some utility scripts

### Preparations

#### Preparation 1: Download binaries and docker images for Hyperledger Fabric

By running the following commands, download the binaries and docker images for Hyperledger Fabric used by test-network:

```sh
$ cd ${FABRIC_OPSSC}/sample-environments/fabric-samples
$ export FABRIC_VERSION=2.5.2
$ export FABRIC_CA_VERSION=1.5.6
$ curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- -f ${FABRIC_VERSION} -c ${FABRIC_CA_VERSION} b d

$ ls bin # Confirm the target version binaries are downloaded
configtxgen  configtxlator  cryptogen  discover  fabric-ca-client  fabric-ca-server  idemixgen  orderer  osnadmin  peer
```

`${FABRIC_OPSSC}` means the `fabric-opssc` directory.

See [the official documentation](https://hyperledger-fabric.readthedocs.io/en/latest/install.html) for more details.

#### Preparation 2: Build Fabric ConfigTX CLI

Build the Fabric ConfigTX CLI by running the following commands:

```sh
$ cd ${FABRIC_OPSSC}/configtx-cli
$ make build

$ ls bin # Command to confirm the binary is created
fabric-configtx-cli
```

#### Preparation 3: Build docker images for OpsSC Agent and API Server

Build the docker images for OpsSC Agent and API Server by running the following commands:

```sh
$ cd ${FABRIC_OPSSC}
$ make docker

$ docker images # Command to confirm the images are created
REPOSITORY                                                     TAG                              IMAGE ID            CREATED             SIZE
(...)
fabric-opssc/opssc-agent                                       latest                           44e30c583566        44 hours ago        1.49GB
fabric-opssc/opssc-api-server                                  latest                           154c4a550823        44 hours ago        1.43GB
(...)
```

By default, the command builds images for Fabric v2.5 series.
If you want to build images for v2.2 series, you should set the `FABRIC_TWO_DIGIT_VERSION` variable.

### Run the test network

Launch the test network by using the following commands:

```sh
$ cd ${FABRIC_OPSSC}/sample-environments/fabric-samples/test-network
$ ./network.sh up -ca -i ${FABRIC_VERSION} -cai ${FABRIC_CA_VERSION}
```

### Initialize the OpsSC on the test network

Create `ops-channel` as the ops channel and OpsSC chaincodes for operating chaincodes and channels to the ops-channel by running the following commands:

```sh
# Create the ops channel
$ export OPS_CHANNEL_ID=ops-channel
$ ./network.sh createChannel -c ${OPS_CHANNEL_ID}

# Deploy the OpsSC chaincodes on the ops channel
$ export OPS_CC_NAME=channel-ops
$ ./network.sh deployCC -c ${OPS_CHANNEL_ID} -ccn ${OPS_CC_NAME} -ccp ../../../chaincode/${OPS_CC_NAME} -ccl go

$ export OPS_CC_NAME=chaincode-ops
$ ./network.sh deployCC -c ${OPS_CHANNEL_ID} -ccn ${OPS_CC_NAME} -ccp ../../../chaincode/${OPS_CC_NAME} -ccl go

# Add channel information (including joining organizations) for the system channel and the ops channel to the OpsSC
$ ./registerNetworkInfoToOpsSC.sh ${OPS_CHANNEL_ID} system-channel system
$ ./registerNetworkInfoToOpsSC.sh ${OPS_CHANNEL_ID} ${OPS_CHANNEL_ID} ops
```

```sh
# Launch the OpsSC agents and API servers for Org1MSP and Org2MSP
$ docker-compose -f docker/docker-compose-opssc-api-servers.yaml up -d
$ docker-compose -f docker/docker-compose-opssc-agents.yaml up -d

# Do health check for the agents and servers
## Check for the API server for Org1MSP
$ curl -X GET http://localhost:5000/healthz
OK
## Check for the API server for Org2MSP
$ curl -X GET http://localhost:5001/healthz
OK
## Check for the agent for Org1MSP
$ curl -X GET http://localhost:5500/healthz
OK
## Check for the agent for Org2MSP
$ curl -X GET http://localhost:5501/healthz
OK
```

The above commands are executed in a centralized manner.
From here on, decentralized operations over multiple organizations in the test network can be executed by using the OpsSC.

### Create a new channel using the OpsSC

To create a new channel (named `mychannel`) based on the "SampleConsortium" consortium, an administrator for `Org1MSP` sends a request for the channel update proposal to the OpsSC API server first.
In the sample environment, the API server for Org1MSP serves on port 5000.

The request is:

```sh
$ curl -X POST http://localhost:5000/api/v1/channel/proposals/create_mychannel \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "description": "Create mychannel",
    "action": "create",
    "opsProfile": {
      "Consortium": "SampleConsortium",
      "Application": {
        "Capabilities": [
          "V2_0"
        ],
        "Policies": {
          "Readers": {
            "Type": "ImplicitMeta",
            "Rule": "ANY Readers"
          },
          "Writers": {
            "Type": "ImplicitMeta",
            "Rule": "ANY Writers"
          },
          "Admins": {
            "Type": "ImplicitMeta",
            "Rule": "ANY Admins"
          },
          "LifecycleEndorsement": {
            "Type": "ImplicitMeta",
            "Rule": "MAJORITY Endorsement"
          },
          "Endorsement": {
            "Type": "ImplicitMeta",
            "Rule": "MAJORITY Endorsement"
          }
        },
        "Organizations": [
          "Org1MSP",
          "Org2MSP"
        ]
      }
    }
  }
}
EOF
"create_mychannel" # 200 OK with the proposal ID
```

Next, an administrator for `Org2MSP` confirms the contents of the proposal and votes for the proposal via the API server.
In the sample environment, the API server for Org2MSP serves on port 5001.

The command to get the proposal with the ID is the following:

```sh
$ curl -X GET http://localhost:5001/api/v1/channel/proposals/create_mychannel | jq
{
  "docType": "proposal",
  "ID": "create_mychannel",
  "channelID": "mychannel",
  "description": "Create mychannel",
  "creator": "Org1MSP",
  "status": "proposed",
  "opsProfile": {
    "Application": {
      "Capabilities": [
        "V2_0"
      ],
      "Organizations": [
        "Org1MSP",
        "Org2MSP"
      ],
(...)
```

The command to vote for the proposal is:

```sh
$ curl -X POST http://localhost:5001/api/v1/channel/proposals/create_mychannel/vote
"" # 200 OK
```

When creating a channel, it will be passed if a majority of the votes for the proposal are collected by the organizations participating in the _system_ channel.

After the proposal is passed (approved), the agents create mychannel based on the proposal and join all their peers described in the connection profile to mychannel automatically.

By using the following command, wait for the status of the proposal to be committed:

```sh
$ curl -X GET http://localhost:5001/api/v1/channel/proposals/create_mychannel | jq
{
  "docType": "proposal",
  "ID": "create_mychannel",
  "channelID": "mychannel",
  "description": "Create mychannel",
  "creator": "Org1MSP",
  "status": "committed", # Updated the status to committed
(...)
```

By using the following command, you can confirm the channel information:

```sh
$ curl -X GET http://localhost:5001/api/v1/channel/getChannels | jq
[
  {
    "docType": "channel",
    "ID": "mychannel",
    "channelType": "application",
    "organizations": {
      "Org1MSP": "",
      "Org2MSP": ""
    }
  },
  (...)
]
```

### Deploy a new chaincode by using the OpsSC

To deploy a new chaincode (`basic`) to `mychannel`, an administrator for `Org1MSP` sends a request for the chaincode update proposal to the OpsSC API server first.

These commands are:

```sh
# Convert the endorsement policy for the chaincode to base64 because the API only accepts base64 encoded endorsement policy.
$ echo -n /Channel/Application/Endorsement | base64
L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ=

# Send the request
$ curl -X POST http://localhost:5000/api/v1/chaincode/proposals/deploy_basic \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "chaincodeName": "basic",
    "chaincodePackage": {
      "repository": "github.com/hyperledger-labs/fabric-opssc",
      "pathToSourceFiles": "sample-environments/fabric-samples/asset-transfer-basic/chaincode-go",
      "commitID": "main",
      "type": "golang"
    },
    "chaincodeDefinition": {
      "sequence": 1,
      "initRequired": false,
      "validationParameter": "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="
    }
  }
}
EOF
{"docType":"proposal","ID":"deploy_basic","creator":"Org1MSP","channelID":"mychannel",(...)} # 200 OK with the requested proposal
```

Next, an administrator for `Org2MSP` confirms the contents of the proposal and votes for the proposal via the API server.

The command to get the proposal with the ID is the following:

```sh
$ curl -X GET http://localhost:5001/api/v1/chaincode/proposals/deploy_basic
{"docType":"proposal","ID":"deploy_basic","creator":"Org1MSP","channelID":"mychannel","chaincodeName":"basic", ... ,"status":"proposed",...}
```

The command to vote for the proposal is:

```sh
$ curl -X POST http://localhost:5001/api/v1/chaincode/proposals/deploy_basic/vote
"" # 200 OK
```

When deploying a chaincode, it will be passed if a majority of the votes for the proposal are collected by the organizations participating in the channel.

After the proposal is passed (approved), an agent for each organization downloads the source code of the chaincode from the remote repository specified in the proposal.
Then, the agents package and install the downloaded source code and approve and commit the chaincode definition based on the installed chaincode and the content of proposal.

By using the following command, wait for the status of the proposal to be committed:

```sh
$ curl -X GET http://localhost:5001/api/v1/chaincode/proposals/deploy_basic | jq 'select(.status == "committed")' # wait for the status to be "committed"
{
  "docType": "proposal",
  "ID": "deploy_basic",
  (...)
  "status": "committed",
  (...)
}
```

By using the following commands, can confirm that the chaincode is deployed:

```sh
# Confirm basic is installed on the organization (the following command is for Org2MSP)
$ curl -X GET 'http://localhost:5001/api/v1/chaincode/getInstalledChaincodes' | jq '.installed_chaincodes[] | select(.label == "basic_1")'
{
  "package_id": "basic_1:7cb90e2dd24972089aaac0180a5c448f3fa7bb9b5cc990d9dcb66ae414e1c027",
  "label": "basic_1",
  "references": {
    "mychannel": {
      "chaincodes": [
        {
          "name": "basic",
          "version": "1"
        }
      ]
    }
  }
}

# Confirm basic is committed on mychannel (the following command is for Org2MSP)
$ curl -X GET 'http://localhost:5001/api/v1/chaincode/queryChaincodeDefinition?channelID=mychannel&chaincodeName=basic' | jq
{
  "sequence": "1",
  "version": "1",
  "endorsement_plugin": "escc",
  "validation_plugin": "vscc",
  "validation_parameter": "EiAvQ2hhbm5lbC9BcHBsaWNhdGlvbi9FbmRvcnNlbWVudA==",
  "collections": {},
  "approvals": {
    "Org2MSP": true,
    "Org1MSP": true
  }
}
```

By using the following commands, can invoke and query the chaincode as a test:

```sh
$ curl -X POST 'http://localhost:5000/api/v1/utils/invokeTransaction' \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "channelID": "mychannel",
  "ccName": "basic",
  "func": "CreateAsset",
  "args": ["asset101", "blue", "5", "Tomoko", "300"]
}
EOF
""

$ curl -X GET 'http://localhost:5000/api/v1/utils/queryTransaction?channelID=mychannel&ccName=basic&func=GetAllAssets&args=[]' | jq
[
  {
    "ID": "asset101",
    "color": "blue",
    "size": 5,
    "owner": "Tomoko",
    "appraisedValue": 300
  }
]
```

### Add a new organization to each channel using the OpsSC

Belows are the steps to add a new organization (named `Org4MSP`), which has a peer and an ordering nodes as same as the other organizations, to each channel.
These steps are a bit more complicated than the operations described above.

First, keys and certificates for peers and orderers for `Org4MSP` should be prepared.
Here, execute the following utility script to launch a CA for `Org4MSP` and then issue keys and certificates for peers and orderers for `Org4MSP` using the CA.

```sh
# Launch a CA for Org4MSP and then issue keys and certificates for peers and orderers for Org4MSP using the CA
$ ./create-org4-artifacts-for-test-network.sh

# Convert keys and certificates (PEM files) to JSON and store them as environment variables (for simplifying the following steps)
$ export ORG4_CA_CERT=$(cat organizations/peerOrganizations/org4.example.com/msp/cacerts/localhost-13054-ca-org4.pem | sed -e ':loop; N; $!b loop; s/\n/\\n/g')
$ echo "$ORG4_CA_CERT"

$ export ORG4_TLS_CA_CERT=$(cat organizations/peerOrganizations/org4.example.com/msp/tlscacerts/ca.crt | sed -e ':loop; N; $!b loop; s/\n/\\n/g')
$ echo "$ORG4_TLS_CA_CERT"

$ export ORG4_ORDERER_SERVER_TLS_CERT=$(cat organizations/peerOrganizations/org4.example.com/orderers/orderer0.org4.example.com/tls/server.crt | sed -e ':loop; N; $!b loop; s/\n/\\n/g')
$ echo "$ORG4_ORDERER_SERVER_TLS_CERT"
```

The above will be done outside of OpsSC because the step itself does not require inter organizational negotiation.
Here, let's assume that `Org4MSP` does the above and shares the MSP information including the certificates with `Org1MSP`.

Next, an administrator for `Org1MSP` sends a request for the channel update proposal to add `Org4MSP` to system-channel to the OpsSC API server.

The request is:

```sh
$ curl -X POST http://localhost:5000/api/v1/channel/proposals/add_org4_to_system-channel \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "system-channel",
    "description": "Add org4 to system channel",
    "action": "update",
    "opsProfile": [
      {
        "Command": "set-org",
        "Parameters": {
          "OrgType": "Consortiums|Orderer",
          "Org": {
            "Name": "Org4MSP",
            "ID": "Org4MSP",
            "MSP": {
              "RootCerts": [
                "$ORG4_CA_CERT"
              ],
              "TLSRootCerts": [
                "$ORG4_TLS_CA_CERT"
              ],
              "NodeOUs": {
                "Enable": true,
                "ClientOUIdentifier": {
                  "OrganizationalUnitIdentifier": "client",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "PeerOUIdentifier": {
                  "OrganizationalUnitIdentifier": "peer",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "AdminOUIdentifier": {
                  "OrganizationalUnitIdentifier": "admin",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "OrdererOUIdentifier": {
                  "OrganizationalUnitIdentifier": "orderer",
                  "Certificate": "$ORG4_CA_CERT"
                }
              }
            },
            "Policies": {
              "Readers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.peer', 'Org4MSP.client')"
              },
              "Writers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.client')"
              },
              "OrderingReaders": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "OrderingWriters": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "Admins": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin')"
              },
              "Endorsement": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.peer')"
              }
            },
            "AnchorPeers": [
              {
                "Host": "peer0.org4.example.com",
                "Port": 13051
              }
            ],
            "OrdererEndpoints": [
              "orderer0.org4.example.com:13050"
            ]
          }
        }
      },
      {
        "Command": "set-consenter",
        "Parameters": {
          "Consenter": {
            "Host": "orderer0.org4.example.com",
            "Port": 13050,
            "ClientTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT",
            "ServerTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT"
          }
        }
      }
    ]
  }
}
EOF
"add_org4_to_system-channel" # 200 OK with the proposal ID
```

Then, an administrator for `Org2MSP` votes for the proposal via the API server:

```sh
$ curl -X POST http://localhost:5001/api/v1/channel/proposals/add_org4_to_system-channel/vote
""
```

Do the same for the remaining channels (ops-channel and mychannel):

```sh
$ curl -X POST http://localhost:5000/api/v1/channel/proposals/add_org4_to_ops-channel \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "ops-channel",
    "description": "Add org4 to ops-channel",
    "action": "update",
    "opsProfile": [
      {
        "Command": "set-org",
        "Parameters": {
          "OrgType": "Application|Orderer",
          "Org": {
            "Name": "Org4MSP",
            "ID": "Org4MSP",
            "MSP": {
              "RootCerts": [
                "$ORG4_CA_CERT"
              ],
              "TLSRootCerts": [
                "$ORG4_TLS_CA_CERT"
              ],
              "NodeOUs": {
                "Enable": true,
                "ClientOUIdentifier": {
                  "OrganizationalUnitIdentifier": "client",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "PeerOUIdentifier": {
                  "OrganizationalUnitIdentifier": "peer",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "AdminOUIdentifier": {
                  "OrganizationalUnitIdentifier": "admin",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "OrdererOUIdentifier": {
                  "OrganizationalUnitIdentifier": "orderer",
                  "Certificate": "$ORG4_CA_CERT"
                }
              }
            },
            "Policies": {
              "Readers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.peer', 'Org4MSP.client')"
              },
              "Writers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.client')"
              },
              "OrderingReaders": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "OrderingWriters": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "Admins": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin')"
              },
              "Endorsement": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.peer')"
              }
            },
            "AnchorPeers": [
              {
                "Host": "peer0.org4.example.com",
                "Port": 13051
              }
            ],
            "OrdererEndpoints": [
              "orderer0.org4.example.com:13050"
            ]
          }
        }
      },
      {
        "Command": "set-consenter",
        "Parameters": {
          "Consenter": {
            "Host": "orderer0.org4.example.com",
            "Port": 13050,
            "ClientTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT",
            "ServerTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT"
          }
        }
      }
    ]
  }
}
EOF
"add_org4_to_ops-channel" # 200 OK with the proposal ID
```

```sh
$ curl -X POST http://localhost:5001/api/v1/channel/proposals/add_org4_to_ops-channel/vote
""
```

```sh
$ curl -X POST http://localhost:5000/api/v1/channel/proposals/add_org4_to_mychannel \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "description": "Add org4 to mychannel",
    "action": "update",
    "opsProfile": [
      {
        "Command": "set-org",
        "Parameters": {
          "OrgType": "Application|Orderer",
          "Org": {
            "Name": "Org4MSP",
            "ID": "Org4MSP",
            "MSP": {
              "RootCerts": [
                "$ORG4_CA_CERT"
              ],
              "TLSRootCerts": [
                "$ORG4_TLS_CA_CERT"
              ],
              "NodeOUs": {
                "Enable": true,
                "ClientOUIdentifier": {
                  "OrganizationalUnitIdentifier": "client",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "PeerOUIdentifier": {
                  "OrganizationalUnitIdentifier": "peer",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "AdminOUIdentifier": {
                  "OrganizationalUnitIdentifier": "admin",
                  "Certificate": "$ORG4_CA_CERT"
                },
                "OrdererOUIdentifier": {
                  "OrganizationalUnitIdentifier": "orderer",
                  "Certificate": "$ORG4_CA_CERT"
                }
              }
            },
            "Policies": {
              "Readers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.peer', 'Org4MSP.client')"
              },
              "Writers": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.client')"
              },
              "OrderingReaders": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "OrderingWriters": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin', 'Org4MSP.orderer')"
              },
              "Admins": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.admin')"
              },
              "Endorsement": {
                "Type": "Signature",
                "Rule": "OR('Org4MSP.peer')"
              }
            },
            "AnchorPeers": [
              {
                "Host": "peer0.org4.example.com",
                "Port": 13051
              }
            ],
            "OrdererEndpoints": [
              "orderer0.org4.example.com:13050"
            ]
          }
        }
      },
      {
        "Command": "set-consenter",
        "Parameters": {
          "Consenter": {
            "Host": "orderer0.org4.example.com",
            "Port": 13050,
            "ClientTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT",
            "ServerTLSCert": "$ORG4_ORDERER_SERVER_TLS_CERT"
          }
        }
      }
    ]
  }
}
EOF
"add_org4_to_mychannel" # 200 OK with the proposal ID
```

```sh
$ curl -X POST http://localhost:5001/api/v1/channel/proposals/add_org4_to_mychannel/vote
""
```

By using the following command, you can see that `Org4MSP` is added to each channel:

```sh
$ curl -X GET http://localhost:5001/api/v1/channel/getChannels | jq
[
  {
    "docType": "channel",
    "ID": "mychannel",
    "channelType": "application",
    "organizations": {
      "Org1MSP": "",
      "Org2MSP": "",
      "Org4MSP": "" # Added
    }
  },
  (...)
]
```

Next, execute the following utility script to launch the peer, the ordering node, the OpsSC agent and the API server for `Org4MSP`:

```sh
# Launch the peer, the ordering node, the OpsSC agent and the API server for Org4MSP
$ ./launch-org4-nodes-for-test-network.sh
```

This script internally fetches the system config block and then launches nodes for `Org4MSP` using the system config block.
In actual use cases, an organization which already has joined in the network (`Org1MSP` or `Org2MSP`) should fetch the block and send it to `Org4MSP`.

After launching the nodes, the agent for `Org4MSP` automatically joins the organization to the channels and deploys the OpsSC chaincodes on the peer.
It also deploys the other existing chaincodes which has been deployed via OpsSC on the peer.

By running the following commands, you can confirm whether the agent and API server for `Org4MSP` get ready:

```sh
## Check for the API server for Org4MSP
$ curl -X GET http://localhost:5003/healthz
OK
## Check for the agent for Or4MSP (NOTE: Take about 2-3 minutes to be "OK")
$ curl -X GET http://localhost:5503/healthz
OK
```

By running the following commands, you can confirm that the nodes for `Org4MSP` get ready and
deployed `basic` chaincode which were deployed the previous operations.

```sh
$ export PATH=${PWD}/../bin:$PATH
$ export FABRIC_CFG_PATH=$PWD/../config/

$ export CORE_PEER_TLS_ENABLED=true
$ export CORE_PEER_LOCALMSPID="Org4MSP"
$ export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org4.example.com/peers/peer0.org4.example.com/tls/ca.crt
$ export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org4.example.com/users/Admin@org4.example.com/msp
$ export CORE_PEER_ADDRESS=localhost:13051

$ peer chaincode query -C mychannel -n basic -c '{"Args":["GetAllAssets"]}'
```

### (Optional.) Tear down the test network

You can tear down the sample environment by using the following command.

```sh
$ cd ${FABRIC_OPSSC}/integration
$ ./teardownDockerEnv.sh
```

If any of the above steps fail in the middle, reset the environment with this command and try again.

### Learn more

#### Integration test
For other more detailed usage, refer to the integration test.

- [integration tests for operating chaincodes](./integration/features/chaincode-ops-on-docker.feature)
  - e.g., deploying a chaincode, upgrading a chaincode
- [integration tests for operating channels](./integration/features/chaincode-ops-on-docker.feature)
  - e.g., adding a channel, adding an organization (with a peer and an orderer)

#### Running with k8s environments

See [the documents](./sample-environments/k8s-support/README.md)

## Limitations

The current implementation has limitations. The main limitations are as follows:

- The conditions for passing a proposal are assumed to be voted by a majority of the members of the target channel
- Does not yet support deploying Java chaincode
- Does not yet support using Channel participating API from v2.3.0

## Future work

- General operations supports: can execute arbitrary command via OpsSC chaincode
- Improving test coverage
- Porting the OpsSC API server and agent implementations from Node SDK-based to Go SDK-based (after the GA is released)

## Changes

### v0.4.0 (TBD, 2023)

- Support both Hyperledger Fabric v2.5 and v2.2 series (Out of scope: v2.4)
- Support chaincode operations on Kubernetes (K8s) environments
  - Support for chaincode operations with External Chaincode Builder for K8s
    - [CCaaS builder ('ccaas')](https://github.com/hyperledger/fabric/releases/tag/v2.4.1) support
    - [K8s chaincode builder ('k8s')](https://github.com/hyperledger-labs/fabric-builder-k8s) support
  - Helm charts to run OpsSC on K8s
    - Helm chart for OpsSC API server
    - Helm chart for OpsSC Agent
    - Helm chart for building a chaincode image and running it as a chaincode server
  - Documents and integration tests for OpsSC on K8s
    - Environment using [fabric-samples/test-network-k8s](https://github.com/hyperledger/fabric-samples/tree/main/test-network-k8s)
    - Environment using [fabric-operator/sample-network](https://github.com/hyperledger-labs/fabric-operator/tree/main/sample-network)
  - Limitations:
    - Channel operations on K8s environments are not yet supported
    - Only support Fabric 2.5+
- Other minor improvements

### v0.3.0 (Oct. 27, 2022)

- Add a voting config option for chaincode-ops - ([Detail](./docs/VotingConfigOption.md))
  - This allows OpsSC users to configure the maximum number of malicious organizations (`f`) in the voting process
  - This config option is not yet supported for channel-ops
- Add a function to update channel type and introduce `disable` channel type to channel-ops
- Support both Hyperledger Fabric v2.4 and v2.2 series (Out of scope: v2.3)
- Other minor improvements

### v0.2.0 (Oct. 5, 2021)

- Support chaincode deployment with private data for chaincode-ops
- Enhance voting functions for chaincode-ops - ([Design memo](./docs/design#design-of-state-transitions-for-operational-proposals-in-opssc))
- Support both Hyperledger Fabric v2.3 and v2.2 series
- Support system configurations when there are multiple peers in an organization

### v0.1.0 (Jan. 9, 2021)

- The initial version of Operations Smart Contract (OpsSC) for Hyperledger Fabric v2.x
  - Support chaincode operations (chaincode-ops): Streamline chaincode deployment with chaincode new lifecycle introduced from v2.x.
  - Support channel operations (channel-ops): Streamline channel configuration updates across multiple organizations (e.g., creating a channel, adding an organization, adding an orderer etc.)

## Authors

- Tatsuya Sato
- Taku Shimosawa

## License

Apache-2.0 (See [LICENSE](LICENSE))
