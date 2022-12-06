/*
 * Copyright 2020-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { setDefaultTimeout } from 'cucumber';
import moment from 'moment';
import axios from 'axios';

setDefaultTimeout(300 * 1000);

type servicePorts = {
  peer: number,
  orderer: number,
  ca: number,
  'api-server': number,
  agent: number
}

type FabricVersion = {
  fabric: string,
  fabricCA: string,
  k8sFabricPeer?: string
}

export default class BaseStepClass {

  protected static OPSSC_VERSION = process.env.npm_package_version;

  protected static FABRIC_TWO_DIGIT_VERSION = process.env.FABRIC_TWO_DIGIT_VERSION ? process.env.FABRIC_TWO_DIGIT_VERSION : '2.4';

  protected static FABRIC_VERSION_MAP: { [char: string]: FabricVersion } = {
    '2.4': { fabric: '2.4.7', fabricCA: '1.5.5', k8sFabricPeer: 'v0.8.0' },
    '2.2': { fabric: '2.2.9', fabricCA: '1.5.5' },
  }

  protected static fabricVersion() {
    return BaseStepClass.FABRIC_VERSION_MAP[BaseStepClass.FABRIC_TWO_DIGIT_VERSION].fabric;
  }

  protected static fabricCAVersion() {
    return BaseStepClass.FABRIC_VERSION_MAP[BaseStepClass.FABRIC_TWO_DIGIT_VERSION].fabricCA;
  }

  protected static k8sFabricPeerVersion() {
    return BaseStepClass.FABRIC_VERSION_MAP[BaseStepClass.FABRIC_TWO_DIGIT_VERSION].k8sFabricPeer;
  }

  protected static opsSCImageTag() {
    return `${BaseStepClass.OPSSC_VERSION}-for-fabric-${BaseStepClass.FABRIC_TWO_DIGIT_VERSION}`;
  }

  protected static K8S_SUPPORT_PATH = '../sample-environments/k8s-support';
  protected static TEST_NETWORK_K8S_PATH = `${BaseStepClass.K8S_SUPPORT_PATH}/fabric-samples/test-network-k8s`;
  protected static SAMPLE_NETWORK_IN_FABRIC_OPERATOR_PATH = `${BaseStepClass.K8S_SUPPORT_PATH}/fabric-operator/sample-network`;
  protected static TEST_NETWORK_PATH = '../sample-environments/fabric-samples/test-network';

  protected static OPS_CHANNEL = 'ops-channel';
  protected static CC_OPS_CC_NAME = 'chaincode-ops';
  protected static CH_OPS_CC_NAME = 'channel-ops';

  protected static RETRY = 30;
  protected static SUFFIX = moment().format('MMDD-HHmmss');
  protected static INTERVAL_SECS = 5 * 1000;
  protected static SERVICE_PORT_MAP: { [char: string]: servicePorts } = {
    org1: { peer: 7051,  orderer: 7050,  ca: 7054,  'api-server': 5000, agent: 5500 },
    org2: { peer: 9051,  orderer: 9050,  ca: 9054,  'api-server': 5001, agent: 5501 },
    org3: { peer: 11051, orderer: 11050, ca: 11054, 'api-server': 5002, agent: 5502 },
    org4: { peer: 13051, orderer: 13050, ca: 13054, 'api-server': 5003, agent: 5503 },
  }

  protected environment:'docker'|'k8s' = 'docker';

  protected delay(ms: number = BaseStepClass.INTERVAL_SECS) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  protected getAgentServiceEndpointDocker(org = 'org1', service: 'api-server'|'agent' = 'api-server'): string {
    return `http://localhost:${BaseStepClass.SERVICE_PORT_MAP[org][service]}`;
  }

  protected getServiceEndpoint(environment: 'docker'|'k8s' = 'docker', org = 'org1', service: 'api-server'|'agent' = 'api-server'): string {
    if (environment === 'k8s') return this.getServiceEndpointK8s(org, service);
    return this.getAgentServiceEndpointDocker(org, service);
  }

  protected getServiceEndpointK8s(org = 'org1', service: 'api-server'|'agent' = 'api-server'): string {
    return `http://${org}-opssc-${service}.localho.st`;
  }

  private async invokeOpsSCFunc(ccName: string, funcName: string, args: string[]): Promise<number> {
    const response = await axios.post(`${this.getServiceEndpoint(this.environment)}/api/v1/utils/invokeTransaction`,
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
