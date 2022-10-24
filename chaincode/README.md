# OpsSC Chaincode

An OpsSC chaincode is a chaincode that provides functions to control operational workflows and stores requests for executing them and the operational histories as states.
This also issues chaincode events including the operational instructions to administrators and agents.

Currently, there are two chaincodes, one for operating channels and one for operating chaincodes. These are written in Go.
- [channel-ops](./channel-ops): is an OpsSC chaincode for operating channels. This streamlines channel config updates across multiple organizations (creating a channel, adding an organization, an orderer etc.).
  - is based on the implementation of CMCC (Consortium Management Chaincode) in [Fabric Interop Working Group](https://wiki.hyperledger.org/display/fabric/Fabric+Interop+Working+Group)
  - provides functionalities to share channel updates and signatures between different channel members
  - provides SC functions to request a channel update proposal (that supports both creating and updating a channel), vote for the proposal by each organization  with the signature and register the status of operations to the proposal by each agent
  - provides SC functions to put / get information on channels (including the joining members) because there is currently no good way to get a list of channels
- [chaincode-ops](./chaincode-ops): is an OpsSC chaincode for operating chaincodes. This streamlines chaincode deployments with chaincode new lifecycle introduced from Fabric v2.x.
  - provides functionalities to communicate information about chaincode source code and chaincode definitions to be deployed between different channel members
  - provides SC functions to request a chaincode update proposal (that supports both deploying a new chaincode and upgrading a chaincode), vote for / against the proposal by each organization and register the status of operations to the proposal by each agent
  - internally calls the SC functions in `channel-ops` to get the information of the members of the channel that the proposed chaincode is deployed
