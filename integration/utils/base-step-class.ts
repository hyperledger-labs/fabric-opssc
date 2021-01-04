/*
 * Copyright 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { setDefaultTimeout } from 'cucumber';
import moment from 'moment';

setDefaultTimeout(240 * 1000);

type servicePorts = {
  peer: number,
  orderer: number,
  ca: number,
  api: number,
  agent: number
}

export default class BaseStepClass {

  protected static FABRIC_VERSION = '2.3.0';
  protected static FABRIC_CA_VERSION = '1.4.9';

  protected static TEST_NETWORK_PATH = '../sample-environments/fabric-samples/test-network';

  protected static RETRY = 15;
  protected static SUFFIX = moment().format('MMDD_HHmmss');
  protected static INTERVAL_SECS = 5 * 1000;
  protected static SERVICE_PORT_MAP: { [char: string]: servicePorts } = {
    org1: { peer: 7051,  orderer: 7050,  ca: 7054,  api: 5000, agent: 5500 },
    org2: { peer: 9051,  orderer: 9050,  ca: 9054,  api: 5001, agent: 5501 },
    org3: { peer: 11051, orderer: 11050, ca: 11054, api: 5002, agent: 5502 },
    org4: { peer: 13051, orderer: 13050, ca: 13054, api: 5003, agent: 5503 },
  }

  protected delay(ms: number = BaseStepClass.INTERVAL_SECS) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  protected getAPIEndpoint(org = 'org1') {
    return `http://localhost:${BaseStepClass.SERVICE_PORT_MAP[org].api}`;
  }

  protected getAgentServiceEndpoint(org = 'org1') {
    return `http://localhost:${BaseStepClass.SERVICE_PORT_MAP[org].agent}`;
  }
}