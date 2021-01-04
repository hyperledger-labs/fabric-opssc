/*
 * Copyright 2020 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { after, before, binding, given} from 'cucumber-tsflow';
import { expect } from 'chai';
import axios from 'axios';
import { execSync } from 'child_process';
import BaseStepClass from '../utils/base-step-class';

@binding()
export class FabricNetworkDeploymentSteps extends BaseStepClass {

  @before('on-docker')
  @after('on-docker')
  public cleanupFabricNetwork() {
    let commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents-org3.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-agents-org4.yaml down --remove-orphans`;
    execSync(commands);

    commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers-org3.yaml -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/docker-compose-opssc-api-servers-org4.yaml down --remove-orphans`;
    execSync(commands);

    commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh down`;
    execSync(commands);
  }

  @given(/bootstrap a Fabric network with CAs/)
  public bootstrapFabricNetwork() {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh up -ca -i ${FabricNetworkDeploymentSteps.FABRIC_VERSION} -cai ${FabricNetworkDeploymentSteps.FABRIC_CA_VERSION}`;
    execSync(commands);
  }

  @given(/create (.+) channel/)
  public createChannel(channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh createChannel -c ${channelID}`;
    execSync(commands);
  }


  @given(/deploy (.+) for opssc on (.+)/)
  public deployChaincodeForOpsSCToFabricNetwork(ccName: string, channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh deployCC -c ${channelID} -ccn ${ccName} -ccp ../../../chaincode/${ccName} -ccl go`;
    execSync(commands);
  }

  @given(/deploy (.+) as a dummy on (.+)/)
  public deployChaincodeAsDummyToFabricNetwork(ccName: string, channelID: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./network.sh deployCC -c ${channelID} -ccn ${ccName} -ccp ../asset-transfer-basic/chaincode-go -ccl go`;
    execSync(commands);
  }

  @given(/register orgs info for (.+) \(type: (system|application|ops)\) to opssc on (..+)/)
  public registerOrgInfo(newChannelName: string, newChannelType: string, opsChannelName: string) {
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./registerNetworkInfoToOpsSC.sh ${opsChannelName} ${newChannelName} ${newChannelType}`;
    execSync(commands);
  }

  @given(/bootstrap opssc-api-servers for initial orgs/)
  public async bootstrapOpsSCAPIServers() {
    const dockerComposeFileName = 'docker-compose-opssc-api-servers.yaml';
    const commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/${dockerComposeFileName} up -d`;
    execSync(commands);

    for (let n = FabricNetworkDeploymentSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getAPIEndpoint()}/healthz`);
        if (response.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail('Fail to bootstrap opssc-api-servers');
  }

  @given(/bootstrap opssc-agents for initial orgs/)
  public async bootstrapOpsSCAgents() {
    const dockerComposeFileName = 'docker-compose-opssc-agents.yaml';
    const commands = `docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/${dockerComposeFileName} up -d`;
    execSync(commands);

    for (let n = FabricNetworkDeploymentSteps.RETRY; n >= 0; n--) {
      await this.delay(10000);
      try {
        const response1 = await axios.get(`${this.getAgentServiceEndpoint('org1')}/healthz`);
        const response2 = await axios.get(`${this.getAgentServiceEndpoint('org2')}/healthz`);
        if (response1.status === 200 && response2.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail('Fail to bootstrap opssc-agents');
  }
}
