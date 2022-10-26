/*
 * Copyright 2019-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeOperatorImpl } from './chaincode-operator';
import { logger } from './logger';
import fs from 'fs-extra';
import { execCommand } from 'opssc-common/utils';

export class ExternalChaincodeOperatorImpl extends ChaincodeOperatorImpl {

  async package() {
    logger.info(`[START] Package external chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Package chaincode', this.proposal.ID);

    const sourceAbsolutePath = this.sourceAbsolutePath();

    fs.mkdirpSync(`${sourceAbsolutePath}/packaging`);
    const connectionFile = `${sourceAbsolutePath}/packaging/connection.json`;

    if (!this.config.ccs) throw new Error('ChaincodeServerConfig should be set');
    const connectionJSON = {
      'address': `${this.config.ccs.servicePrefix}-${this.proposal.chaincodeName}-${this.config.ccs.serviceSuffix}.${this.config.ccs.namespace}:${this.config.ccs.servicePort}`,
      'dial_timeout': '10s',
      'tls_required': false,
      'client_auth_required': false,
      'client_key': '',
      'client_cert': '',
      'root_cert': ''
    };
    fs.outputJSONSync(connectionFile, connectionJSON);

    const ccLabel = this.chaincodeLabel();
    const metadataFile = `${sourceAbsolutePath}/packaging/metadata.json`;
    const metadataJSON = {
      'type': 'ccaas',
      'label': ccLabel
    };
    fs.outputJSONSync(metadataFile, metadataJSON);
    execCommand(`pushd ${sourceAbsolutePath}/packaging && tar cfz code.tar.gz connection.json && tar cfz ${ccLabel}.tgz code.tar.gz metadata.json && popd`);

    this.packagedChaincode = fs.readFileSync(`${sourceAbsolutePath}/packaging/${ccLabel}.tgz`);

    logger.info(`[END] Package chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Package chaincode', this.proposal.ID);
  }

  async install() {
    if (!this.config.ccs) throw new Error('ChaincodeServerConfig should be set');

    // First step: Install chaincode as same as standard chaincode deployment
    await super.install();

    // Next step: Build and launch external chaincode server
    if (this.config.ccs.launchFromAgent) {
      logger.info(`[START] Build and launch external chaincode server (proposal ID: ${this.proposal.ID})`);
      this.notifier?.notifyProgress('[START] Build and launch external chaincode server', this.proposal.ID);

      const chaincodeName = this.proposal.chaincodeName;
      const imageTag = this.proposal.chaincodeDefinition.sequence.toString();
      const remote = this.remoteGitRepositoryURL(true);
      const pathToSourceFiles = this.proposal.chaincodePackage.pathToSourceFiles ? `/${this.proposal.chaincodePackage.pathToSourceFiles}` : '';

      // Workaround: Avoid 'UPGRADE FAILED: another operation (install/upgrade/rollback) is in progress'
      // execCommand(`helm -n ${this.config.ccs.namespace} uninstall ${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix} || true`);

      // Build chaincode server image
      const setOptionForImageNameOverride = (this.config.ccs.pullRegistry !== '') ? `--set pullImageNameOverride=${this.config.ccs.pullRegistry}/chaincode/${chaincodeName}` : '';
      const helmCommand = `helm -n ${this.config.ccs.namespace} upgrade ${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix} --set name=${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix} --set imageName=${this.config.ccs.registry}/chaincode/${chaincodeName} ${setOptionForImageNameOverride} --set imageTag=${imageTag} --set git.repositoryURL=${remote} --set git.commitID=${this.proposal.chaincodePackage.commitID} --set git.pathToSourceFiles=${pathToSourceFiles} --set chaincode.ccID=${this.packageID} --set imagePullSecretName=${this.config.ccs.ccServerImagePullSecretName} /opt/chart --install --timeout 3m --wait`;
      logger.info(`Helm command: ${helmCommand}`);
      execCommand(helmCommand);

      logger.info(`[END] Build and launch external chaincode server (proposal ID: ${this.proposal.ID})`);
      this.notifier?.notifyProgress('[END] Build and launch external chaincode server', this.proposal.ID);
    }
  }

  async download() {
    // The download process is skipped because the source code of the chaincode is cloned when building the images in the other step.
  }
}
