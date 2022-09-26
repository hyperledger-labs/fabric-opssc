/*
 * Copyright 2019-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeUpdateProposal, TaskStatusUpdate } from 'opssc-common/opssc-types';
import { ChaincodeOperatorConfig } from './config';
import { logger } from './logger';
import path from 'path';
import simplegit from 'simple-git';
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
   * Validate that the proposal is deployable. Throw some error if it is not deployable.
   *
   * @returns {Promise<void>}
   */
   validate(): Promise<void>;

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
  private mspID: string;

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
    this.mspID = identity.mspId;
  }

  /**
   * Prepare to deploy (download, package, install and approve) the chaincode based on the proposal.
   *
   * @async
   * @returns {Promise<TaskStatusUpdate>} The result of the above tasks
   */
  async prepareToDeploy(): Promise<TaskStatusUpdate> {
    try {
      await this.validate();
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
      await this.notifier?.notifyError(`[ERROR] Prepare to deploy error\n${e.message}`, this.proposal.ID);
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
      await this.notifier?.notifyError(`[ERROR] Deploy error\n${e}`, this.proposal.ID);
      return commitTaskStatus;
    } finally {
      this.lifecycleCommands.close();
    }
  }

  /**
   * Validate that the proposal is deployable. Throw some error if it is not deployable.
   *
   * @async
   * @returns {Promise<void>}
   */
  async validate(): Promise<void> {
    logger.info(`[START] Validate chaincode update proposal (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Validate chaincode update proposal', this.proposal.ID);
    try {
      const chaincodeDefinition = await this.lifecycleCommands.queryChaincodeDefinition(this.proposal.chaincodeName);
      logger.debug('chaincodeDefinition\n%s', JSON.stringify(chaincodeDefinition));
      if (chaincodeDefinition.sequence != (this.proposal.chaincodeDefinition.sequence - 1)) {
        if (chaincodeDefinition.approvals![this.mspID]) {
          throw new Error(`VALIDATION ERROR: The proposed sequence should be committed sequence + 1 (proposed: ${this.proposal.chaincodeDefinition.sequence}, committed: ${chaincodeDefinition.sequence})`);
        }
        logger.warn('This organization is subject to the catch-up process for `prepareDeploy`. NOTE: The process for `deploy` should not be done in the subsequent');
      }
    } catch (e) {
      if (e.message == null || !(e.message as string).includes('is not defined')) {
        throw e;
      }
    }
    logger.info(`[END] Validate chaincode update proposal (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Validate chaincode update proposal', this.proposal.ID);
  }

  /**
   * Download the source code of the chaincode from the remote repository specified in the proposal.
   *
   * @async
   * @returns {Promise<void>}
   */
  async download(): Promise<void> {
    try {
      logger.info(`[START] Download chaincode (proposal ID: ${this.proposal.ID})`);
      this.notifier?.notifyProgress('[START] Download chaincode', this.proposal.ID);

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

      logger.info(`[END] Download chaincode (proposal ID: ${this.proposal.ID})`);
      this.notifier?.notifyProgress('[END] Download chaincode', this.proposal.ID);
    } catch (e) {
      if (this.config.gitUser && this.config.gitPassword) {
        const maskedErrorMessage = (e.message as string).replace(this.config.gitUser, '*****').replace(this.config.gitPassword, '*****');
        throw new Error(maskedErrorMessage);
      }
      throw e;
    }
  }

  /**
   * Package the chaincode using the downloaded source code by download().
   *
   * @async
   * @returns {Promise<void>}
   */
  async package(): Promise<void> {
    logger.info(`[START] Package chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Package chaincode', this.proposal.ID);

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

    logger.info(`[END] Package chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Package chaincode', this.proposal.ID);
  }

  /**
   * Install the chaincode using the packaged source code by package().
   *
   * @async
   * @returns {Promise<void>}
   */
  async install(): Promise<void> {
    logger.info(`[START] Install chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Install chaincode', this.proposal.ID);

    if (this.packagedChaincode == null) {
      throw new Error('package is not set');
    }

    const installRequest: InstallRequest = {
      package: this.packagedChaincode
    };
    const result = await this.lifecycleCommands.install(installRequest);
    this.packageID = result ? result : this.packageID = computePackageID(this.chaincodeLabel(), this.packagedChaincode);
    logger.info(`[END] Install chaincode (proposal ID: ${this.proposal.ID})`);
  }

  /**
   * Approve the chaincode definition with the package created by the other methods and the content of the proposal.
   * @async
   * @returns {Promise<void>}
   */
  async approve(): Promise<void> {
    logger.info(`[START] Approve chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Approve chaincode', this.proposal.ID);

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

    logger.info(`[END] Approve chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Approve chaincode', this.proposal.ID);
  }

  /**
   * Commit the chaincode definition based on the content of the proposal.
   *
   * @async
   * @returns {Promise<void>}
   */
  async commit(): Promise<void> {
    logger.info(`[START] Commit chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Commit chaincode', this.proposal.ID);

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

    logger.info(`[END] Commit chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Commit chaincode', this.proposal.ID);
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