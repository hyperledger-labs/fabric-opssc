/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { config } from './config';
import { logger } from './logger';
import { Notifier } from './notifier';
import express from 'express';
import cors from 'cors';
import morgan from 'morgan';
import bodyParser from 'body-parser';
import { exit } from 'process';
import { FabricClient } from 'opssc-common/fabric-client';
import { ChannelOpsAgent } from './channel-ops-agent';
import { ChaincodeOpsAgent } from './chaincode-ops-agent';

/**
 * OpsSCAgent is a main class to works as an OpsSC agent for an organization. This internally uses agents to operate channels and chaincodes.
 * An instance of the class listens the chaincode events from the Ops chaincodes for operating channels and chaincodes,
 * and executes operations based on the events. Also, this makes the target organization's nodes to make available
 * for the OpsSC chaincodes and the existing chaincodes when the agent is launched.
 */
export class OpsSCAgent {

  private notifier?: Notifier;
  private fabricClient: FabricClient;
  private channelOpsAgent: ChannelOpsAgent;
  private chaincodeOpsAgent: ChaincodeOpsAgent;

  private _isReady: boolean;

  /**
   * OpsSCAgent constructor
   */
  constructor() {
    this.fabricClient = new FabricClient(config.fabric);
    if (config.ws.enabled) {
      this.notifier = new Notifier(config.ws, config.fabric.adminMSPID);
    }
    this.channelOpsAgent = new ChannelOpsAgent(this.fabricClient, config.core, this.notifier);
    this.chaincodeOpsAgent = new ChaincodeOpsAgent(this.fabricClient, config.core, this.notifier);

    this._isReady = false;
  }

  /**
   * Start the OpsSCAgent
   *
   * @async
   */
  public async start() {
    try {
      await this.channelOpsAgent.bootstrap();
    } catch (e) {
      logger.error(e);
      exit(1);
    }

    await this.fabricClient.addContractEventListener(
      config.core.opssc.channelID,
      config.core.opssc.chaincodes.channelOpsCCName,
      this.channelOpsAgent.createContractListener()
    );
    logger.info('channelOpsSC EventListener is ready');

    await this.fabricClient.addContractEventListener(
      config.core.opssc.channelID,
      config.core.opssc.chaincodes.chaincodeOpsCCName,
      this.chaincodeOpsAgent.createContractListener()
    );
    logger.info('chaincodeOpsSC EventListener is ready');

    this._isReady = true;
  }

  /**
   * Return whether the OpsSCAgent is ready or not
   *
   * @returns {boolean} whether the OpsSCAgent is ready or not
   */
  public isReady(): boolean {
    return this._isReady;
  }
}

// Start the OpsSCAgent
const agent = new OpsSCAgent();
agent.start().then(() => {
  logger.info('OpsSC agent is ready.');
});

// Start the API server for the OpsSCAgent
const app = express();
app.use(bodyParser.json());
const httpLogger = morgan('combined');
app.use(httpLogger);
const _cors = cors();
app.use(_cors);

const port = process.env.AGENT_SERVICE_PORT || 5500;
app.listen(port, () => {
  logger.info('Started OpsSC agent for %s with Port %d', config.fabric.adminMSPID, port);
});

// Provide API for health check
app.get('/healthz', (req, res) => {
  if (agent.isReady()) {
    res.send('OK');
  } else {
    res.status(500);
    res.send('NG');
  }
});