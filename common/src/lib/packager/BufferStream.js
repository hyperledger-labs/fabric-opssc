/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
// This code is based on fabric-client v2.0.0-beta.2 in https://github.com/hyperledger/fabric-sdk-node.

'use strict';

const stream = require('stream');

class BufferStream extends stream.PassThrough {

	constructor() {
		super();
		this.buffers = [];
		this.on('data', (chunk) => {
			this.buffers.push(chunk);
		});
	}

	toBuffer() {
		return Buffer.concat(this.buffers);
	}

}

module.exports = BufferStream;