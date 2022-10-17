/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as fs from 'fs';
import yaml from 'js-yaml';
import { FabricConfig } from 'opssc-common/fabric-client';
import { OpsSCConfig } from 'opssc-common/config';
import { readSingleFileOnThePath } from 'opssc-common/utils';

export interface WebSocketConfig {
  enabled: boolean;
  websocketUrl: string;
}

export interface OpsSCAgentConfig {
  fabric: FabricConfig;
  ws: WebSocketConfig;
  core: OpsSCAgentCoreConfig;
}

export interface OpsSCAgentCoreConfig {
  opssc: OpsSCConfig;
  ccops: ChaincodeOperatorConfig;
}

export interface ChaincodeOperatorConfig {
  gitUser?: string | undefined;
  gitPassword?: string | undefined;
  goPath?: string | undefined;
  ccs?: ChaincodeServerConfig;
}

export interface ChaincodeServerConfig {
  launchFromAgent: boolean;
  registry: string;
  pullRegistry: string;
  ccServerImagePullSecretName: string;
  namespace: string;
  servicePrefix: string;
  serviceSuffix: string;
  servicePort: string;
}

export const config: OpsSCAgentConfig = {
  fabric: {
    adminMSPID: process.env.ADMIN_MSPID || 'Org1MSP',
    adminMSPConfigPath: process.env.MSP_CONFIG_PATH || '/opt/fabric/msp',
    adminKey: readSingleFileOnThePath(process.env.ADMIN_KEY || '/opt/fabric/msp/admin.key'),
    adminCert: readSingleFileOnThePath(process.env.ADMIN_CERT || '/opt/fabric/msp/admin.crt'),
    connectionProfile: yaml.safeLoad(fs.readFileSync(process.env.CONNECTION_PROFILE || '/opt/fabric/config/connection-profile.yaml', 'utf8')) as Record<string, any>,
    discoverAsLocalhost: process.env.DISCOVER_AS_LOCALHOST === 'true',
  },
  ws: {
    websocketUrl: process.env.WS_URL || 'ws://localhost:5000',
    enabled: process.env.WS_ENABLED === 'true',
  },
  core: {
    opssc: {
      channelID: process.env.CHANNEL_NAME || 'ops-channel',
      chaincodes: {
        chaincodeOpsCCName: process.env.CC_OPS_CC_NAME || 'chaincode_ops',
        channelOpsCCName: process.env.CH_OPS_CC_NAME || 'channel_ops'
      }
    },
    ccops: {
      gitUser: process.env.GIT_USER,
      gitPassword: process.env.GIT_PASSWORD,
      goPath: process.env.GOPATH,
      ccs: {
        launchFromAgent: process.env.CC_SERVER_LAUNCH_FROM_AGENT === 'true',
        registry: process.env.CC_SERVER_REGISTRY || '',
        pullRegistry: process.env.CC_SERVER_PULL_REGISTRY || process.env.CC_SERVER_REGISTRY || '',
        ccServerImagePullSecretName: process.env.CC_SERVER_IMAGE_PULL_SECRET_NAME || '',
        namespace: process.env.CC_SERVER_NAMESPACE || 'hyperledger',
        servicePrefix: process.env.CC_SERVER_PREFIX || 'chaincode',
        serviceSuffix: process.env.CC_SERVER_SUFFIX || 'org1',
        servicePort: process.env.CC_SERVER_PORT || '7052',
      }
    }
  }
};