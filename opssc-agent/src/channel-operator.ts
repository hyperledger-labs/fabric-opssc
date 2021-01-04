/*
 * Copyright 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from './logger';
import { ChannelCommands, ConfigTxProfile } from 'opssc-common/channel-commands';
import { ChannelUpdateProposal } from 'opssc-common/opssc-types';

/**
 * ChannelOperator is an interface which provides functions to execute operations to create or update a channel based on a channel update proposal.
 */
export interface ChannelOperator {

  /**
   * Execute operations to create or update a channel based on a channel update proposal.
   *
   * @returns {Promise<void>}
   */
  updateConfig(): Promise<void>;
}

/**
 * ChannelOperatorImpl is a basic implementation of ChannelOperator.
 */
export class ChannelOperatorImpl implements ChannelOperator {
  protected readonly proposal: ChannelUpdateProposal;
  protected channelCommands: ChannelCommands;

  /**
   * ChannelOperatorImpl constructor
   *
   * @param {ChannelUpdateProposal} proposal the channel update proposal
   * @param {string} mspID the MSP ID to interact with nodes to operate channels
   * @param mspConfigPath the MSP config path which has MSP to interact with nodes to operate channels
   * @param connectionProfile the connection profile that provides the necessary connection information for the client organization
   */
  constructor(proposal: ChannelUpdateProposal, mspID: string, mspConfigPath: string, connectionProfile: Record<string, any>) {
    this.proposal = proposal;
    this.channelCommands = new ChannelCommands(mspID, mspConfigPath, connectionProfile);
  }

  /**
   * Execute operations to create or update a channel based on a channel update proposal.
   *
   * @returns {Promise<void>}
   */
  async updateConfig(): Promise<void> {
    try {
      // Create envelope using the proposal
      const configTxPath = this.channelCommands.createEnvelope(this.proposal.artifacts as ConfigTxProfile);

      // Update channel config based on the proposal
      switch (this.proposal.action) {
        case 'update':
          this.channelCommands.update(this.proposal.channelID, configTxPath);
          break;
        case 'create':
          this.channelCommands.create(this.proposal.channelID, configTxPath);
          break;
        default:
          throw new Error(`Invalid action ${this.proposal.action}`);
      }
    } catch (e) {
      logger.error(e);
    } finally {
      try {
        this.channelCommands.cleanUp();
      } catch (e) {
        logger.error(`fail to clean up channelCommands: ${e}`);
      }
    }
  }
}