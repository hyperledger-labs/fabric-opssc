/*
 * Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { execSync } from 'child_process';
import fs from 'fs-extra';
import path from 'path';
import { logger } from './logger';

/**
 * This function execute an OS command.
 *
 * @param {string} command the command to run, with space-separated arguments
 * @param {boolean} [ignoreError] whether to ignore error
 * @param {string} [cwd] current working directory
 * @param [env] environment key-value pairs
 * @param {Buffer} [input] the value which will be passed as stdin
 * @return {string} the stdout from the command
 */
export function execCommand(command: string, ignoreError?: boolean, cwd?: string, env?: { [n: string]: string }, input?: Buffer): string {
  try {
    const result = execSync(command,
      { shell: '/bin/bash',
        env: env,
        cwd: cwd,
        input: input
      }).toString('utf-8').trimEnd();
    logger.debug(`Command: ${command}\nResult:\n${result}`);
    return result;
  } catch (error) {
    if (ignoreError === true) {
      logger.warn(`Command: ${command}\nError(ignored):\n${error}`);
      return error;
    }
    throw error;
  }
}

/**
 * This function returns file contents as UTF-8 format by reading the specified file or a file on the specified dir path.
 *
 * @param {string} fileOrDirPath file path or directory path
 * @returns {string} the file contents
 */
export function readSingleFileOnThePath(fileOrDirPath: string): string {
  const stat = fs.statSync(fileOrDirPath);
  if (stat.isFile()) {
    return fs.readFileSync(fileOrDirPath).toString('utf-8');
  }
  if (stat.isDirectory()) {
    const list = fs.readdirSync(fileOrDirPath);
    if (list !== undefined && list.length > 0) {
      return fs.readFileSync(path.join(fileOrDirPath, list[0])).toString('utf-8');
    }
  }
  throw new Error(`File is not found on ${fileOrDirPath}.`);
}

/**
 * This function returns file path of the specified file or a file on the specified dir path.
 *
 * @param {string} fileOrDirPath file path or directory path
 * @returns {string} the file path
 */
export function findSingleFileOnThePath(fileOrDirPath: string): string {
  const stat = fs.statSync(fileOrDirPath);
  if (stat.isFile()) {
    return fileOrDirPath;
  }
  if (stat.isDirectory()) {
    const list = fs.readdirSync(fileOrDirPath);
    if (list !== undefined && list.length > 0) {
      return path.join(fileOrDirPath, list[0]);
    }
  }
  throw new Error(`File is not found on ${fileOrDirPath}.`);
}