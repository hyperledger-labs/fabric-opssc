/*
 * Copyright 2022 Hitachi, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeOperatorImpl } from './chaincode-operator';
import { logger } from './logger';
import fs from 'fs-extra';
import { execCommand } from 'opssc-common/utils';

export class K8sBuilderBasedExternalChaincodeOperatorImpl extends ChaincodeOperatorImpl {

  async package() {
    // Build chaincode server image
    const digest = this.buildChaincodeServerImage();

    // Packaging chaincode
    logger.info(`[START] Package external chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Package chaincode', this.proposal.ID);

    const sourceAbsolutePath = this.sourceAbsolutePath();

    fs.mkdirpSync(`${sourceAbsolutePath}/packaging`);
    const imageJSONFile = `${sourceAbsolutePath}/packaging/image.json`;

    if (!this.config.ccs) throw new Error('ChaincodeServerConfig should be set');

    const registryName = (this.config.ccs.pullRegistry !== '') ? this.config.ccs.pullRegistry : this.config.ccs.registry;
    const imageJSON = {
      'name': `${registryName}/${this.proposal.chaincodeName}`,
      'digest': digest
    };
    fs.outputJSONSync(imageJSONFile, imageJSON);

    const ccLabel = this.chaincodeLabel();
    const metadataFile = `${sourceAbsolutePath}/packaging/metadata.json`;
    const metadataJSON = {
      'type': 'k8s',
      'label': ccLabel
    };
    fs.outputJSONSync(metadataFile, metadataJSON);
    execCommand(`pushd ${sourceAbsolutePath}/packaging && tar cfz code.tar.gz image.json && tar cfz ${ccLabel}.tgz code.tar.gz metadata.json && popd`);

    this.packagedChaincode = fs.readFileSync(`${sourceAbsolutePath}/packaging/${ccLabel}.tgz`);

    logger.info(`[END] Package chaincode (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Package chaincode', this.proposal.ID);
  }

  private buildChaincodeServerImage(): string {
    if (!this.config.ccs) throw new Error('ChaincodeServerConfig should be set');

    // Build external chaincode server
    logger.info(`[START] Build external chaincode server (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[START] Build external chaincode server', this.proposal.ID);

    const chaincodeName = this.proposal.chaincodeName;
    const imageTag = this.proposal.chaincodeDefinition.sequence.toString();
    const remote = this.remoteGitRepositoryURL(true);
    const pathToSourceFiles = this.proposal.chaincodePackage.pathToSourceFiles ? `/${this.proposal.chaincodePackage.pathToSourceFiles}` : '';

    // Build chaincode server image
    const helmCommand = `helm -n ${this.config.ccs.namespace} upgrade ${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix} --set name=${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix} --set imageName=${this.config.ccs.registry}/${chaincodeName} --set launchChaincodeServer=false --set imageTag=${imageTag} --set git.repositoryURL=${remote} --set git.commitID=${this.proposal.chaincodePackage.commitID} --set git.pathToSourceFiles=${pathToSourceFiles} --set chaincode.ccID=${this.packageID} --set imagePullSecretName=${this.config.ccs.ccServerImagePullSecretName} /opt/chart --install --timeout 10m --wait`;
    logger.info(`Helm command: ${helmCommand}`);
    execCommand(helmCommand);

    // Get image digest
    const digest = execCommand(`kubectl get pods -l "job-name=${this.config.ccs.servicePrefix}-${chaincodeName}-${this.config.ccs.serviceSuffix}-buildjob" -o jsonpath --template {.items[0].status.containerStatuses[0].state.terminated.message}`);
    logger.info(`Image digest: ${digest}`);

    logger.info(`[END] Build external chaincode server (proposal ID: ${this.proposal.ID})`);
    this.notifier?.notifyProgress('[END] Build external chaincode server', this.proposal.ID);

    return digest;
  }

  async download() {
    // The download process is skipped because the source code of the chaincode is cloned when building the images in the other step.
  }
}
