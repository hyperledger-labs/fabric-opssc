/*
 * Copyright 2020-2021 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from './logger';
import { ChaincodeUpdateProposal, Channel } from 'opssc-common/opssc-types';
import { Notifier } from './notifier';
import { ChaincodeOperatorImpl } from './chaincode-operator';
import { ChaincodeLifecycleCommands, computePackageID } from 'opssc-common/chaincode-lifecycle-commands';
import { ChannelCommands } from 'opssc-common/channel-commands';
import { FabricClient } from 'opssc-common/fabric-client';
import { OpsSCAgentCoreConfig } from './config';

/**
 * BootstrapOperator is an interface which provides functions to bootstrap an organization's nodes
 * to make them available for the OpsSC chaincodes and the existing chaincodes.
 * It is assumed that the functions provided by the interface will be used not only during
 * the initial startup of the agent, but also when a network configuration change occurs.
 */
export interface BootstrapOperator {

  /**
   * Join all the peers for the organization to the ops channel, on which the OpsSC chaincodes works.
   *
   * @returns {Promise<void>}
   */
  joinMyPeersToOpsChannel(): Promise<void>;

  /**
   * Join all the peers for the organization to all the application channels which are managed on the OpsSC chaincodes.
   *
   * @returns {Promise<void>}
   */
  joinMyPeersToApplicationChannels(): Promise<void>;

  /**
   * Deploy the initial OpsSC chaincodes using the local files to all the peers for the organization.
   *
   * @returns {Promise<void>}
   */
  deployOpsSCOnMyPeers(): Promise<void>;

  /**
   * Deploy the existing chaincodes which are managed on the OpsSC chaincodes to all the peers for the organization.
   *
   * @returns {Promise<void>}
   */
  deployExistingChaincodesOnMyPeers(): Promise<void>;

  /**
   * Bootstrap the organization's nodes to make them available for the OpsSC chaincodes and the existing chaincodes.
   * This method should uses the above methods internally.
   *
   * @returns {Promise<void>}
   */
  bootstrap(): Promise<void>;
}

/**
 * BootstrapOperatorImpl is a basic implementation of BootstrapOperator.
 */
export class BootstrapOperatorImpl implements BootstrapOperator {

  private readonly notifier?: Notifier;
  private readonly fabricClient: FabricClient;
  private readonly channelCommands: ChannelCommands;
  private readonly mspID: string;
  private readonly config: OpsSCAgentCoreConfig;

  /**
   * BootstrapOperatorImpl constructor
   *
   * @param {FabricClient} fabricClient FabricClient to interact with the OpsSC chaincodes
   * @param {OpsSCAgentCoreConfig} config the configuration used for this operator
   * @param {Notifier} [notifier] notification destination of work progress of this operator
   */
  constructor(fabricClient: FabricClient, config: OpsSCAgentCoreConfig, notifier?: Notifier) {
    this.fabricClient = fabricClient;
    this.notifier = notifier;
    this.mspID = fabricClient.config.adminMSPID;
    this.config = config;
    this.channelCommands = new ChannelCommands(this.mspID, fabricClient.config.adminMSPConfigPath, fabricClient.config.connectionProfile);
  }

  /**
   * Bootstrap the organization's nodes to make them available for the OpsSC chaincodes and the existing chaincodes.
   *
   * @async
   * @returns {Promise<void>}
   */
  async bootstrap(): Promise<void> {
    try {
      await this.joinMyPeersToOpsChannel();
      await this.deployOpsSCOnMyPeers();
      await this.joinMyPeersToApplicationChannels();
      await this.deployExistingChaincodesOnMyPeers();
    } finally {
      try {
        this.channelCommands.cleanUp();
      } catch (e) {
        logger.error(`fail to clean up channelCommands: ${e}`);
      }
    }
  }

  /**
   * Join all the peers for the organization to all the application channels which are managed on the OpsSC chaincodes.
   *
   * @async
   * @returns {Promise<void>}
   */
  async joinMyPeersToApplicationChannels(): Promise<void> {
    logger.info('[START] Join my peers to application channel');

    const channels = await this.getAllChannels();
    for (const targetChannelInfo of channels) {
      logger.info(`Join my peers to ${targetChannelInfo.ID} (type: ${targetChannelInfo.channelType})`);
      if (targetChannelInfo.channelType !== 'application') {
        logger.info(`Skip channel ${targetChannelInfo.ID} (type: ${targetChannelInfo.channelType})`);
        continue;
      }
      if (targetChannelInfo.organizations == null || !(this.mspID in targetChannelInfo.organizations)) {
        logger.info(`Skip channel ${targetChannelInfo.ID} (This org is not a member of the channel)`);
        continue;
      }
      await this.channelCommands.joinAllPeers(targetChannelInfo.ID);
      logger.info('[END] Join my peers to application channel');
    }
  }

  /**
   * Join all the peers for the organization to the ops channel, on which the OpsSC chaincodes works.
   *
   * @async
   * @returns {Promise<void>}
   */
  async joinMyPeersToOpsChannel(): Promise<void> {
    logger.info('[START] Join my peers to ops channel');
    await this.channelCommands.joinAllPeers(this.config.opssc.channelID);
    logger.info('[END] Join my peers to ops channel');
  }

  /**
   * Deploy the initial OpsSC chaincodes using the local files to all the peers for the organization.
   *
   * @async
   * @returns {Promise<void>}
   */
  async deployOpsSCOnMyPeers(): Promise<void> {
    logger.info('[START] Deploy chaincodes for OpsSC on my peers');

    const lifecycleCommands = this.createChaincodeLifecycleCommands(this.config.opssc.channelID);
    try {
      for (const opsSCName of this.getOpsSCNames()) {
        // Get chaincode definition for an OpsSC
        logger.info(`[START] Query chaincode definition: ${opsSCName}`);
        const chaincodeDefinition = await lifecycleCommands.queryChaincodeDefinition(opsSCName);
        logger.debug('chaincodeDefinition\n%s', JSON.stringify(chaincodeDefinition));

        if (chaincodeDefinition.approvals![this.mspID] === true) {
          logger.info(`Skip deploying the chaincode ${opsSCName} (Already deployed for the org)`);
          continue;
        }

        // Make the OpsSC ready on the peers for the org based on the fetched chaincode definition
        // -- Package chaincode
        logger.info('[START] Package chaincode');
        const label = opsSCName + '_' + chaincodeDefinition.sequence.toString();
        const packageRequest = {
          lang: 'golang',
          label: label,
          chaincodePath: `/bootstrap/${opsSCName}`,
          goPath: this.getGoPath()
        };
        logger.debug('Package Request\n%s', JSON.stringify(packageRequest));
        const packagedChaincode = await lifecycleCommands.package(packageRequest);
        logger.info('[END] Package chaincode');

        // -- Install to multiple peers for an org
        logger.info('[START] Install chaincode');
        const installRequest = {
          package: packagedChaincode,
        };
        const result = await lifecycleCommands.install(installRequest);
        const packageID = result ? result : computePackageID(label, packagedChaincode);
        logger.info('[END] Install chaincode');

        // -- Approve chaincode
        logger.info('[START] Approve chaincode');
        const chaincodeRequest = {
          chaincode: {
            name: opsSCName,
            sequence: chaincodeDefinition.sequence,
            version: chaincodeDefinition.version,
            validation_parameter: chaincodeDefinition.validation_parameter,
            init_required: chaincodeDefinition.init_required,
            package_id: packageID
          }
        };
        await lifecycleCommands.approve(chaincodeRequest, false);
        logger.info('[END] Approve chaincode');
      }
    } finally {
      lifecycleCommands.close();
    }
  }

  /**
   * Deploy the existing chaincodes which are managed on the OpsSC chaincodes to all the peers for the organization.
   *
   * @returns {Promise<void>}
   */
  async deployExistingChaincodesOnMyPeers(): Promise<void> {
    logger.info('[START] Deploy existing application chaincodes on my peers');
    const channels = await this.getAllChannels();

    for (const targetChannelInfo of channels) {

      if (targetChannelInfo.channelType === 'system' || targetChannelInfo.channelType === 'disable') {
        logger.info(`Skip channel ${targetChannelInfo.ID} (type: ${targetChannelInfo.channelType})`);
        continue;
      }
      if (targetChannelInfo.organizations == null || !(this.mspID in targetChannelInfo.organizations)) {
        logger.info(`Skip channel ${targetChannelInfo.ID} (This org is not a member of the channel)`);
        continue;
      }

      const targetChannelID = targetChannelInfo.ID;
      const lifecycleCommands = this.createChaincodeLifecycleCommands(targetChannelID);

      const chaincodeDefinitions = await lifecycleCommands.queryChaincodeDefinitions();
      logger.debug('chaincodeDefinitions\n%s', JSON.stringify(chaincodeDefinitions));

      if (!chaincodeDefinitions.chaincode_definitions) {
        continue;
      }
      for (const chaincode of chaincodeDefinitions.chaincode_definitions) {
        const chaincodeName = chaincode.name!;
        const sequence = chaincode.sequence;

        if (this.getOpsSCNames().includes(chaincodeName)) {
          continue;
        }

        // Get the chaincode definition using `queryChaincodeDefinition` to get approval status.
        const chaincodeDefinition = await lifecycleCommands.queryChaincodeDefinition(chaincodeName);
        logger.debug('chaincodeDefinition\n%s', JSON.stringify(chaincodeDefinition));

        if (chaincodeDefinition.approvals![this.mspID] === true) {
          logger.info(`Skip deploying the chaincode ${chaincodeName} (Already deployed for the org)`);
          continue;
        }

        // Find the latest proposal for the chaincode on OpsSC
        const proposal = await this.findCommittedProposal(targetChannelID, chaincodeName, Number(sequence));
        if (proposal === null) {
          logger.warn(`Skip deploying the chaincode (name: ${chaincodeName}, seq:  ${sequence}) (Proposal is not found.)`);
          continue;
        }

        // Deploy the chaincode by using a chaincode operator
        const chaincodeOperator = this.createChaincodeOperator(proposal);
        await chaincodeOperator.prepareToDeploy();
      }
    }
    logger.info('[END] Deploy existing application chaincodes on my peers');
  }

  // ----- Utility functions

  protected getOpsSCNames(): string[] {
    return Object.values(this.config.opssc.chaincodes);
  }

  protected getGoPath(): string {
    return this.config.ccops.goPath || '';
  }

  private async findCommittedProposal(channelID: string, chaincodeName: string, sequence: number): Promise<ChaincodeUpdateProposal | null> {
    const request = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.chaincodeOpsCCName,
      func: 'GetAllProposals',
      args: []
    };

    const proposals = JSON.parse(await this.fabricClient.evaluateTransaction(request));

    for (const key in proposals) {
      const p = proposals[key] as ChaincodeUpdateProposal;
      if (p.channelID === channelID &&
        p.chaincodeName === chaincodeName &&
        p.chaincodeDefinition.sequence === sequence &&
        p.status === 'committed') {
        return p;
      }
    }
    return null;
  }

  private createChaincodeLifecycleCommands(channelID: string) {
    return new ChaincodeLifecycleCommands(channelID, this.fabricClient.getIdentity(), this.fabricClient.config.connectionProfile, this.fabricClient.config.discoverAsLocalhost);
  }

  private createChaincodeOperator(proposal: ChaincodeUpdateProposal) {

    switch (proposal.chaincodePackage.type) {
      case 'golang':
      case 'javascript':
      case 'typescript':
        return new ChaincodeOperatorImpl(this.config.ccops, proposal, this.fabricClient.getIdentity(), this.fabricClient.config.connectionProfile, this.fabricClient.config.discoverAsLocalhost, this.notifier);
      default:
        throw new Error(`Unsupported chaincode type: ${proposal.chaincodePackage.type}`);
    }
  }

  private async getAllChannels(): Promise<Channel[]> {
    const request = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.channelOpsCCName,
      func: 'GetAllChannels',
      args: []
    };

    return JSON.parse(await this.fabricClient.evaluateTransaction(request)) as Channel[];
  }
}
