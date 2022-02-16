/*
 * Copyright 2020-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { setDefaultTimeout } from 'cucumber';
import moment from 'moment';
import axios from 'axios';

setDefaultTimeout(240 * 1000);

type servicePorts = {
  peer: number,
  orderer: number,
  ca: number,
  api: number,
  agent: number
}

type FabricVersion = {
  fabric: string,
  fabricCA: string
}

export default class BaseStepClass {

  protected static OPSSC_VERSION = process.env.npm_package_version;

  protected static FABRIC_TWO_DIGIT_VERSION = process.env.FABRIC_TWO_DIGIT_VERSION ? process.env.FABRIC_TWO_DIGIT_VERSION : '2.4';

  protected static FABRIC_VERSION_MAP: { [char: string]: FabricVersion } = {
    '2.4': { fabric: '2.4.2', fabricCA: '1.5.2' },
    '2.2': { fabric: '2.2.5', fabricCA: '1.5.2' },
  }

  protected static fabricVersion() {
    return BaseStepClass.FABRIC_VERSION_MAP[BaseStepClass.FABRIC_TWO_DIGIT_VERSION].fabric;
  }

  protected static fabricCAVersion() {
    return BaseStepClass.FABRIC_VERSION_MAP[BaseStepClass.FABRIC_TWO_DIGIT_VERSION].fabricCA;
  }

  protected static opsSCImageTag() {
    return `${BaseStepClass.OPSSC_VERSION}-for-fabric-${BaseStepClass.FABRIC_TWO_DIGIT_VERSION}`;
  }

  protected static TEST_NETWORK_PATH = '../sample-environments/fabric-samples/test-network';
  protected static OPS_CHANNEL = 'ops-channel';
  protected static CC_OPS_CC_NAME = 'chaincode_ops';
  protected static CH_OPS_CC_NAME = 'channel_ops';

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

  private async invokeOpsSCFunc(ccName: string, funcName: string, args: string[]): Promise<number> {
    const response = await axios.post(`${this.getAPIEndpoint()}/api/v1/utils/invokeTransaction`,
      {
        channelID: BaseStepClass.OPS_CHANNEL,
        ccName: ccName,
        func: funcName,
        args: args,
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
    return response.status;
  }

  protected async invokeChaincodeOpsFunc(funcName: string, args: string[]): Promise<number> {
    return this.invokeOpsSCFunc(BaseStepClass.CC_OPS_CC_NAME, funcName, args);
  }

  protected async invokeChannelOpsFunc(funcName: string, args: string[]): Promise<number> {
    return this.invokeOpsSCFunc(BaseStepClass.CH_OPS_CC_NAME, funcName, args);
  }
}
