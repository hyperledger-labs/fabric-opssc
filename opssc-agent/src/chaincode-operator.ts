/*
 * Copyright 2019-2021 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeUpdateProposal, TaskStatusUpdate } from 'opssc-common/opssc-types';
import { ChaincodeOperatorConfig } from './config';
import { logger } from './logger';
import path from 'path';
import simplegit from 'simple-git/promise';
import fs from 'fs-extra';
import { Notifier } from './notifier';
import { ChaincodeLifecycleCommands, ChaincodeRequest, computePackageID, InstallRequest, PackageRequest } from 'opssc-common/chaincode-lifecycle-commands';
import { Identity } from 'fabric-network';
import { execCommand } from 'opssc-common/utils';

/**
 * ChaincodeOperator is an interface which provides functions to execute operations to deploy a chaincode based on a chaincode update proposal.
 */
export interface ChaincodeOperator {

  /**
   * Download the source code of the chaincode from the remote repository specified in the proposal.
   *
   * @returns {Promise<void>}
   */
  download(): Promise<void>;

  /**
   * Package the chaincode using the downloaded source code by download().
   *
   * @returns {Promise<void>}
   */
  package(): Promise<void>;

  /**
   * Install the chaincode using the packaged source code by package().
   *
   * @returns {Promise<void>}
   */
  install(): Promise<void>;

  /**
   * Approve the chaincode definition with the package created by the other methods and the content of the proposal.
   *
   * @returns {Promise<void>}
   */
  approve(): Promise<void>;

  /**
   * Commit the chaincode definition based on the content of the proposal.
   *
   * @returns {Promise<void>}
   */
  commit(): Promise<void>;

  /**
   * Prepare to deploy (download, package, install and approve) the chaincode based on the proposal.
   *
   * @returns {Promise<TaskStatusUpdate>} The result of the above tasks
   */
  prepareToDeploy(): Promise<TaskStatusUpdate>;

  /**
   * Deploy (commit) the chaincode based on the proposal.
   *
   * @returns {Promise<TaskStatusUpdate>} The result of the above tasks
   */
  deploy(): Promise<TaskStatusUpdate>;
}

/**
 * ChaincodeOperatorImpl is a basic implementation of ChaincodeOperator.
 */
export class ChaincodeOperatorImpl implements ChaincodeOperator {
  protected proposal: ChaincodeUpdateProposal;
  protected packagedChaincode: Buffer | null;
  protected packageID: string;
  private lifecycleCommands: ChaincodeLifecycleCommands;
  readonly notifier?: Notifier;
  protected config: ChaincodeOperatorConfig;

  /**
   * ChaincodeOperatorImpl constructor
   *
   * @param {ChaincodeOperatorConfig} config the configuration of the operator
   * @param {ChaincodeUpdateProposal} proposal the client identity to interact with peers to operate chaincodes
   * @param {Identity} identity the client identity to interact with peers to operate chaincodes
   * @param {Record<string, any>} connectionProfile the connection profile that provides the necessary connection information for the client organization
   * @param {boolean} [discoverAsLocalhost] whether to discover the target nodes as localhost
   * @param {Notifier} [notifier] notification destination of work progress of this operator
   */
  constructor(config: ChaincodeOperatorConfig, proposal: ChaincodeUpdateProposal, identity: Identity, connectionProfile: Record<string, any>, discoverAsLocalhost?: boolean, notifier?: Notifier) {
    this.proposal = proposal;
    this.lifecycleCommands = new ChaincodeLifecycleCommands(proposal.channelID, identity, connectionProfile, discoverAsLocalhost);
    this.notifier = notifier;
    this.packagedChaincode = null;
    this.packageID = '';
    this.config = config;
  }

  /**
   * Prepare to deploy (download, package, install and approve) the chaincode based on the proposal.
   *
   * @async
   * @returns {Promise<TaskStatusUpdate>} The result of the above tasks
   */
  async prepareToDeploy(): Promise<TaskStatusUpdate> {
    try {
      await this.download();
      await this.package();
      await this.install();
      await this.approve();

      const approvalTaskStatus: TaskStatusUpdate = {
        proposalID: this.proposal.ID,
        status: 'success',
        data: `{"packageID": "${this.packageID}"}`
      };
      return approvalTaskStatus;
    } catch (e) {
      logger.error(e.message);
      const approvalTaskStatus: TaskStatusUpdate = {
        proposalID: this.proposal.ID,
        status: 'failure',
        data: `{"error": "${e.message}"}`
      };
      await this.notifier?.notifyError(`[ERROR] Prepare to deploy error\n${e.message}`);
      return approvalTaskStatus;
    } finally {
      this.lifecycleCommands.close();
    }
  }

  /**
   * Deploy (commit) the chaincode based on the proposal.
   *
   * @async
   * @returns {Promise<TaskStatusUpdate>} The result of the above tasks
   */
  async deploy(): Promise<TaskStatusUpdate> {
    try {
      await this.commit();

      const commitTaskStatus: TaskStatusUpdate = {
        proposalID: this.proposal.ID,
        status: 'success',
        data: ''
      };
      return commitTaskStatus;
    } catch (e) {
      logger.error(e.message);
      const commitTaskStatus: TaskStatusUpdate = {
        proposalID: this.proposal.ID,
        status: 'failure',
        data: `{"error": "${e.message}"}`
      };
      await this.notifier?.notifyError(`[ERROR] Deploy error\n${e}`);
      return commitTaskStatus;
    } finally {
      this.lifecycleCommands.close();
    }
  }

  /**
   * Download the source code of the chaincode from the remote repository specified in the proposal.
   *
   * @async
   * @returns {Promise<void>}
   */
  async download(): Promise<void> {
    logger.info('[START] Download chaincode');

    const sourceAbsolutePath = this.sourceAbsolutePath();
    const sourceParentAbsolutePath = path.resolve(sourceAbsolutePath, '..');

    fs.mkdirpSync(sourceParentAbsolutePath);

    if (fs.existsSync(sourceAbsolutePath)) {
      logger.debug('Move old existing chaincode sourcecode to \'.orginal\'');
      fs.removeSync(`${sourceAbsolutePath}.original`);
      fs.moveSync(sourceAbsolutePath, `${sourceAbsolutePath}.original`);
    }

    let git = simplegit(sourceParentAbsolutePath);
    const remote = this.remoteGitRepositoryURL(true);
    await git.clone(remote);
    git = simplegit(sourceAbsolutePath);
    await git.checkout(this.proposal.chaincodePackage.commitID);

    // Build chaincode for typescript
    if (this.proposal.chaincodePackage.type === 'typescript') {
      execCommand('npm install', false, this.chaincodeAbsolutePath());
      execCommand('npm run build', false, this.chaincodeAbsolutePath());
    }

    logger.info('[END] Download chaincode');
  }

  /**
   * Package the chaincode using the downloaded source code by download().
   *
   * @async
   * @returns {Promise<void>}
   */
  async package(): Promise<void> {
    logger.info('[START] Package chaincode');
    this.notifier?.notifyProgress('[START] Package chaincode');

    let lang = this.proposal.chaincodePackage.type;
    switch (this.proposal.chaincodePackage.type) {
      case 'typescript':
        lang = 'node';
        break;
      case 'javascript':
        lang = 'node';
        break;
    }

    const packageRequest: PackageRequest = {
      lang: lang,
      label: this.chaincodeLabel(),
      chaincodePath: this.proposal.chaincodePackage.type === 'golang' ? this.chaincodePath() : this.chaincodeAbsolutePath(),
      goPath: this.proposal.chaincodePackage.type === 'golang' ? this.getGoPath() : undefined
    };
    this.packagedChaincode = await this.lifecycleCommands.package(packageRequest);

    logger.info('[END] Package chaincode');
    this.notifier?.notifyProgress('[END] Package chaincode');
  }

  /**
   * Install the chaincode using the packaged source code by package().
   *
   * @async
   * @returns {Promise<void>}
   */
  async install(): Promise<void> {
    logger.info('[START] Install chaincode');
    this.notifier?.notifyProgress('[START] Install chaincode');

    if (this.packagedChaincode == null) {
      throw new Error('package is not set');
    }

    const installRequest: InstallRequest = {
      package: this.packagedChaincode
    };
    const result = await this.lifecycleCommands.install(installRequest);
    this.packageID = result ? result : this.packageID = computePackageID(this.chaincodeLabel(), this.packagedChaincode);
    logger.info('[END] Install chaincode');
  }

  /**
   * Approve the chaincode definition with the package created by the other methods and the content of the proposal.
   * @async
   * @returns {Promise<void>}
   */
  async approve(): Promise<void> {
    logger.info('[START] Approve chaincode');
    this.notifier?.notifyProgress('[START] Approve chaincode');

    const chaincodeRequest: ChaincodeRequest = {
      chaincode: {
        name: this.proposal.chaincodeName,
        sequence: this.proposal.chaincodeDefinition.sequence,
        version: this.proposal.chaincodeDefinition.sequence.toString(),
        validation_parameter: this.decodeValidationParameterFromBase64(),
        init_required: this.proposal.chaincodeDefinition.initRequired,
        package_id: this.packageID,
        collections: this.decodeCollectionsFromBase64(),
      }
    };
    await this.lifecycleCommands.approve(chaincodeRequest);

    logger.info('[END] Approve chaincode');
    this.notifier?.notifyProgress('[END] Approve chaincode');
  }

  /**
   * Commit the chaincode definition based on the content of the proposal.
   *
   * @async
   * @returns {Promise<void>}
   */
  async commit(): Promise<void> {
    logger.info('[START] Commit chaincode');
    this.notifier?.notifyProgress('[START] Commit chaincode');

    const chaincodeRequest: ChaincodeRequest = {
      chaincode: {
        name: this.proposal.chaincodeName,
        sequence: this.proposal.chaincodeDefinition.sequence,
        version: this.proposal.chaincodeDefinition.sequence.toString(),
        validation_parameter: this.decodeValidationParameterFromBase64(),
        init_required: this.proposal.chaincodeDefinition.initRequired,
        collections: this.decodeCollectionsFromBase64(),
      }
    };
    await this.lifecycleCommands.commit(chaincodeRequest);

    logger.info('[END] Commit chaincode');
    this.notifier?.notifyProgress('[END] Commit chaincode');
  }

  // Internal methods

  private decodeValidationParameterFromBase64(): any {
    const validationParameterAsString = Buffer.from(this.proposal.chaincodeDefinition.validationParameter.toString(), 'base64').toString();
    logger.info('Validation Parameter:\n%s', validationParameterAsString);
    let decodedValidationParameter: string | any = validationParameterAsString as string;
    try {
      const validationParameterAsJSON = JSON.parse(validationParameterAsString);
      decodedValidationParameter = validationParameterAsJSON;
    } catch (e) {
      logger.warn('Validation Parameter can not be parsed as JSON. It will be processed as String');
    }
    return decodedValidationParameter;
  }

  private decodeCollectionsFromBase64(): any {
    if (this.proposal.chaincodeDefinition.collections === undefined) return undefined;

    const collectionsAsString = Buffer.from(this.proposal.chaincodeDefinition.collections.toString(), 'base64').toString();
    logger.info('Private collections:\n%s', collectionsAsString);
    try {
      return JSON.parse(collectionsAsString);
    } catch (e) {
      logger.error('Private collections can not be parsed as JSON.');
      throw e;
    }
  }


  protected getGoPath(): string {
    return this.config.goPath || '';
  }

  protected sourceAbsolutePath(): string {
    const basePath = this.proposal.chaincodePackage.type === 'golang' ? this.getGoPath() : '/tmp';
    // const basePath = this.getGoPath();
    return path.join(basePath, 'src', this.repositoryFolderName());
  }

  protected repositoryFolderName(): string {
    return this.proposal.chaincodePackage.repository.replace(/\.git$/g, '');
  }

  protected chaincodePath(): string {
    if (!this.proposal.chaincodePackage.pathToSourceFiles || this.proposal.chaincodePackage.pathToSourceFiles === null) {
      return this.repositoryFolderName();
    }
    return path.join(this.repositoryFolderName(), this.proposal.chaincodePackage.pathToSourceFiles);
  }

  protected chaincodeAbsolutePath(): string {
    if (!this.proposal.chaincodePackage.pathToSourceFiles || this.proposal.chaincodePackage.pathToSourceFiles === null) {
      return this.sourceAbsolutePath();
    }
    return path.join(this.sourceAbsolutePath(), this.proposal.chaincodePackage.pathToSourceFiles);
  }

  protected chaincodeLabel(): string {
    return `${this.proposal.chaincodeName}_${this.proposal.chaincodeDefinition.sequence.toString()}`;
  }

  protected remoteGitRepositoryURL(withGitCredentials: boolean): string {
    const repository = this.repositoryFolderName();

    let gitURLLoginCredentials = '';
    if (withGitCredentials &&
      this.config.gitUser !== undefined &&
      this.config.gitPassword !== undefined) {
      gitURLLoginCredentials = `${this.config.gitUser}:${this.config.gitPassword}@`;
    }

    return `https://${gitURLLoginCredentials}${repository}.git`;
  }
}