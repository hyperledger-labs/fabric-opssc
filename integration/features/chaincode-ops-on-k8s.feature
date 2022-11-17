#
# Copyright 2022 Hitachi, Ltd., Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

@on-k8s
Feature: Chaincode ops on k8s-based Fabric network
  Background: Bootstrap a Fabric network with OpsSC on k8s
    Given bootstrap a Fabric network with CAs

    Given create ops-channel channel
    Given create mychannel channel
    Given put msp info and ccp into k8s

    Given deploy channel-ops for opssc on ops-channel
    Given deploy chaincode-ops for opssc on ops-channel

    Given register orgs info for ops-channel (type: ops) to opssc on ops-channel
    Given register orgs info for mychannel (type: application) to opssc on ops-channel

    Given bootstrap opssc-api-servers for initial orgs
    Given bootstrap opssc-agents for initial orgs
    Then 2 chaincodes should be committed on ops-channel

  Scenario: Chaincode ops on k8s-based Fabric network by using OpsSC

    # New chaincode deployment (ccaas)
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 1, channel: mychannel) based on basic ccaas template via opssc-api-server
    And org2 votes for the proposal for chaincode (name: basic, seq: 1, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 1, channel: mychannel) should be committed
    And chaincode (name: basic, seq: 1, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel) based on basic should be able to register the asset (ID: asset101) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel) based on basic ccaas should be able to get the asset (ID: asset101) by querying ReadAsset func

    # New chaincode deployment (k8s)
    When org1 requests a proposal to deploy the chaincode (name: basic-k8s, seq: 1, channel: mychannel) based on basic k8s template via opssc-api-server
    And org2 votes for the proposal for chaincode (name: basic-k8s, seq: 1, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic-k8s, seq: 1, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic-k8s, seq: 1, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic-k8s, seq: 1, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic-k8s, seq: 1, channel: mychannel) should be committed
    And chaincode (name: basic-k8s, seq: 1, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic-k8s, channel: mychannel) based on basic should be able to register the asset (ID: asset101) by invoking CreateAsset func
    And chaincode (name: basic-k8s, channel: mychannel) based on basic k8s should be able to get the asset (ID: asset101) by querying ReadAsset func

    # Chaincode update (ccaas -> k8s)
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 2, channel: mychannel) based on basic k8s template via opssc-api-server
    And org2 votes for the proposal for chaincode (name: basic, seq: 2, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be voted (with agreed) by 1 or more orgs
    And the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 2, channel: mychannel) should be committed
    And chaincode (name: basic, seq: 2, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel) based on basic should be able to register the asset (ID: asset102) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel) based on basic k8s should be able to get the asset (ID: asset102) by querying ReadAsset func

    # Chaincode update (k8s -> ccaas)
    When org1 requests a proposal to deploy the chaincode (name: basic-k8s, seq: 2, channel: mychannel) based on basic ccaas template via opssc-api-server
    And org2 votes for the proposal for chaincode (name: basic-k8s, seq: 2, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic-k8s, seq: 2, channel: mychannel) should be voted (with agreed) by 1 or more orgs
    And the proposal for chaincode (name: basic-k8s, seq: 2, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic-k8s, seq: 2, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic-k8s, seq: 2, channel: mychannel) should be committed
    And chaincode (name: basic-k8s, seq: 2, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic-k8s, channel: mychannel) based on basic should be able to register the asset (ID: asset102) by invoking CreateAsset func
    And chaincode (name: basic-k8s, channel: mychannel) based on basic ccaas should be able to get the asset (ID: asset102) by querying ReadAsset func

    # Chaincode update for one not yet deployed
    When org1 requests a proposal to deploy the chaincode (name: basic2, seq: 2, channel: mychannel) based on basic ccaas template via opssc-api-server
    And org2 votes for the proposal for chaincode (name: basic2, seq: 2, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic2, seq: 2, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic2, seq: 2, channel: mychannel) should be acknowledged (with failure) by 2 or more orgs

    # Chaincode update proposal is withdrawn
    When org1 requests a proposal to deploy the chaincode (name: basic4, seq: 1, channel: mychannel) based on basic ccaas template via opssc-api-server
    And org1 withdraws the proposal for chaincode (name: basic4, seq: 1, channel: mychannel) with opssc-api-server
    Then the proposal status for chaincode (name: basic4, seq: 1, channel: mychannel) should be withdrawn

    # Vote from each org cannot be updated
    When org1 requests a proposal to deploy the chaincode (name: basic3, seq: 1, channel: mychannel) based on basic ccaas template via opssc-api-server
    Then org1 fails to approve the proposal for chaincode (name: basic3, seq: 1, channel: mychannel) with an error (the state is already exists: Org1MSP)
    ## -- A proposal is not withdrawn with the request of anyone other than the proposer
    And org2 fails to withdraw the proposal for chaincode (name: basic3, seq: 1, channel: mychannel) with an error (only the proposer (Org1MSP) can withdraw the proposal)

    # Chaincode update rejected
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 3, channel: mychannel) based on basic ccaas template via opssc-api-server
    And org2 votes against the proposal for chaincode (name: basic, seq: 3, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 3, channel: mychannel) should be voted (with disagreed) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 3, channel: mychannel) should be rejected
    ## -- A proposal is not withdrawn after the decision
    Then org1 fails to withdraw the proposal for chaincode (name: basic, seq: 3, channel: mychannel) with an error (the voting is already closed)