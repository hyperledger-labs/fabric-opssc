/*
 * Copyright 2019-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { config } from './config';
import bodyParser from 'body-parser';
import cors from 'cors';
import express from 'express';
import morgan from 'morgan';
import WebSocket from 'ws';
import { ContractEvent, ContractListener } from 'fabric-network';
import { FabricClient } from 'opssc-common/fabric-client';
import api from './api';
import { logger } from './logger';

// Initialize API server
const app = express();
app.use(bodyParser.json());
const httpLogger = morgan('dev');
app.use(httpLogger);
const _cors = cors();
app.use(_cors);

const fabricClient = new FabricClient(config.fabric);
app.use('/api/v1', api(fabricClient, config));

app.get('/healthz', (_req, res) => {
  // TODO: More diagnosis
  res.send('OK');
});

const port = process.env.CLIENT_SERVICE_PORT || 5000;
const server = app.listen(port, () => {
  logger.info('Started OpsSC API Server for %s with Port %d', config.fabric.adminMSPID, port);
});

// Initialize websocket server and client
let ws: WebSocket|null = null;

if (config.ws.enabled) {
  logger.info('Starting ws server and client');

  const wss = new WebSocket.Server({ server: server });
  let connections: WebSocket[] = [];

  wss.on('connection', function (ws) {
    logger.info('Websocket new connection');
    connections.push(ws);

    ws.on('open', function () {
      logger.info('Opened ws server');
    });

    ws.on('close', function () {
      connections = connections.filter(function (conn, _i) {
        return (conn == ws) ? false : true;
      });
    });

    ws.on('message', function incoming(data: string, _flags: any) {
      logger.info('message: ' + data);
      connections.forEach(function (conn, _i) {
        conn.send(data);
      });
    });

    ws.on('error', function (err) {
      logger.error(`Web socket server-side error: ${err}`);
    });
  });

  ws = new WebSocket(`ws://localhost:${port}`);
  ws.on('open', function open() {
    logger.info('Opened ws client');
  });
  ws.on('error', function (err) {
    logger.error(`Web socket client error: ${err}`);
  });
}

// Initialize contract listeners
async function registerContractListeners() {
  for (const opsSCName of Object.values(config.opsSC.chaincodes)) {
    await fabricClient.addContractEventListener(config.opsSC.channelID, opsSCName, createCommonContractListener());
  }
}

registerContractListeners().catch((e) => {
  logger.error(`Fail to register contract listeners: ${e}`);
});

function createCommonContractListener(): ContractListener {
  return (
    async (event: ContractEvent) => {
      try {
        logger.info(`Receive chaincode event: ${event.eventName}`);
        handleChaincodeEvent(event);
      } catch (e) {
        logger.error(`Fail to handle chaincode event: ${e}`);
      }
    });
}

function handleChaincodeEvent(event: ContractEvent) {
  const message = {
    type: 'chaincodeEvent',
    chaincodeName: event.chaincodeId,
    eventName: event.eventName
  };
  try {
    ws?.send(JSON.stringify(message));
  } catch (e) {
    logger.error('Fail to send message by web socket: %s', e);
  }
}