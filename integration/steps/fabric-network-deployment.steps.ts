/*
 * Copyright 2020-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { after, before, binding, given } from 'cucumber-tsflow';
import { expect } from 'chai';
import axios from 'axios';
import fs from 'fs-extra';
import { execSync } from 'child_process';
import BaseStepClass from '../utils/base-step-class';

@binding()
export class FabricNetworkDeploymentSteps extends BaseStepClass {

  @before('on-docker')
  public beforeDockerScenarios() {
    this.cleanupFabricNetwork();

    // Assume a case where garbage remains on the K8s scenarios
    this.cleanupTestNetworkK8s();
  }

  @after('on-docker')
  public afterDockerScenarios() {
    if (process.env.PRESERVE_TEST_NETWORK !== 'true') {
      this.cleanupFabricNetwork();
    }
  }

  private cleanupFabricNetwork() {
    let commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents-org3.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents-org4.yaml down --volumes --remove-orphans`;
    execSync(commands);

    commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers-org3.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers-org4.yaml down --remove-orphans`;
    execSync(commands);

    commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh down`;
    execSync(commands);
  }

  @before('on-k8s')
  public beforeK8sScenarios() {

    // Assume a case where garbage remains on the Docker scenarios
    this.cleanupFabricNetwork();

    this.environment = 'k8s';
    this.cleanupTestNetworkK8s();

    this.cloneAndSetupFabricSamples();
    this.createKINDCluster();
    this.loadDockerImagesForOpsSCIntoKIND();
  }

  @after('on-k8s')
  public afterK8sScenarios() {
    if (process.env.PRESERVE_TEST_NETWORK !== 'true') {
      this.cleanupTestNetworkK8s();
    }
  }

  private applyPatchToTestNetwork8s(fileName: string) {
    const commands = `cd ${BaseStepClass.K8S_SUPPORT_PATH} && patch -u fabric-samples/test-network-k8s/scripts/${fileName} < ${fileName}.patch`;
    execSync(commands);
  }

  private cleanupTestNetworkK8s() {
    if (fs.existsSync(BaseStepClass.TEST_NETWORK_K8S_PATH)) {
      const commands = `cd ${BaseStepClass.TEST_NETWORK_K8S_PATH} && ./network down && ./network cluster clean || true && ./network unkind`;
      execSync(commands);
    }
  }

  private cloneAndSetupFabricSamples() {
    // Remove old fabric-samples and clone fabric-samples
    const commands = `cd ${BaseStepClass.K8S_SUPPORT_PATH} && rm -rf fabric-samples && git clone https://github.com/hyperledger/fabric-samples.git`;
    execSync(commands);

    // Apply patches
    this.applyPatchToTestNetwork8s('chaincode.sh');
    this.applyPatchToTestNetwork8s('ccp-template.json');
    this.applyPatchToTestNetwork8s('rest_sample.sh');
  }

  private createKINDCluster() {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_K8S_PATH} && ./network kind && ./network cluster init`;
    execSync(commands);
  }

  private loadDockerImagesForOpsSCIntoKIND() {
    const imageNames =['fabric-opssc/opssc-api-server:latest', 'fabric-opssc/opssc-agent:latest'];
    for (const imageName of imageNames) {
      const commands = `kind load docker-image ${imageName}`;
      execSync(commands, {
        cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
      });
    }
  }

  @given(/download Fabric binaries/, 'on-docker')
  public downloadFabricBinaries() {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH}/.. && curl -sSL https://bit.ly/2ysbOFE | bash -s -- ${BaseStepClass.fabricVersion()} ${BaseStepClass.fabricCAVersion()} -s -d`;
    execSync(commands);
  }

  @given(/bootstrap a Fabric network with CAs/, 'on-docker')
  public bootstrapFabricNetwork() {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh up -ca -i ${BaseStepClass.fabricVersion()} -cai ${BaseStepClass.fabricCAVersion()}`;
    execSync(commands);
  }

  @given(/create (.+) channel/, 'on-docker')
  public createChannel(channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh createChannel -c ${channelID}`;
    execSync(commands);
  }

  @given(/deploy (.+) for opssc on (.+)/, 'on-docker')
  public deployChaincodeForOpsSCToFabricNetwork(ccName: string, channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh deployCC -c ${channelID} -ccn ${ccName} -ccp ../../../chaincode/${ccName} -ccl go`;
    execSync(commands);
  }

  @given(/deploy (.+) as a dummy on (.+)/, 'on-docker')
  public deployChaincodeAsDummyToFabricNetwork(ccName: string, channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh deployCC -c ${channelID} -ccn ${ccName} -ccp ../asset-transfer-basic/chaincode-go -ccl go`;
    execSync(commands);
  }

  @given(/register orgs info for (.+) \(type: (system|application|ops)\) to opssc on (.+)/, 'on-docker')
  public registerOrgInfo(newChannelName: string, newChannelType: string, opsChannelName: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./registerNetworkInfoToOpsSC.sh ${opsChannelName} ${newChannelName} ${newChannelType}`;
    execSync(commands);
  }

  @given(/bootstrap opssc-(.*)s for initial orgs/, 'on-docker')
  public async bootstrapOpsSCServicesForDocker(service: 'api-server'|'agent') {
    const dockerComposeFileName = `docker-compose-opssc-${service}s.yaml`;
    const commands = `IMAGE_TAG=${BaseStepClass.opsSCImageTag()} docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/${dockerComposeFileName} up -d`;
    execSync(commands);

    for (let n = FabricNetworkDeploymentSteps.RETRY; n >= 0; n--) {
      await this.delay(10000);
      try {
        const response1 = await axios.get(`${this.getServiceEndpoint('docker', 'org1', service)}/healthz`);
        const response2 = await axios.get(`${this.getServiceEndpoint('docker', 'org2', service)}/healthz`);
        if (response1.status === 200 && response2.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail(`Fail to bootstrap opssc-${service}s`);
  }

  @given(/bootstrap a Fabric network with CAs/, 'on-k8s')
  public bootstrapTestNetworkK8s() {
    const commands = './network up';
    execSync(commands, {
      cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
      env: {
        TEST_NETWORK_CHAINCODE_BUILDER: 'k8s',
        ...process.env,
      }
    });
  }

  @given(/create (.+) channel/, 'on-k8s')
  public createChannelForTestNetworkK8s(channelID: string) {
    const commands = './network channel create';
    execSync(commands, {
      cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
      env: {
        TEST_NETWORK_CHANNEL_NAME: channelID,
        ...process.env,
      }
    });
  }

  @given(/put msp info into k8s/, 'on-k8s')
  public putAdminMSPInfoIntoK8s() {
    const orgList =['org1', 'org2'];
    for (const org of orgList) {
      const commandList = [
        './network rest-easy',
        `tar -C build/enrollments/${org}/users/${org}admin -cvf build/admin-msp.tar msp`,
        `kubectl -n test-network delete configmap ${org}-admin-msp || true`,
        `kubectl -n test-network create configmap ${org}-admin-msp --from-file=build/admin-msp.tar`,
        'rm build/admin-msp.tar'];
      for (const commands of commandList) {
        execSync(commands, {
          cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
        });
      }
    }
  }

  @given(/deploy (.+) for opssc on (.+)/, 'on-k8s')
  public deployChaincodeForOpsSCToFabricNetworkK8s(ccName: string, channelID: string) {
    const commands = `./network chaincode deploy ${ccName} ../../../../chaincode/${ccName}`;
    execSync(commands, {
      cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
      env: {
        TEST_NETWORK_CHANNEL_NAME: channelID,
        TEST_NETWORK_CHAINCODE_BUILDER: 'k8s',
      }
    });
  }

  @given(/register orgs info for (.+) \(type: (system|application|ops)\) to opssc on (.+)/, 'on-k8s')
  public registerOrgInfoForTestNetworkK8s(newChannelName: string, newChannelType: string, opsChannelName: string) {
    const commandsList =[
      `./network chaincode invoke ${BaseStepClass.CH_OPS_CC_NAME} '{"Args":["CreateChannel","${newChannelName}","${newChannelType}","[]"]}'`,
      `./network chaincode invoke ${BaseStepClass.CH_OPS_CC_NAME} '{"Args":["AddOrganization","${newChannelName}","Org1MSP"]}'`,
      `./network chaincode invoke ${BaseStepClass.CH_OPS_CC_NAME} '{"Args":["AddOrganization","${newChannelName}","Org2MSP"]}'`,
    ];
    for (const commands of commandsList) {
      execSync(commands, {
        cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
        env: {
          TEST_NETWORK_CHANNEL_NAME: opsChannelName,
          ...process.env,
        }
      });
    }
  }

  @given(/bootstrap opssc-(.*)s for initial orgs/, 'on-k8s')
  public async bootstrapOpsSCServicesForTestNetworkK8s(service: 'api-server'|'agent') {
    const orgList =['org1', 'org2'];
    for (const org of orgList) {
      const commandsList =[
        `helm -n test-network uninstall ${org}-opssc-${service} || true`,
        `helm upgrade -n test-network ${org}-opssc-${service} ../../../../opssc-${service}/charts/opssc-${service} -f ../../helm_values/${org}-opssc-${service}.yaml --install`
      ];
      for (const commands of commandsList) {
        execSync(commands, {
          cwd: BaseStepClass.TEST_NETWORK_K8S_PATH,
        });
      }
    }

    for (let n = FabricNetworkDeploymentSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getServiceEndpoint('k8s', orgList[0], service)}/healthz`);
        if (response.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail(`Fail to bootstrap opssc-${service}s`);
  }

  @given(/disable (.+) channel on opssc via opssc-api-server/)
  public async disableChannel(channelName: string) {
    await this.invokeChannelOpsFunc('UpdateChannelType', [channelName, 'disable']);
  }
}
