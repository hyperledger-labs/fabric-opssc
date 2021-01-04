/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Contract, ContractListener, DefaultCheckpointers, Gateway, GatewayOptions, Identity, X509Identity } from 'fabric-network';
import fs from 'fs-extra';
import * as path from 'path';
import { logger } from './logger';

export type TransactionRequest = {
  channelID: string;
  chaincodeName: string;
  func: string;
  args: string[];
  retry?: number;
}

export interface FabricConfig {
  adminMSPID: string;
  adminKey: string;
  adminCert: string;
  adminMSPConfigPath: string;
  connectionProfile: Record<string, any>;
  discoverAsLocalhost: boolean;
}

export interface RegisteredContractListener {
  channelID: string;
  chaincodeName: string;
  listener: ContractListener;
}

const DEFAULT_RETRY_MAX = 10;

/**
 * <p> FabricClient is a class to invoke, query chaincodes. Also, this manages contract listeners.
 * This class internally uses Gateway in fabric-network.
 * </p>
 *
 */
/*
 * NOTE: Currently, FabricClient does not support getting Identity using FabricCAServices
 * because peer channel commands use the MSP but there is no way to restore config.yaml from the capability of FabricCAServices.
 */
export class FabricClient {
  public readonly config: FabricConfig;
  private gateway: Gateway | null;

  // Save the registered listeners because the events will not be listened to even if the gateway is reconnected once the gateway is disconnected.
  // When the gateway is reset, the FabricClient automatically add the listeners again.
  private registeredListeners: {
    [key: string]: RegisteredContractListener;
  }

  /**
   * FabricClient constructor
   *
   * @param {FabricConfig} config the configuration to interact with chaincodes
   */
  constructor(config: FabricConfig) {
    this.config = config;
    this.gateway = null;
    this.registeredListeners = {};
  }

  /**
   * Invoke a chaincode. This method will automatically retry if the invocation fails after reset the connection.
   *
   * @async
   * @param {TransactionRequest} request the request to invoke a chaincode
   * @returns {Promise<string>} the result of the invocation
   */
  async submitTransaction(request: TransactionRequest): Promise<string> {
    for (let retry = (request.retry ? request.retry : DEFAULT_RETRY_MAX); retry > 0; retry--) {
      try {
        const contract = await this.getContract(request.channelID, request.chaincodeName);
        const submitResult = await contract.submitTransaction(request.func, ...request.args);
        return submitResult.toString('utf-8');
      } catch (error) {
        if (error.message != null) {
          if ((error.message as string).includes('MVCC_READ_CONFLICT') || (error.message as string).includes('PHANTOM_READ_CONFLICT')
            || (error.message as string).includes('ENDORSEMENT_POLICY_FAILURE')) { // As workaround: Considering failure when updating organization
            // Retry
            await new Promise((resolve) => setTimeout(resolve, Math.random() * 800));
            logger.info('Retry transaction: %s', error.message);
            this.close(); // To update service discovery results
            continue;
          }
        }
        logger.error('Failed:', JSON.stringify(error));
        throw error;
      }
    }
    throw new Error('Failed too many times');
  }

  /**
   * Query a chaincode.
   *
   * @async
   * @param {TransactionRequest} request the request to query a chaincode
   * @returns {Promise<string>} the result of the query
   */
  async evaluateTransaction(request: TransactionRequest): Promise<string> {
    const contract = await this.getContract(request.channelID, request.chaincodeName);
    const result = await contract.evaluateTransaction(request.func, ...request.args);
    return result.toString('utf-8');
  }

  /**
   * Return X509 identity with the client identity based on the FabricConfig.
   *
   * @returns {Identity} the identity
   */
  getIdentity(): Identity {
    return {
      credentials: {
        certificate: this.config.adminCert,
        privateKey: this.config.adminKey,
      },
      mspId: this.config.adminMSPID,
      type: 'X.509',
    } as X509Identity;
  }

  /**
   * Close the Fabric network connection for the FabricClient.
   */
  close() {
    if (this.gateway == null) {
      return;
    }
    this.gateway.disconnect();
    this.gateway = null;
  }

  /*
   * Internal method to return Contract object to interact with the given chaincode on the given channel.
   */
  private async getContract(channelID: string, chaincodeName: string): Promise<Contract> {
    await this.prepareGateway();

    const network = await this.gateway!.getNetwork(channelID);
    return network.getContract(chaincodeName);
  }

  /**
   * Add a contract event listener. The added listeners are automatically set up even if the connection is reset.
   *
   * @async
   * @param {string} channelID the target channelID to which the listener is added
   * @param {string} chaincodeName the target chaincodeName to which the listener is added
   * @param {ContractListener} listener the listener which is added
   * @returns {Promise<void>}
   */
  async addContractEventListener(channelID: string, chaincodeName: string, listener: ContractListener): Promise<void> {
    const listenerID = await this.addContractEventListenerInternal(channelID, chaincodeName, listener);

    this.registeredListeners[listenerID] = {
      channelID: channelID,
      chaincodeName: chaincodeName,
      listener: listener
    };
  }

  /*
   * Internal method to add a contract event listener.
   */
  private async addContractEventListenerInternal(channelID: string, chaincodeName: string, listener: ContractListener): Promise<string> {
    const contract = await this.getContract(channelID, chaincodeName);
    const listenerID = `${channelID}_${chaincodeName}`;
    const checkPointFolderPath = path.join('/', 'opt', 'opssc', 'data', 'checkpoint');
    fs.mkdirpSync(checkPointFolderPath);
    const checkpointer = await DefaultCheckpointers.file(path.join(checkPointFolderPath, `${listenerID}.json`));
    await contract.addContractListener(listener,
      {
        startBlock: await checkpointer.getBlockNumber(),
        checkpointer: checkpointer
      });
    return listenerID;
  }

  /*
   * Internal method to add the event listeners again for when the connection is reset.
   */
  private async setupAllRegisteredContractEventListeners() {
    for (const listenerID in this.registeredListeners) {
      const listener = this.registeredListeners[listenerID];
      await this.addContractEventListenerInternal(listener.channelID, listener.chaincodeName, listener.listener);
    }
  }

  /*
   * Internal method to prepare a gateway used in the FabricClient instance.
   */
  private async prepareGateway(): Promise<void> {

    if (this.gateway != null) {
      return;
    }

    const identity = this.getIdentity();
    const connectOpt: GatewayOptions = {
      identity: identity,
      discovery: {
        enabled: true,
        asLocalhost: this.config.discoverAsLocalhost
      },
    };
    this.gateway = new Gateway();
    await this.gateway.connect(this.config.connectionProfile, connectOpt);
    await this.setupAllRegisteredContractEventListeners();
  }
}