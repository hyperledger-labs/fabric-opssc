/*
 * Copyright 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export function package(chaincodePath: string, chaincodeType: string, metadataPath?: string, goPath?: string): Promise<Buffer>;
export function finalPackage(label: string, chaincodeType: string, packageBytes: Buffer, chaincodePath?: string): Promise<Buffer>;