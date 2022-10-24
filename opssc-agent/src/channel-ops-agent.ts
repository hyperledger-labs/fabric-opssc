/*
 * Copyright 2019-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from './logger';
import { ChannelOpsEventDetail, ChannelUpdateProposal } from 'opssc-common/opssc-types';
import { Notifier } from './notifier';
import { ChannelOperator, ChannelOperatorImpl } from './channel-operator';
import { ContractEvent, ContractListener } from 'fabric-network';
import { FabricClient } from 'opssc-common/fabric-client';
import { OpsSCAgentCoreConfig } from './config';
import { BootstrapOperatorImpl } from './bootstrap-operator';

/**
 * ChannelOpsAgent is a class to works as an OpsSC agent to operate channels.
 * An instance of the class listens the chaincode events from the Ops chaincode for operating channels,
 * and executes operations based on the events.
 *
 * <ul>
 *   <li> When the agent receives a readyToUpdateConfigEvent, this sends a request to update or create the channel to th target nodes
 *   based on the content of the proposal (if only selected as the executor) by using createChannelOperator,
 *   and then submits the result of the commit to the OpsSC chaincode. </li>
 *   <li> When the agent receives an updateConfigEvent, this update the organization nodes by using BootstrapOperator.</li>
 * </ul>
 */
export class ChannelOpsAgent {

  private readonly notifier?: Notifier;
  private readonly fabricClient: FabricClient;
  private readonly config: OpsSCAgentCoreConfig;

  /**
   * ChannelOpsAgent constructor
   *
   * @param {FabricClient} fabricClient FabricClient to interact with the OpsSC chaincodes
   * @param {OpsSCAgentCoreConfig} config the configuration used for this agent
   * @param {Notifier} [notifier] notification destination of work progress of this agent
   */
  constructor(fabricClient: FabricClient, config: OpsSCAgentCoreConfig, notifier?: Notifier) {
    this.fabricClient = fabricClient;
    this.notifier = notifier;
    this.config = config;
  }

  /**
   * Create a contract listener for the OpsSC chaincode for operating channels.
   *
   * @returns {ContractListener} the constructor listener
   */
  createContractListener(): ContractListener {
    const channelOpsSCListener = (
      async (event: ContractEvent) => {
        logger.info(`Receive channel ops event: ${event.eventName}`);
        try {
          if (event.eventName.startsWith('updateConfigEvent')) {
            this.handleUpdateConfigEvent(event);
          } else if (event.eventName.startsWith('readyToUpdateConfigEvent')) {
            this.handleReadyToUpdateConfigEvent(event);
          }
        } catch (e) {
          logger.error('Got error : %s', e.toString());
        }
      });
    return channelOpsSCListener;
  }

  /*
   * Handle a readyToUpdateConfigEvent.
   */
  private async handleReadyToUpdateConfigEvent(chaincodeEvent: { [key: string]: any }) {
    try {
      logger.debug('Chaincode event: \n%s', JSON.stringify(chaincodeEvent));
      const eventDetail = JSON.parse(chaincodeEvent.payload) as ChannelOpsEventDetail;
      logger.info('Deploy event: \n%s', JSON.stringify(eventDetail));
      const proposalID = eventDetail.proposalID;
      this.notifier?.notifyEvent('readyToUpdateConfigEvent',
        `[EVENT] Receive readyToUpdateConfigEvent (ID: ${proposalID})`, proposalID);

      if (eventDetail.operationTargets.includes(this.fabricClient.config.adminMSPID)) {
        const operator = await this.createChannelOperator(eventDetail.proposalID);
        await operator.updateConfig();
        this.notifyCommit(eventDetail.proposalID);
      }
    } catch (e) {
      logger.error(e);
    }
  }

  /*
   * Get the channel update proposal with querying to the OpsSC chaincode.
   */
  private async getProposal(proposalID: string) {
    // Get the proposal from channel-ops
    const acquisitionRequest = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.channelOpsCCName,
      func: 'GetProposal',
      args: [proposalID]
    };
    return JSON.parse(await this.fabricClient.evaluateTransaction(acquisitionRequest)) as ChannelUpdateProposal;
  }

  /*
   * Get the channel update proposal with querying to the OpsSC chaincode.
   */
  private async notifyCommit(proposalID: string) {
    // Send transaction to notify commit
    const request = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.channelOpsCCName,
      func: 'NotifyCommitResult',
      args: [proposalID]
    };
    await this.fabricClient.submitTransaction(request);
  }

  /*
   * Handle an updateConfigEvent.
   */
  private async handleUpdateConfigEvent(chaincodeEvent: { [key: string]: any }) {
    const [_, proposalID] = chaincodeEvent.eventName.split('.');
    this.notifier?.notifyEvent('updateConfigEvent', '[EVENT] Receive updateConfigEvent', proposalID);
    await this.bootstrap();
  }

  /*
   * Create a ChannelOperator instance to update or create a channel based on the proposal.
   */
  private async createChannelOperator(proposalID: string): Promise<ChannelOperator> {
    return new ChannelOperatorImpl(await this.getProposal(proposalID), this.fabricClient.config.adminMSPID, this.fabricClient.config.adminMSPConfigPath, this.fabricClient.config.connectionProfile);
  }

  /**
   * Bootstrap or update the organization nodes. This method internally uses BootstrapOperator.
   *
   * @returns {Promise<void>}
   */
  async bootstrap(): Promise<void> {
    const bootstrapOperator = new BootstrapOperatorImpl(this.fabricClient, this.config, this.notifier);
    const MAX_RETRY = 10;
    for (let retry = MAX_RETRY; retry > 0; retry--) {
      try {
        await bootstrapOperator.bootstrap();
        return;
      } catch (e) {
        let errorMessage = e;
        if (e.message != null) {
          errorMessage = e.message;
        }
        logger.error(`retry bootstrapping => retry: ${MAX_RETRY - retry}, error: ${errorMessage}`);
        await new Promise((resolve) => setTimeout(resolve, 5000));
      }
    }
    throw new Error('failed to bootstrap');
  }
}