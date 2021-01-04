/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from './logger';
import { WebSocketConfig } from './config';
import WebSocket from 'ws';

/**
 * Notifier is a class to send messages on the OpsSC to a WebSocket server.
 */
export class Notifier {

  private readonly mspID: string;
  private ws?: WebSocket;

  /**
   * Notifier constructor
   *
   * @param {WebSocketConfig} config the configuration on the WebSocket server
   * @param {string} mspID the MSP ID of the sender
   */
  constructor(config: WebSocketConfig, mspID: string) {
    this.mspID = mspID;
    if (!config.enabled) throw new Error('WebSocket should be enabled.');

    logger.debug('WS URL: %s', config.websocketUrl);
    try {
      this.ws = new WebSocket(config.websocketUrl);
      this.ws.on('open', function open() {
        logger.info('Opened ws');
      });
    } catch (e) {
      logger.error('Websocket opening fails: %s', e);
    }
  }

  /**
   * Send a message on a task progress.
   *
   * @param {string} message the message
   */
  notifyProgress(message: string) {
    this.sendMessageByWebSocket({
      type: 'log',
      org: this.mspID,
      message: message
    });
  }

  /**
   * Send a message on an event.
   *
   * @param {string} eventName the event name
   * @param {string} message the message
   */
  notifyEvent(eventName: string, message: string) {
    this.sendMessageByWebSocket({
      type: 'event',
      eventName: eventName,
      org: this.mspID,
      message: message
    });
  }

  /**
   * Send a message on an error.
   *
   * @param {string} message the message
   */
  notifyError(message: string) {
    this.sendMessageByWebSocket({
      type: 'error',
      org: this.mspID,
      message: message
    });
  }

  /**
   * Close the connection to the WebSocket server.
   */
  close() {
    try {
      this.ws?.close();
    } catch (e) {
      logger.error('Websocket closing fails: %s', e);
    }
  }

  /*
   * Internal method to send message to the WebSocket server.
   */
  private sendMessageByWebSocket(message: any) {
    try {
      this.ws?.send(JSON.stringify(message));
    } catch (e) {
      logger.error('Fail to send message by web socket: %s', e);
    }
  }
}
