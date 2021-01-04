/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface OpsSCConfig {
  channelID: string;
  chaincodes: {
    chaincodeOpsCCName: string;
    channelOpsCCName: string;
  }
}