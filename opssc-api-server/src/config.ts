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

export interface OpsSCAPIServerConfig {
    fabric: FabricConfig;
    opsSC: OpsSCConfig;
}

export const config: OpsSCAPIServerConfig = {
  fabric: {
    adminCert: readSingleFileOnThePath(process.env.ADMIN_CERT || '/opt/fabric/msp/signcerts'),
    adminKey: readSingleFileOnThePath(process.env.ADMIN_KEY || '/opt/fabric/msp/keystore'),
    adminMSPConfigPath: process.env.MSP_CONFIG_PATH || '/opt/fabric/msp',
    adminMSPID: process.env.ADMIN_MSPID || 'Org1MSP',
    discoverAsLocalhost: process.env.DISCOVER_AS_LOCALHOST !== 'false',
    connectionProfile: yaml.safeLoad(fs.readFileSync(process.env.CONNECTION_PROFILE || '/opt/fabric/config/connection-profile.yaml', 'utf8')) as Record<string, any>,
  },
  opsSC: {
    channelID: process.env.CHANNEL_NAME || 'ops-channel',
    chaincodes: {
      chaincodeOpsCCName: process.env.CC_OPS_CC_NAME || 'chaincode_ops',
      channelOpsCCName: process.env.CH_OPS_CC_NAME || 'channel_ops',
    }
  }
};