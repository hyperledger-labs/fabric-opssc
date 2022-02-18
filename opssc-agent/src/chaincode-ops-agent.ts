/*
 * Copyright 2019-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from './logger';
import { OpsSCAgentCoreConfig } from './config';
import { ChaincodeDeploymentEventDetail, ChaincodeUpdateProposal, TaskStatusUpdate } from 'opssc-common/opssc-types';
import { Notifier } from './notifier';
import { ChaincodeOperator, ChaincodeOperatorImpl } from './chaincode-operator';
import { ContractEvent, ContractListener } from 'fabric-network';
import { FabricClient } from 'opssc-common/fabric-client';


/**
 * ChaincodeOpsAgent is a class to works as an OpsSC agent to operate chaincodes.
 * An instance of the class listens the chaincode events from the Ops chaincode for operating chaincodes,
 * and executes operations based on the events.
 *
 * <ul>
 *   <li> When the agent receives a prepareToDeployEvent, this executes the following operations.
 *     <ul>
 *       <li> downloads the source code of the chaincode from the remote repository specified in the proposal </li>
 *       <li> packages and installs the downloaded source code </li>
 *       <li> approves the chaincode definition with the above package based on the content of the proposal </li>
 *       <li> submits the result of the above as acknowledge to the OpsSC chaincode </li>
 *     </ul>
 *   </li>
 *   <li> When the agent receives a deployEvent, this executes the following operations.
 *     <ul>
 *       <li> commits the chaincode definition based on the content of the proposal (if only selected as the executor) </li>
 *       <li> submits the result of the commit to the OpsSC chaincode </li>
 *     </ul>
 *   </li>
 * </ul>
 */
export class ChaincodeOpsAgent {

  private readonly notifier?: Notifier;
  private readonly fabricClient: FabricClient;
  private readonly config: OpsSCAgentCoreConfig;

  // Manage events in process, to avoid to Prevent duplicate execution of the same events
  private listInProcessOfPrepareToDeploy: string[];
  private listInProcessOfDeploy: string[];

  /**
   * ChaincodeOpsAgent constructor
   * @param {FabricClient} fabricClient FabricClient to interact with the OpsSC chaincodes
   * @param {OpsSCAgentCoreConfig} config the configuration used for this agent
   * @param {Notifier} [notifier] notification destination of work progress of this agent
   */
  constructor(fabricClient: FabricClient, config: OpsSCAgentCoreConfig, notifier?: Notifier) {
    this.fabricClient = fabricClient;
    this.notifier = notifier;
    this.config = config;
    this.listInProcessOfPrepareToDeploy = [];
    this.listInProcessOfDeploy = [];
  }

  /**
   * Create a contract listener for the OpsSC chaincode for operating chaincodes.
   *
   * @returns {ContractListener} the constructor listener
   */
  createContractListener(): ContractListener {
    const chaincodeOpsSCListener = (
      async (event: ContractEvent) => {
        logger.info(`Receive chaincode ops event: ${event.eventName}`);
        try {
          if (event.eventName.startsWith('prepareToDeployEvent')) {
            this.handlePrepareToDeployEvent(event);
          } else if (event.eventName.startsWith('deployEvent')) {
            this.handleDeployEvent(event);
          }
        } catch (e) {
          logger.error('Got error : %s', e.toString());
        }
      });
    return chaincodeOpsSCListener;
  }

  /*
   * Handle a prepareToDeployEvent.
   */
  async handlePrepareToDeployEvent(chaincodeEvent: { [key: string]: any }) {
    let proposalID = '';
    let skipped = false;
    try {
      logger.debug('Chaincode event: \n%s', JSON.stringify(chaincodeEvent));
      const eventDetail = JSON.parse(chaincodeEvent.payload) as ChaincodeDeploymentEventDetail;
      logger.info('Prepare to deploy event: \n%s', JSON.stringify(eventDetail));
      proposalID = eventDetail.proposal.ID;
      this.notifier?.notifyEvent('prepareToDeployEvent',
        `[EVENT] Receive prepareToDeployEvent (ID: ${proposalID})`, proposalID);

      if (eventDetail.operationTargets.includes(this.fabricClient.config.adminMSPID)) {
        if (this.listInProcessOfPrepareToDeploy.includes(proposalID)) {
          skipped = true;
          logger.warn(`Skip processing the duplicated prepareToDeployEvent (ID: ${proposalID}) because an existing operator has been processing it`);
          this.notifier?.notifyProgress(`[WARN] Skip prepareToDeployEvent (ID: ${proposalID})`, proposalID);
          return;
        }
        this.listInProcessOfPrepareToDeploy.push(proposalID);
        const operator = this.createChaincodeOperator(eventDetail.proposal);
        const history = await operator.prepareToDeploy();
        await this.registerAcknowledgeResult(history);
      }
    } catch (e) {
      logger.error(e);
    } finally {
      if (!skipped) {
        this.listInProcessOfPrepareToDeploy = this.listInProcessOfPrepareToDeploy.filter(n => n !== proposalID);
      }
      logger.info(`List in process of prepareToDeployEvent\n${this.listInProcessOfPrepareToDeploy}`);
    }
  }

  /*
   * Handle a deployEvent.
   */
  async handleDeployEvent(chaincodeEvent: { [key: string]: any }) {
    let proposalID = '';
    let skipped = false;
    try {
      logger.debug('Chaincode event: \n%s', JSON.stringify(chaincodeEvent));
      const eventDetail = JSON.parse(chaincodeEvent.payload) as ChaincodeDeploymentEventDetail;
      logger.info('Deploy event: \n%s', JSON.stringify(eventDetail));
      proposalID = eventDetail.proposal.ID;
      this.notifier?.notifyEvent('deployEvent',
        `[EVENT] Receive deploy event (ID: ${proposalID})`, proposalID);

      if (eventDetail.operationTargets.includes(this.fabricClient.config.adminMSPID)) {
        if (this.listInProcessOfDeploy.includes(proposalID)) {
          skipped = true;
          logger.warn(`Skip processing the duplicated deployEvent (ID: ${proposalID}) because an existing operator has been processing it`);
          this.notifier?.notifyProgress(`[WARN] Skip deployEvent (ID: ${proposalID})`, proposalID);
          return;
        }
        this.listInProcessOfDeploy.push(proposalID);
        const operator = this.createChaincodeOperator(eventDetail.proposal);
        const history = await operator.deploy();
        await this.registerCommitResult(history);
      }
    } catch (e) {
      logger.error(e);
    } finally {
      if (!skipped) {
        this.listInProcessOfDeploy = this.listInProcessOfDeploy.filter(n => n !== proposalID);
      }
      logger.info(`List in process of deployEvent\n${this.listInProcessOfDeploy}`);
    }
  }

  /*
   * Invoke an transaction to the OpsSC chaincode to register the result of the preparation of the chaincode deployment.
   */
  private async registerAcknowledgeResult(taskStatusUpdate: TaskStatusUpdate) {
    logger.info(`[START] Register results on ACK (proposalID ${taskStatusUpdate.proposalID})`);

    const request = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.chaincodeOpsCCName,
      func: 'Acknowledge',
      args: [JSON.stringify(taskStatusUpdate)]
    };
    await this.fabricClient.submitTransaction(request);
    logger.info(`[END] Register results on ACK (proposalID ${taskStatusUpdate.proposalID})`);
  }

  /*
   * Invoke an transaction to the OpsSC chaincode to register the commit result of the chaincode deployment.
   */
  private async registerCommitResult(taskStatusUpdate: TaskStatusUpdate) {
    logger.info(`[START] Register results on commit (proposalID ${taskStatusUpdate.proposalID})`);

    const request = {
      channelID: this.config.opssc.channelID,
      chaincodeName: this.config.opssc.chaincodes.chaincodeOpsCCName,
      func: 'NotifyCommitResult',
      args: [JSON.stringify(taskStatusUpdate)]
    };
    await this.fabricClient.submitTransaction(request);
    logger.info(`[END] Register results on commit (proposalID ${taskStatusUpdate.proposalID})`);
  }

  /*
   * Create a ChaincodeOperator instance to deploy or update a chaincode based on the proposal.
   */
  private createChaincodeOperator(proposal: ChaincodeUpdateProposal): ChaincodeOperator {

    switch (proposal.chaincodePackage.type) {
      case 'golang':
      case 'javascript':
      case 'typescript':
        return new ChaincodeOperatorImpl(this.config.ccops, proposal, this.fabricClient.getIdentity(), this.fabricClient.config.connectionProfile, this.fabricClient.config.discoverAsLocalhost, this.notifier);
      default:
        throw new Error(`Unsupported chaincode type: ${proposal.chaincodePackage.type}`);
    }
  }
}