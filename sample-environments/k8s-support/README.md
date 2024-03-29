# K8s-based sample environments

This introduces sample environments for running the OpsSC on K8s.
- Sample 1: Environment with [test-network-k8s in fabric-samples](https://github.com/hyperledger/fabric-samples/tree/main/test-network-k8s).
  - See [here](./README.md#try-the-opssc-in-a-k8s-based-sample-environment-with-test-network-k8s)
- Sample 2: Environment with [sample-network in fabric-operator](https://github.com/hyperledger-labs/fabric-operator/tree/main/sample-network).
  - See [here](./README.md#try-the-opssc-in-a-k8s-based-sample-environment-with-fabric-operator)

## Limitations

OpsSC on k8s has the following limitations.

- Only support Fabric 2.5+
- Only support `chaincode-ops`
  - `channel-ops` is not supported

## Try the OpsSC in a k8s-based sample environment with test-network-k8s

### Prerequisites

- Setup should be complete [Preparation 2](../../README.md#preparation-2-build-fabric-configtx-cli) and [Preparation 3](../../README.md#preparation-3-build-docker-images-for-opssc-agent-and-api-server)
- Additional required software
  - kubectl
  - helm
  - KIND

### Prepare test-network-k8s

Clone fabric-samples to use test-network-k8s by running the following commands:

```bash
cd ${FABRIC_OPSSC}/sample-environments/k8s-support

# Remove old clone
rm -rf fabric-samples

# Clone fabric-samples
git clone https://github.com/hyperledger/fabric-samples.git

```

_NOTE:_ The following instructions are tested with commit `c323c9580717dd6376e5e2b07ecbb00af5b3bf00` in fabric-samples. Other commits may require some modifications.

To run test-network-k8s with OpsSC, need to make some updates to test-network-k8s:

- Add operations for installing and approving a chaincode for org2
  - The original implementation does them only for org1
- Modify channel.sh to avoid an error when network channel create is called multiple times

Apply patches to test-network to make the above updates by running the following commands:

```bash
# Add operations for installing and approving a chaincode for org2 to test-network-k8s
patch -u fabric-samples/test-network-k8s/scripts/chaincode.sh < patches/test-network-k8s/chaincode.sh.patch

# Temporary repairs to avoid an error when network channel create is called multiple times
patch -u fabric-samples/test-network-k8s/scripts/channel.sh < patches/test-network-k8s/channel.sh.patch
```

### Run the test network on a KIND cluster

Before running, clean up the test network:

```bash
# Move test-network-k8s dir
cd fabric-samples/test-network-k8s

# Clean up test-network-k8s
./network down
./network cluster clean
./network unkind

```

Create a KIND cluster:

```bash
# Create and initialize a KIND cluster
./network kind
./network cluster init

```

By the default, a Fabric network is deployed in a namespace `test-network` and a channel named `mychannel` is created.
In the following instructions, two channels `mychannel` and `ops-channel` are created.
`mychannel` is an application channel and `ops-channel` is the channel on which the OpsSC chaincodes runs.

Launch the test network and create channels:

```bash
# Set the namespace name
export TEST_NETWORK_NETWORK_NAME="test-network"

# Set the builder as 'k8s', to setup k8s builder to use it the following instructions
export TEST_NETWORK_CHAINCODE_BUILDER="k8s"

./network up # Internally setup the k8s builder

export TEST_NETWORK_CHANNEL_NAME="mychannel"
./network channel create # create mychannel

export TEST_NETWORK_CHANNEL_NAME="ops-channel"
./network channel create # create ops-channel

```
### Initialize the OpsSC on the test network

Deploy OpsSC chaincodes as chaincode servers and set up the initial chaincode info:

```bash
# Deploy OpsSC chaincodes via k8s or ccaas chaincode builder
export TEST_NETWORK_CHAINCODE_BUILDER="k8s" # You can also use "ccaas"
./network chaincode deploy channel-ops ../../../../chaincode/channel-ops
./network chaincode deploy chaincode-ops ../../../../chaincode/chaincode-ops

# Put initial channel info into OpsSC chaincodes
./network chaincode invoke channel-ops '{"Args":["CreateChannel","ops-channel","ops","[]"]}'
./network chaincode invoke channel-ops '{"Args":["AddOrganization","ops-channel","Org1MSP"]}'
./network chaincode invoke channel-ops '{"Args":["AddOrganization","ops-channel","Org2MSP"]}'

./network chaincode invoke channel-ops '{"Args":["CreateChannel","mychannel","application","[]"]}'
./network chaincode invoke channel-ops '{"Args":["AddOrganization","mychannel","Org1MSP"]}'
./network chaincode invoke channel-ops '{"Args":["AddOrganization","mychannel","Org2MSP"]}'

```

Deploy OpsSC Agent and API Server for each org on k8s by using helm.

The helm charts used are respectively as follows:

- [OpsSC API Server](../../../../opssc-api-server/charts/opssc-api-server)
  - [Sample values for org1](./helm_values/test-network-k8s/org1-opssc-api-server.yaml)
- [OpsSC Agent](../../../../opssc-agent/charts/opssc-agent)
  - [Sample values for org1](./helm_values/test-network-k8s/org1-opssc-agent.yaml)

To deploy OpsSC Agent and API Server for each org, run the following commands:

```bash
# Load Docker images of OpsSC API Server and Agent from local to KIND
kind load docker-image fabric-opssc/opssc-api-server:latest
kind load docker-image fabric-opssc/opssc-agent:latest

# Create connection profiles with peer and orderer information and put them into K8s as ConfigMap
../../utils/create_ccp_comfigmap.sh

# Put admin MSP info for each org into K8s
ls build/enrollments/org1/users/org1admin
tar -C build/enrollments/org1/users/org1admin -cvf build/admin-msp.tar msp
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete configmap org1-admin-msp || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create configmap org1-admin-msp --from-file=build/admin-msp.tar

ls build/enrollments/org2/users/org2admin
tar -C build/enrollments/org2/users/org2admin -cvf build/admin-msp.tar msp
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete configmap org2-admin-msp || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create configmap org2-admin-msp --from-file=build/admin-msp.tar

rm build/admin-msp.tar

# Deploy OpsSC Agent and API server for each org on K8s by using helm
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org1-opssc-api-server ../../../../opssc-api-server/charts/opssc-api-server -f ../../helm_values/test-network-k8s/org1-opssc-api-server.yaml --install
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org2-opssc-api-server ../../../../opssc-api-server/charts/opssc-api-server -f ../../helm_values/test-network-k8s/org2-opssc-api-server.yaml --install

helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org1-opssc-agent ../../../../opssc-agent/charts/opssc-agent -f ../../helm_values/test-network-k8s/org1-opssc-agent.yaml --install
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org2-opssc-agent ../../../../opssc-agent/charts/opssc-agent -f ../../helm_values/test-network-k8s/org2-opssc-agent.yaml --install

```

Each chart reads a connection profile with peer and orderer information put as ConfigMap.
So, the above calls an utility script to create connection profiles and put them into K8s as ConfigMap.

Each agent or API server exposes a `Ingress` route binding the virtual host name to the corresponding endpoint.
External clients can access each agent or API server with `*.localho.st` domain (For more information, see [here](https://github.com/hyperledger/fabric-samples/tree/main/test-network-k8s/docs/KUBERNETES.md)).

Do health check for the agents and servers with `*.localho.st` domain:

```bash
# Do health check for the agents and servers
## Check for the API server for Org1MSP
curl -X GET http://org1-opssc-api-server.localho.st/healthz
# OK

## Check for the API server for Org2MSP
curl -X GET http://org2-opssc-api-server.localho.st/healthz
# OK

## Check for the agent for Org1MSP
curl -X GET http://org1-opssc-agent.localho.st/healthz
# OK

## Check for the agent for Org2MSP
curl -X GET http://org2-opssc-agent.localho.st/healthz
# OK
```

_NOTE_: Since it is difficult to input a folder structure into k8s resources like ConfigMap,
as a workaround, the helm charts take an Admin MSP folder as a compressed file with tar format and decompresses that file, and uses it.

### Deploy a new chaincode by using the OpsSC

For k8s environment, the OpsSC supports both [CCaaS builder ('ccaas')](https://github.com/hyperledger/fabric/releases/tag/v2.4.1) and [K8s chaincode builder ('k8s')](https://github.com/hyperledger-labs/fabric-builder-k8s).

Simply specify `chaincodePackage.type` with `ccaas` or `k8s` in the chaincode proposal, then the OpsSC agents deploy the proposed chaincode with the specified builder.
The OpsSC agents (or the peers for k8s builder) internally build the Docker image of the proposed chaincode,
publish it to the specified registry, and launch the chaincode server as a pod by accessing the k8s cluster with helm and/or kubectl.

#### Deploy via the CCaaS builder

To deploy a new chaincode (`basic`) to `mychannel`, an administrator for `Org1MSP` sends a request for the chaincode update proposal to the OpsSC API server first.

Next, an administrator for `Org2MSP` confirms the contents of the proposal and votes for the proposal via the API server.

```bash
# Set Chaincode type as 'ccaas' to use the CCaaS builder
export CC_TYPE=ccaas

# Set Proposal ID
export PROPOSAL_ID=${CC_TYPE}_01

# Send the request
curl -X POST http://org1-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID} \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "chaincodeName": "basic",
    "chaincodePackage": {
      "repository": "github.com/hyperledger/fabric-samples",
      "pathToSourceFiles": "asset-transfer-basic/chaincode-java",
      "commitID": "main",
      "type": "${CC_TYPE}"
    },
    "chaincodeDefinition": {
      "sequence": 1,
      "initRequired": false,
      "validationParameter": "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="
    }
  }
}
EOF
# {"docType":"proposal" (...)} -- 200 OK with the requested proposal

# Vote for the proposal
curl -X POST http://org2-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID}/vote
# "" -- 200 OK

```

By using the following command, wait for the status of the proposal to be committed (may take 2-3 minutes):

```bash
curl -X GET http://org1-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID} | jq 'select(.status == "committed")' # wait for the status to be "committed"

```

By using the following commands, can invoke and query the chaincode as a test:

```bash
export TEST_NETWORK_CHANNEL_NAME="mychannel"
./network chaincode invoke basic '{"Args":["CreateAsset","asset1","blue","35","tom","1000"]}'
./network chaincode query basic '{"Args":["ReadAsset","asset1"]}'

```

By using the following commands, you can see pods for the deployed chaincode:

```bash
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" get pods | grep basic

# You can get the result as follow
chaincode-basic-org1-84d79f4667-bfchw            1/1     Running     0          3m5s
chaincode-basic-org1-buildjob-q55jz              0/1     Completed   0          5m2s
chaincode-basic-org2-6f5dcbc749-5d7rm            1/1     Running     0          3m5s
chaincode-basic-org2-buildjob-64wc5              0/1     Completed   0          5m2s
```

In the current implementation for 'ccaas' builder, In the current implementation,
a chaincode server is created per org1 and per chaincode not per peer.

#### Deploy via the k8s builder

Simply change the chaincode type in the above instructions for CCaaS builder from `ccaas` to `k8s` to deploy the chaincode with k8s builder. The following is a partial example of deploying a chaincode named `basic2`.

```bash
# Set the channel as 'ops-channel'
export TEST_NETWORK_CHANNEL_NAME="ops-channel"

# Set Chaincode type as 'k8s' to use the k8s builder
export CC_TYPE=k8s

# Set Proposal ID
export PROPOSAL_ID=${CC_TYPE}_01

# Send the request
curl -X POST http://org1-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID} \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "chaincodeName": "basic2",
    "chaincodePackage": {
      "repository": "github.com/hyperledger/fabric-samples",
      "pathToSourceFiles": "asset-transfer-basic/chaincode-java",
      "commitID": "main",
      "type": "${CC_TYPE}"
    },
    "chaincodeDefinition": {
      "sequence": 1,
      "initRequired": false,
      "validationParameter": "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="
    }
  }
}
EOF
# {"docType":"proposal" (...)} -- 200 OK with the requested proposal

# Vote for the proposal
curl -X POST http://org2-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID}/vote
# "" -- 200 OK

```

By using the following commands, you can see pods for the deployed chaincode:

```bash
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" get pods | grep basic2

# You can get the result as follow
cc-org1msp-org1-peer1.org1.example.combasic2-1-a0fefbefdf85c51b   1/1     Running     0          40s
cc-org1msp-org1-peer2.org1.example.combasic2-1-a0fefbefdf85c51b   1/1     Running     0          40s
cc-org2msp-org2-peer1.org2.example.combasic2-1-5b3968c9f6cf393f   1/1     Running     0          40s
cc-org2msp-org2-peer2.org2.example.combasic2-1-5b3968c9f6cf393f   1/1     Running     0          40s
chaincode-basic2-org1-buildjob-s5rcd                              0/1     Completed   0          2m44s
chaincode-basic2-org2-buildjob-qdgbt                              0/1     Completed   0          2m44s
```

### Use your external container image registry

By default, test-network-k8s sets up a local container registry and uses it to push/pull chaincode images.
Instead, any external container registry can be used.

The following describes an example of using [Amazon Elastic Container Registry (ECR)](https://aws.amazon.com/ecr/).
Here assume that the ECR setup has been completed.

When using an external registry, OpsSC agents must have credential information about the registry
to push and pull the proposed chaincode images with the registry via Helm.

Before deploying OpsSC agents on K8s by using helm,
rewrite the helm values about image registry [in this folder](./helm_values/test-network-k8s):

```yaml
# For your private container image registry
# <your_private_container_image_registry_name>/<chaincode_id> is used.
registry: <your_your_private_container_image_registry_name>
pullRegistryOverride: ""
imagePullSecretName: "docker-secret"
```

Next, login the registry and put that credential information as a `Secret` to k8s:

```bash
# Example of using ECR
export REGISTRY_HOST=<your_account_id>.dkr.ecr.<region>.amazonaws.com
export REGISTRY_USER=AWS
export REGISTRY_PASSWORD=$(aws ecr get-login-password)

kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete secret docker-secret || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create secret docker-registry docker-secret --docker-server="${REGISTRY_HOST}" --docker-username="${REGISTRY_USER}" --docker-password="${REGISTRY_PASSWORD}"
```

Then, perform the same steps after deploying the OpsSC agents as described above.

#### When using k8s builder and an external registry

In k8s builder, peers pull the proposed chaincode images from the external registry and run them.
So, credential information of the external registry must be registered in the service account used by the peers.

- [Related issue](https://github.com/hyperledger-labs/fabric-builder-k8s/issues/65)
- [How to configure in k8s official docs](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account)

By executing the following command, credential information can be added to the service account in test-network-k8s:

```bash
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" patch serviceaccount default -p '{"imagePullSecrets": [{"name": "docker-secret"}]}'

```

### Use your private git repository

OpsSC supports to download the chaincode source code from a private git repository.
The following instructions explains an example of how to use a private git repository when running OpsSC on k8s.

The aforementioned OpsSC agent helm chart reads a git credential information put as `Secret`.
Set your credential information as follows:

```bash
export GIT_USER=<put_your_git_username>
export GIT_PASSWORD=<put_your_git_password>

kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete secret git || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create secret generic git --from-literal=username="${GIT_USER}" --from-literal=password="${GIT_PASSWORD}"
```

After setting up the above, start OpsSC agents.
Then, when you send a proposal specifying your private git repository as follows,
the agents will clone the chaincode source from the repository using the git user and password you set in `Secret` above.

```bash
# Send the request
curl -X POST http://org1-opssc-api-server.localho.st/api/v1/chaincode/proposals/deploy_basic_${PROPOSAL_ID} \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "proposal": {
    "channelID": "mychannel",
    "chaincodeName": "basic",
    "chaincodePackage": {
      "repository": "<put_your_private_git_repository>",
      "pathToSourceFiles": "asset-transfer-basic/chaincode-java",
      "commitID": "main",
  (...)
}
EOF
```

## Try the OpsSC in a k8s-based sample environment with fabric-operator

The following instructions describe how to try the OpsSC with [fabric-operator](https://github.com/hyperledger-labs/fabric-operator).
fabric-operator is an open-sourced Operator implementation, which follows the CNCF Operator Pattern, for managing Fabric networks on K8s.
`sample-network` in `fabric-operator` repository is an environment similar to `test-network-k8s` (in other words, `fabric-operator` version of `test-network-k8s`).

Note that in the following instructions, the same parts as in `test-network-k8s` are omitted.

### Prerequisites

Same as the example with `test-network-k8s`.
See [here](./README.md#try-the-opssc-in-a-k8s-based-sample-environment-with-test-network-k8s).

### Prepare fabric-operator

Clone fabric-operator by running the following commands:

```bash
cd ${FABRIC_OPSSC}/sample-environments/k8s-support

# Remove old clone
rm -rf fabric-operator

# Clone fabric-operator
git clone https://github.com/hyperledger-labs/fabric-operator.git

```

_NOTE:_ The following instructions are tested with commit `4b222fe2909803e94241474c1f05d5321b02bfc0` in fabric-operator. Other commits may require some modifications.

To run sample-network in fabric-operator with OpsSC, need to make some updates to sample-network:

- Add operations for installing and approving a chaincode for org2
  - The original implementation does them only for org1

Apply patches to fabric-operator to make the above updates by running the following commands:

```bash
# Add operations for installing and approving a chaincode for org2 to sample-network
patch -u fabric-operator/sample-network/scripts/chaincode.sh < patches/fabric-operator/chaincode.sh.patch

```

### Run the sample-network on a KIND cluster

Before running, clean up the sample network:

```bash
# Move sample-network dir
cd fabric-operator/sample-network

# Clean up sample-network
./network down
./network cluster clean
./network unkind

```

Create a KIND cluster:

```bash
# Create and initialize a KIND cluster
./network kind
./network cluster init

```

By the default, a Fabric network is deployed in a namespace `test-network` and a channel named `mychannel` is created.
In the following instructions, two channels `mychannel` and `ops-channel` are created.
`mychannel` is an application channel and `ops-channel` is the channel on which the OpsSC chaincodes runs.

Launch the test network and create channels:

```bash
# Set the namespace name
export TEST_NETWORK_NETWORK_NAME="test-network"

# Set the environment variables, to use a peer image with k8s builder to use it the following instructions
export TEST_NETWORK_PEER_IMAGE=ghcr.io/hyperledger-labs/k8s-fabric-peer
export TEST_NETWORK_PEER_IMAGE_LABEL=v0.8.0

./network up

export TEST_NETWORK_CHANNEL_NAME="mychannel"
./network channel create # create mychannel

export TEST_NETWORK_CHANNEL_NAME="ops-channel"
./network channel create # create ops-channel

```
### Initialize the OpsSC on the test network

Deploy OpsSC chaincodes as chaincode servers and set up the initial chaincode info:

```bash
# Deploy OpsSC chaincodes
TEST_NETWORK_CHAINCODE_IMAGE=chaincode/channel-ops ./network cc deploy channel-ops channel-ops_1.0 ../../../../chaincode/channel-ops
TEST_NETWORK_CHAINCODE_IMAGE=chaincode/chaincode-ops ./network cc deploy chaincode-ops chaincode-ops_1.0 ../../../../chaincode/chaincode-ops

# Put initial channel info into OpsSC chaincodes
./network cc invoke channel-ops '{"Args":["CreateChannel","ops-channel","ops","[]"]}'
./network cc invoke channel-ops '{"Args":["AddOrganization","ops-channel","Org1MSP"]}'
./network cc invoke channel-ops '{"Args":["AddOrganization","ops-channel","Org2MSP"]}'

./network cc invoke channel-ops '{"Args":["CreateChannel","mychannel","application","[]"]}'
./network cc invoke channel-ops '{"Args":["AddOrganization","mychannel","Org1MSP"]}'
./network cc invoke channel-ops '{"Args":["AddOrganization","mychannel","Org2MSP"]}'

```

Deploy OpsSC Agent and API Server for each org on k8s by using helm.

The helm values used are respectively as follows:

- [OpsSC API Server](../../../../opssc-api-server/charts/opssc-api-server)
  - [Sample values for org1](./helm_values/fabric-operator/org1-opssc-api-server.yaml)
- [OpsSC Agent](../../../../opssc-agent/charts/opssc-agent)
  - [Sample values for org1](./helm_values/fabric-operator/org1-opssc-agent.yaml)

To deploy OpsSC Agent and API Server for each org, run the following commands:

```bash
# Load Docker images of OpsSC API Server and Agent from local to KIND
kind load docker-image fabric-opssc/opssc-api-server:latest
kind load docker-image fabric-opssc/opssc-agent:latest

# Create connection profiles and put them into K8s as ConfigMap
export TEST_NETWORK_SAMPLE_ENV_NAME=fabric-operator
export TEST_NETWORK_TEMP_DIR=${PWD}/temp
../../utils/create_ccp_comfigmap.sh

# Put admin MSP info for each org into K8s
ls temp/enrollments/org1/users/org1admin
tar -C temp/enrollments/org1/users/org1admin -cvf temp/admin-msp.tar msp
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete configmap org1-admin-msp || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create configmap org1-admin-msp --from-file=temp/admin-msp.tar

ls temp/enrollments/org2/users/org2admin
tar -C temp/enrollments/org2/users/org2admin -cvf temp/admin-msp.tar msp
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" delete configmap org2-admin-msp || true
kubectl -n "${TEST_NETWORK_NETWORK_NAME}" create configmap org2-admin-msp --from-file=temp/admin-msp.tar

rm temp/admin-msp.tar

# Deploy OpsSC Agent and API server for each org on K8s by using helm
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org1-opssc-api-server ../../../../opssc-api-server/charts/opssc-api-server -f ../../helm_values/fabric-operator/org1-opssc-api-server.yaml --install
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org2-opssc-api-server ../../../../opssc-api-server/charts/opssc-api-server -f ../../helm_values/fabric-operator/org2-opssc-api-server.yaml --install

helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org1-opssc-agent ../../../../opssc-agent/charts/opssc-agent -f ../../helm_values/fabric-operator/org1-opssc-agent.yaml --install
helm upgrade -n "${TEST_NETWORK_NETWORK_NAME}" org2-opssc-agent ../../../../opssc-agent/charts/opssc-agent -f ../../helm_values/fabric-operator/org2-opssc-agent.yaml --install

```

Do health check for the agents and servers with `*.localho.st` domain:

```bash
# Do health check for the agents and servers
## Check for the API server for Org1MSP
curl -X GET http://org1-opssc-api-server.localho.st/healthz
# OK

## Check for the API server for Org2MSP
curl -X GET http://org2-opssc-api-server.localho.st/healthz
# OK

## Check for the agent for Org1MSP
curl -X GET http://org1-opssc-agent.localho.st/healthz
# OK

## Check for the agent for Org2MSP
curl -X GET http://org2-opssc-agent.localho.st/healthz
# OK
```

### Deploy a new chaincode by using the OpsSC

Same as the example with `test-network-k8s`.
See [here](./README.md#deploy-a-new-chaincode-by-using-the-opssc).

NOTE: `./network chaincode` should be replaced `./network cc` in the instructions for `fabric-samples/test-network-k8s`.
