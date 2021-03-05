#
# Copyright 2020-2021 Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

@on-docker
Feature: Channel ops on docker-based Fabric network
  Background: Bootstrap a Fabric network with OpsSC on docker
    Given download Fabric binaries
    Given bootstrap a Fabric network with CAs

    Given create mychannel channel
    Given create ops-channel channel

    Given deploy channel_ops for opssc on ops-channel
    Given deploy chaincode_ops for opssc on ops-channel

    Given register orgs info for system-channel (type: system) to opssc on ops-channel
    Given register orgs info for ops-channel (type: ops) to opssc on ops-channel
    Given register orgs info for mychannel (type: application) to opssc on ops-channel

    Given bootstrap opssc-api-servers for initial orgs
    Given bootstrap opssc-agents for initial orgs

    Then 2 chaincodes should be committed on ops-channel

  Scenario: Add a new organization on docker-based Fabric network by using OpsSC
    # Add org3 to channels

    Given prepare org3

    # -- Add org3 (including the consenter) to system-channel
    When org1 requests a proposal to add org for org3 to system-channel via opssc-api-server
    And org2 approves the proposal to add org for org3 to system-channel via opssc-api-server
    Then the proposal to add org for org3 to system-channel should be committed

    When launch nodes for org3

    # -- Add org3 (including the consenter) to ops-channel
    When org1 requests a proposal to add org for org3 to ops-channel via opssc-api-server
    And org2 approves the proposal to add org for org3 to ops-channel via opssc-api-server
    Then the proposal to add org for org3 to ops-channel should be committed

    # -- Add org3 (including the consenter) to mychannel
    When org1 requests a proposal to add org for org3 to mychannel via opssc-api-server
    And org2 approves the proposal to add org for org3 to mychannel via opssc-api-server
    Then the proposal to add org for org3 to mychannel should be committed

    # Bootstrap OpsSC agents and clients for org3
    When bootstrap opssc-api-servers for org3
    And bootstrap opssc-agents for org3
    Then 2 chaincodes should be installed on org3's peer0
    And 2 chaincodes should be installed on org3's peer1

    # New chaincode deployment
    When org3 requests a proposal to deploy the chaincode (name: basic, seq: 1, channel: mychannel) based on basic golang template via opssc-api-server
    And org1 approves the proposal for chaincode (name: basic, seq: 1, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be acknowledged (with success) by 3 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 1, channel: mychannel) should be committed
    And chaincode (name: basic, seq: 1, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel) based on basic should be able to register the asset (ID: asset100) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel) based on basic golang should be able to get the asset (ID: asset100) by querying ReadAsset func
    And 3 chaincodes should be installed on org3's peer0
    And 3 chaincodes should be installed on org3's peer1

 Scenario: Add a new organization with bootstraping on docker-based Fabric network by using OpsSC

    Given deploy basic_dummy as a dummy on mychannel
    Then 1 chaincodes should be committed on mychannel

    # New chaincode deployment
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 1, channel: mychannel) based on basic golang template via opssc-api-server
    And org2 approves the proposal for chaincode (name: basic, seq: 1, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 1, channel: mychannel) should be committed
    And chaincode (name: basic, seq: 1, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel) based on basic should be able to register the asset (ID: asset100) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel) based on basic golang should be able to get the asset (ID: asset100) by querying ReadAsset func

    # Chaincode update
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 2, channel: mychannel) based on basic golang template via opssc-api-server
    And org2 approves the proposal for chaincode (name: basic, seq: 2, channel: mychannel) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 2, channel: mychannel) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 2, channel: mychannel) should be committed
    And chaincode (name: basic, seq: 2, channel: mychannel) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel) based on basic should be able to register the asset (ID: asset101) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel) based on basic golang should be able to get the asset (ID: asset101) by querying ReadAsset func

    # Add org3 to channels

    Given prepare org3

    # -- Add org3 (including the consenter) to system-channel
    When org1 requests a proposal to add org for org3 to system-channel via opssc-api-server
    And org2 approves the proposal to add org for org3 to system-channel via opssc-api-server
    Then the proposal to add org for org3 to system-channel should be committed

    When launch nodes for org3

    # -- Add org3 (including the consenter) to ops-channel
    When org1 requests a proposal to add org for org3 to ops-channel via opssc-api-server
    And org2 approves the proposal to add org for org3 to ops-channel via opssc-api-server
    Then the proposal to add org for org3 to ops-channel should be committed

    # -- Add org3 (including the consenter) to mychannel
    When org1 requests a proposal to add org for org3 to mychannel via opssc-api-server
    And org2 approves the proposal to add org for org3 to mychannel via opssc-api-server
    Then the proposal to add org for org3 to mychannel should be committed

    # Bootstrap OpsSC agents and clients for org3
    When bootstrap opssc-api-servers for org3
    And bootstrap opssc-agents for org3
    Then 3 chaincodes should be installed on org3's peer0
    And 3 chaincodes should be installed on org3's peer1

  Scenario: Create a new channel on docker-based Fabric network by using OpsSC

    # Create mychannel2
    When org1 requests a proposal to create mychannel2 via opssc-api-server
    And org2 approves the proposal to create mychannel2 via opssc-api-server

    # New chaincode deployment
    When org1 requests a proposal to deploy the chaincode (name: basic, seq: 1, channel: mychannel2) based on basic golang template via opssc-api-server
    And org2 approves the proposal for chaincode (name: basic, seq: 1, channel: mychannel2) with opssc-api-server
    Then the proposal for chaincode (name: basic, seq: 1, channel: mychannel2) should be voted (with agreed) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel2) should be acknowledged (with success) by 2 or more orgs
    And the proposal for chaincode (name: basic, seq: 1, channel: mychannel2) should be committed (with success) by 1 or more orgs
    And the proposal status for chaincode (name: basic, seq: 1, channel: mychannel2) should be committed
    And chaincode (name: basic, seq: 1, channel: mychannel2) should be committed over the fabric network
    And chaincode (name: basic, channel: mychannel2) based on basic should be able to register the asset (ID: asset100) by invoking CreateAsset func
    And chaincode (name: basic, channel: mychannel2) based on basic golang should be able to get the asset (ID: asset100) by querying ReadAsset func
