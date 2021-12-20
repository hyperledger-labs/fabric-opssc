/*
 * Copyright 2020-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { binding, given, then, when } from 'cucumber-tsflow';
import { expect } from 'chai';
import axios from 'axios';
import BaseStepClass from '../utils/base-step-class';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

@binding()
export class ChannelOpsSteps extends BaseStepClass {

  @given(/prepare org(3|4)/)
  public async prepareOrg(orgIndex: number) {
    let commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && IMAGE_TAG=${BaseStepClass.fabricCAVersion()} docker-compose -f docker/docker-compose-ca-org${orgIndex}.yaml up -d`;
    execSync(commands);

    await this.delay(3000);

    const org = `org${orgIndex}`;
    const ports = BaseStepClass.SERVICE_PORT_MAP[org];
    commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./registerEnroll.sh ${orgIndex} ca-${org} ${ports.ca} ${ports.orderer} ${ports.peer}`;
    if (Number(orgIndex) === 3) {
      // For preparing 2nd peer
      commands = `${commands} ${ports.peer + 1000}`;
    }
    execSync(commands);
  }

  @when(/launch nodes for org(3|4)/)
  public async launchNodesForOrg(orgIndex: number) {
    let commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ./fetchSystemConfigBlock.sh ${orgIndex}`;
    execSync(commands);

    commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && IMAGE_TAG=${BaseStepClass.fabricVersion()} docker-compose -f docker/docker-compose-orderer-org${orgIndex}.yaml up -d`;
    execSync(commands);

    commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && IMAGE_TAG=${BaseStepClass.fabricVersion()} docker-compose -f docker/docker-compose-peer-org${orgIndex}.yaml up -d`;
    execSync(commands);
  }

  @when(/bootstrap opssc-api-servers for org(3|4)/)
  public async bootstrapOpsSCAPIServersForOrg(orgIndex: number) {
    const dockerComposeFileName = `docker-compose-opssc-api-servers-org${orgIndex}.yaml`;
    const commands = `IMAGE_TAG=${BaseStepClass.opsSCImageTag()} docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/${dockerComposeFileName} up -d`;
    execSync(commands);

    for (let n = ChannelOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getAPIEndpoint(`org${orgIndex}`)}/healthz`);
        if (response.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail('Fail to bootstrap opssc-api-server');
  }

  @given(/bootstrap opssc-agents for org(3|4)/)
  public async bootstrapOpsSCAgents(orgIndex: number) {
    const dockerComposeFileName = `docker-compose-opssc-agents-org${orgIndex}.yaml`;
    const commands = `IMAGE_TAG=${BaseStepClass.opsSCImageTag()} docker-compose -f ${BaseStepClass.TEST_NETWORK_PATH}/docker/${dockerComposeFileName} up -d`;
    execSync(commands);

    for (let n = ChannelOpsSteps.RETRY; n >= 0; n--) {
      await this.delay(10000);
      try {
        const response = await axios.get(`${this.getAgentServiceEndpoint(`org${orgIndex}`)}/healthz`);
        if (response.status === 200) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect.fail('Fail to bootstrap opssc-agent');
  }

  @when(/(.+) requests a proposal to add org for (org3|org4) to (.+) via opssc-api-server/)
  public async requestChannelUpdateProposalToAddOrg(creatorOrg: string, targetOrg: string, channelID: string) {
    const [proposalID, description, opsProfile] = this.createProposalParametersToAddOrg(targetOrg, channelID);
    const proposal = {
      channelID: channelID,
      description: description,
      opsProfile: opsProfile
    };

    // console.log(proposal) // For debug
    const _response = await axios.post(`${this.getAPIEndpoint(creatorOrg)}/api/v1/channel/proposals/${proposalID}`,
      {
        proposal: proposal
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    // Wait to stabilize subsequent operations
    await this.delay();
  }

  @when(/(.+) requests a proposal to create (.+) via opssc-api-server/)
  public async requestChannelUpdateProposalToCreateChannel(org: string, channelID: string) {
    const [proposalID, description, opsProfile] = this.createProposalParametersToCreateChannel(channelID);
    const proposal = {
      channelID: channelID,
      description: description,
      opsProfile: opsProfile,
      action: 'create'
    };

    // console.log(proposal) // For debug
    const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/channel/proposals/${proposalID}`,
      {
        proposal: proposal
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    // Wait to stabilize subsequent operations
    await this.delay();
  }

  @when(/(.+) approves the proposal to add org for (org3|org4) to (.+) via opssc-api-server/)
  public async approveChannelUpdateProposalToAddOrg(creatorOrg: string, targetOrg: string, channelID: string) {
    const [proposalID] = this.createProposalParametersToAddOrg(targetOrg, channelID);
    const _response = await axios.post(`${this.getAPIEndpoint(creatorOrg)}/api/v1/channel/proposals/${proposalID}/vote`,
      {
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
  }

  @when(/(.+) approves the proposal to create (.+) via opssc-api-server/)
  public async approveChannelUpdateProposalToCreateChannel(org: string, channelID: string) {
    const [proposalID] = this.createProposalParametersToCreateChannel(channelID);
    const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/channel/proposals/${proposalID}/vote`,
      {
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
  }

  @then(/the proposal to add org for (org3|org4) to (.+) should be committed/)
  public async checkGetProposal(targetOrg: string, channelID: string) {
    const [proposalID, description] = this.createProposalParametersToAddOrg(targetOrg, channelID);

    let response = null;
    for (let n = ChannelOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        response = await axios.get(`${this.getAPIEndpoint()}/api/v1/utils/queryTransaction?channelID=ops-channel&ccName=channel_ops&func=GetProposal&args=["${proposalID}"]`);
        const proposal = response.data;
        // console.log(proposal); // For debug
        if (proposal.status === 'committed') {
          break;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
    expect(response).to.not.equals(null);
    if (response === null) return;
    expect(response.status).to.equals(200);
    expect(response.data).to.not.equals(null);
    expect(response.data.description).to.equals(description);
    expect(response.data.creator).to.equals('Org1MSP');
    expect(response.data.status).to.equals('committed');
    expect(response.data.artifacts.configUpdate).to.not.equals(null);
    expect(response.data.artifacts.signatures).to.not.equals(null);
    expect(response.data.artifacts.signatures.length).to.not.equals(2);
  }

  @then(/(\d+) chaincodes should be installed on org(\d+)'s peer(\d+)/)
  public async checkInstalledChaincodesOnPeer(ccNum: number, orgIndex: number, peerIndex: number) {
    const ports = BaseStepClass.SERVICE_PORT_MAP[`org${orgIndex}`];
    const envs = ['PATH=$PWD/../bin:$PATH',
      'FABRIC_CFG_PATH=$PWD/../config/',
      'CORE_PEER_TLS_ENABLED=true',
      `CORE_PEER_MSPCONFIGPATH=$PWD/organizations/peerOrganizations/org${orgIndex}.example.com/users/Admin@org${orgIndex}.example.com/msp`,
      `CORE_PEER_LOCALMSPID=Org${orgIndex}MSP`,
      `CORE_PEER_TLS_ROOTCERT_FILE=$PWD/organizations/peerOrganizations/org${orgIndex}.example.com/peers/peer${peerIndex}.org${orgIndex}.example.com/tls/ca.crt`,
      `CORE_PEER_ADDRESS=localhost:${ports.peer + peerIndex * 1000}`];
    const commands = `cd ${BaseStepClass.TEST_NETWORK_PATH} && ${envs.join(' ')} peer lifecycle chaincode queryinstalled --output json`;
    const result = execSync(commands);
    const installed = JSON.parse(result.toString());
    expect(installed.installed_chaincodes).to.not.equals(null);
    expect(installed.installed_chaincodes.length).to.equals(ccNum);
  }

  @then(/channel (.+) should be created/)
  public async newChannelCreated(channelID: string) {
    let response = null;
    for (let n = ChannelOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        response = await axios.get(`${this.getAPIEndpoint()}/api/v1/channel/getChannel?channelID=${channelID}`);
        const channel = response.data;
        // console.log(channel); // For debug
        if (channel !== null) {
          break;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
    }
  }

  private createOpsProfileToAddOrg(org: string, channelID: string, peerPort: number, ordererPort: number): any {
    const mspID = `${this.capitalizeFirstLetter(org)}MSP`;
    let orgType = 'Application|Orderer';
    if (channelID === 'system-channel') {
      orgType = 'Consortiums|Orderer';
    }

    return [
      {
        'Command': 'set-org',
        'Parameters': {
          'OrgType': orgType,
          'Org': {
            'Name': mspID,
            'ID': mspID,
            'MSP': {
              'RootCerts': [
                this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/cacerts`)
              ],
              'TLSRootCerts': [
                this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/tlscacerts`)
              ],
              'NodeOUs': {
                'Enable': true,
                'ClientOUIdentifier': {
                  'OrganizationalUnitIdentifier': 'client',
                  'Certificate': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/cacerts`)
                },
                'PeerOUIdentifier': {
                  'OrganizationalUnitIdentifier': 'peer',
                  'Certificate': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/cacerts`)
                },
                'AdminOUIdentifier': {
                  'OrganizationalUnitIdentifier': 'admin',
                  'Certificate': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/cacerts`)
                },
                'OrdererOUIdentifier': {
                  'OrganizationalUnitIdentifier': 'orderer',
                  'Certificate': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/msp/cacerts`)
                }
              }
            },
            'Policies': {
              'Readers': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.admin', '${mspID}.peer', '${mspID}.client')`
              },
              'Writers': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.admin', '${mspID}.client')`
              },
              'OrderingReaders': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.admin', '${mspID}.orderer')`
              },
              'OrderingWriters': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.admin', '${mspID}.orderer')`
              },
              'Admins': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.admin')`
              },
              'Endorsement': {
                'Type': 'Signature',
                'Rule': `OR('${mspID}.peer')`
              }
            },
            'AnchorPeers': [
              {
                'Host': `peer0.${org}.example.com`,
                'Port': peerPort
              }
            ],
            'OrdererEndpoints': [
              `orderer0.${org}.example.com:${ordererPort}`
            ]
          }
        }
      }
    ];
  }

  private createOpsProfileToAddConsenter(org: string, ordererPort: number): any {
    return [
      {
        'Command': 'set-consenter',
        'Parameters': {
          'Consenter': {
            'Host': `orderer0.${org}.example.com`,
            'Port': ordererPort,
            'ClientTLSCert': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/orderers/orderer0.${org}.example.com/tls/server.crt`),
            'ServerTLSCert': this.readSingleFileOnThePath(`${BaseStepClass.TEST_NETWORK_PATH}/organizations/peerOrganizations/${org}.example.com/orderers/orderer0.${org}.example.com/tls/server.crt`)
          }
        }
      }
    ];
  }

  private createOpsProfileToCreateChannel(): any {
    return {
      'Consortium': 'SampleConsortium',
      'Application': {
        'Capabilities': [
          'V2_0'
        ],
        'Policies': {
          'Readers': {
            'Type': 'ImplicitMeta',
            'Rule': 'ANY Readers'
          },
          'Writers': {
            'Type': 'ImplicitMeta',
            'Rule': 'ANY Writers'
          },
          'Admins': {
            'Type': 'ImplicitMeta',
            'Rule': 'ANY Admins'
          },
          'LifecycleEndorsement': {
            'Type': 'ImplicitMeta',
            'Rule': 'MAJORITY Endorsement'
          },
          'Endorsement': {
            'Type': 'ImplicitMeta',
            'Rule': 'MAJORITY Endorsement'
          }
        },
        'Organizations': [
          'Org1MSP',
          'Org2MSP'
        ]
      }
    };
  }

  private readSingleFileOnThePath(fileOrDirPath: string): string {
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

  private createProposalParametersToAddOrg(targetOrg: string, channelID: string): [string, string, any] {
    const ports = BaseStepClass.SERVICE_PORT_MAP[targetOrg];
    const opsProfile = this.createOpsProfileToAddOrg(targetOrg, channelID, ports.peer, ports.orderer).concat(this.createOpsProfileToAddConsenter(targetOrg, ports.orderer));
    return [`proposal_to_add_${targetOrg}_to_${channelID}_${ChannelOpsSteps.SUFFIX}`,
      `Add ${targetOrg} to ${channelID}`,
      opsProfile];
  }

  private createProposalParametersToCreateChannel(channelID: string): [string, string, any] {
    return [`proposal_to_create_${channelID}_${ChannelOpsSteps.SUFFIX}`,
      `Create ${channelID}`,
      this.createOpsProfileToCreateChannel()];
  }

  private capitalizeFirstLetter(source: string): string {
    return source.charAt(0).toUpperCase() + source.slice(1);
  }

}
