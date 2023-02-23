/*
 * Copyright 2020-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { before, binding, then, when } from 'cucumber-tsflow';
import { expect } from 'chai';
import axios from 'axios';
import BaseStepClass from '../utils/base-step-class';
import { execSync } from 'child_process';

type TaskType = 'vote' | 'acknowledge' | 'commit';
type TaskTypeAsPastParticiple = 'voted' | 'acknowledged' | 'committed';

const TaskTypeFuncs = {
  toSingular: function(type: TaskTypeAsPastParticiple): TaskType {
    switch (type) {
      case 'voted':
        return 'vote';
      case 'acknowledged':
        return 'acknowledge';
      case 'committed':
        return 'commit';
    }
  }
};


@binding()
export class ChaincodeOpsSteps extends BaseStepClass {

  @before('on-k8s')
  public beforeK8sScenarios() {
    this.environment = 'k8s';
  }

  @when(/(.+) requests a proposal to deploy the chaincode \(name: (.+), seq: (\d+), channel: (.+)\) based on (basic|private) (golang|javascript|typescript|ccaas|k8s) template via opssc-api-server/)
  private async requestChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, ccTemplate: string, lang: string) {

    const [repository, commitID, pathToSourceFiles, validationParameter, collections] = this.createCCParameters(ccTemplate, lang);
    const proposal = {
      ID: `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`,
      channelID: channelID,
      chaincodeName: `${ccName}-${ChaincodeOpsSteps.SUFFIX}`,
      chaincodePackage: {
        repository: repository,
        pathToSourceFiles: pathToSourceFiles,
        commitID: commitID,
        type: lang
      },
      chaincodeDefinition:
      {
        sequence: sequence,
        initRequired: false,
        validationParameter: validationParameter,
        collections: collections
      }
    };
    // console.log(proposal) // For debug
    const _response = await axios.post(`${this.getServiceEndpoint(this.environment, org)}/api/v1/chaincode/proposals/${proposal.ID}`,
      {
        proposal: proposal
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
  }

  @then(/(.+) fails to request a proposal to deploy the chaincode \(name: (.+), seq: (\d+), channel: (.+)\) based on (basic|private) (golang|javascript|typescript|ccaas|k8s) template via opssc-api-server/)
  public async failToRequestChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, ccTemplate: string, lang: string) {
    try {
      await this.requestChaincodeDeploymentProposal(org, ccName, sequence, channelID, ccTemplate, lang);
    } catch (error) {
      // Should fail
      expect(error.response).to.not.equals(null);
      expect(error.response.status).to.equals(500);
      expect(error.response.data.message).to.includes('proposal is not accepted by the channel');
      return;
    }
    expect.fail('the request should fail');
  }

  private createCCParameters(ccTemplate: string, lang: string): [string, string, string, string, string|undefined] {
    switch (this.environment) {
      case 'k8s':
        return this.createCCParametersForK8s(ccTemplate, lang);
      default:
        return this.createCCParametersForDocker(ccTemplate, lang);
    }
  }

  private createCCParametersForDocker(ccTemplate: string, lang: string): [string, string, string, string, string|undefined] {
    const repository = process.env.IT_REMOTE_CC_REPO || 'github.com/hyperledger-labs/fabric-opssc';
    const commitID = process.env.IT_REMOTE_COMMIT_ID || 'main';
    switch (ccTemplate) {
      case 'basic':
        return [repository, commitID, ...this.createBasicCCParameters(lang)];
      case 'private':
        return [repository, commitID, ...this.createPrivateCCParameters(lang)];
      default:
        expect.fail(`currently, ${ccTemplate} is not supported`);
    }
  }

  private createCCParametersForK8s(ccTemplate: string, lang: string): [string, string, string, string, string|undefined] {
    const repository = process.env.IT_K8S_REMOTE_CC_REPO || 'github.com/hyperledger-labs/fabric-opssc';
    const commitID = process.env.IT_K8S_REMOTE_COMMIT_ID || 'main';
    switch (ccTemplate) {
      case 'basic':
        return [repository, commitID, ...this.createBasicCCParametersForK8s(lang)];
      default:
        expect.fail(`currently, ${ccTemplate} is not supported`);
    }
  }

  private createBasicCCParametersForK8s(lang: string): [string, string, string|undefined] {
    const basePath = 'sample-environments/fabric-samples/asset-transfer-basic';
    const collectionsBase64 = undefined;
    let validationParameterBase64;
    switch (lang) {
      case 'ccaas':
      case 'k8s':
        validationParameterBase64 = Buffer.from('/Channel/Application/Endorsement').toString('base64'); // PEER-CLI-SYNTAX
        return [process.env.IT_K8S_REMOTE_BASIC_CC_PATH || `${basePath}/chaincode-java`, validationParameterBase64, collectionsBase64];
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }
  }

  private createBasicCCParameters(lang: string): [string, string, string|undefined] {
    const basePath = 'sample-environments/fabric-samples/asset-transfer-basic';
    const collectionsBase64 = undefined;
    let validationParameter;
    let validationParameterBase64;
    switch (lang) {
      case 'golang':
        validationParameterBase64 = Buffer.from('/Channel/Application/Endorsement').toString('base64'); // PEER-CLI-SYNTAX
        return [process.env.IT_REMOTE_BASIC_GO_CC_PATH || `${basePath}/chaincode-go`, validationParameterBase64, collectionsBase64];
      case 'javascript':
        validationParameter = JSON.stringify( // SDK-SYNTAX
          {
            identities: [
              { role: { name: 'peer', mspId: 'Org1MSP' } },
              { role: { name: 'peer', mspId: 'Org2MSP' } }
            ],
            policy: {
              '2-of': [{ 'signed-by': 0 }, { 'signed-by': 1 }]
            }
          });
        validationParameterBase64 = Buffer.from(validationParameter).toString('base64');
        return [process.env.IT_REMOTE_BASIC_JS_CC_PATH || `${basePath}/chaincode-javascript`, validationParameterBase64, collectionsBase64];
      case 'typescript':
        validationParameter = 'AND(\'Org1MSP.peer\', \'Org2MSP.peer\')'; // PEER-CLI-SYNTAX
        validationParameterBase64 = Buffer.from(validationParameter).toString('base64');
        return [process.env.IT_REMOTE_BASIC_TS_CC_PATH || `${basePath}/chaincode-typescript`, validationParameterBase64, collectionsBase64];
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }
  }

  private createPrivateCCParameters(lang: string): [string, string, string|undefined] {
    const basePath = 'sample-environments/fabric-samples/asset-transfer-private-data';
    const validationParameter = JSON.stringify( // SDK-SYNTAX
      {
        identities: [
          { role: { name: 'peer', mspId: 'Org1MSP' } },
          { role: { name: 'peer', mspId: 'Org2MSP' } }
        ],
        policy: {
          '1-of': [{ 'signed-by': 0 }, { 'signed-by': 1 }]
        }
      });
    const validationParameterBase64 = Buffer.from(validationParameter).toString('base64');
    const collections = JSON.stringify(
      [
        {
          name: 'assetCollection',
          member_orgs_policy: 'OR(\'Org1MSP.member\', \'Org2MSP.member\')', // PEER-CLI-SYNTAX
          required_peer_count: 1,
          maximum_peer_count: 1,
          block_to_live: 1000000,
          member_only_read: true,
          member_only_write: true
        },
        {
          name: 'Org1MSPPrivateCollection',
          member_orgs_policy: { // FABRIC-SYNTAX
            identities: [{
              principal_classification: 0,
              principal: {
                msp_identifier: 'Org1MSP',
                role: 'MEMBER'
              },
            }],
            rule: {
              n_out_of: {
                n: 1,
                rules: [{
                  signed_by: 0
                }]
              }
            }
          },
          required_peer_count: 0,
          maximum_peer_count: 1,
          block_to_live: 3,
          member_only_read: true,
          member_only_write: false,
          endorsement_policy: {
            signature_policy: 'OR(\'Org1MSP.member\')' // PEER-CLI-SYNTAX
          }
        },
        {
          name: 'Org2MSPPrivateCollection',
          member_orgs_policy: 'OR(\'Org2MSP.member\')', // PEER-CLI-SYNTAX
          required_peer_count: 0,
          maximum_peer_count: 1,
          block_to_live: 3,
          member_only_read: true,
          member_only_write: false,
          endorsement_policy: {
            signature_policy: { // SDK-SYNTAX
              identities: [
                { role: { name: 'member', mspId: 'Org2MSP' } }
              ],
              policy: {
                '1-of': [{ 'signed-by': 0 }]
              }
            }
          }
        }
      ]
    );
    const collectionsBase64 = Buffer.from(collections).toString('base64');
    switch (lang) {
      case 'golang':
        return [process.env.IT_REMOTE_PRIVATE_GO_CC_PATH || `${basePath}/chaincode-go`, validationParameterBase64, collectionsBase64];
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }
  }

  @when(/(.+) votes (for|against) the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with opssc-api-server/)
  public async voteChaincodeDeploymentProposal(org: string, vote: string, ccName: string, sequence: number, channelID: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    const _response = await axios.post(`${this.getServiceEndpoint(this.environment, org)}/api/v1/chaincode/proposals/${proposalID}/vote`,
      {
        updateRequest: {
          status: (vote == 'for') ? 'agreed' : 'disagreed'
        }
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
  }

  @when(/(.+) withdraws the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with opssc-api-server/)
  public async withdrawChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    const _response = await axios.post(`${this.getServiceEndpoint(this.environment,org) }/api/v1/chaincode/proposals/${proposalID}/withdraw`,
      {
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
  }

  @then(/the proposal status for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) should be (.+)/)
  public async checkStatusForChaincodeDeploymentProposal(ccName: string, sequence: number, channelID: string, status: string) {

    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    let proposal;
    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/chaincode/proposals/${proposalID}`);
        proposal = response.data;
        // console.log(proposal); // For debug
        if (proposal.status === status) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
      // eslint-disable-next-line no-console
      console.log('.');
    }
    expect.fail(`The proposal has not been ${status}.\nLast acquired proposal: ${JSON.stringify(proposal, null, 2)}`);
  }

  @then(/the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) should be (.+) \(with (.+)\) by (\d+) or more orgs/)
  public async checkTaskStatusForChaincodeDeploymentProposal(ccName: string, sequence: number, channelID: string, taskPP: TaskTypeAsPastParticiple, taskStatus: string, numOfOrgsVoting: number) {

    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    let voteHistories;
    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay(10000);
      try {
        const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/chaincode/proposals/${proposalID}/histories?taskID=${TaskTypeFuncs.toSingular(taskPP)}`);
        voteHistories = response.data;
        // console.log(voteHistories); // For debug
        if (Object.keys(voteHistories).length >= numOfOrgsVoting) {
          for (const history of voteHistories) {
            expect(history.status).to.equals(taskStatus);
          }
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
      // eslint-disable-next-line no-console
      console.log('.');
    }
    expect.fail(`The proposal has not been ${taskPP} by ${numOfOrgsVoting} orgs.\nLast acquired histories: ${JSON.stringify(voteHistories, null, 2)}`);
  }

  @then(/chaincode \(name: (.+), seq: (\d+), channel: (.+)\) should be committed over the fabric network/)
  public async isCommittedChaincode(ccName: string, sequence: number, channelID: string) {

    let committed;
    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/chaincode/queryChaincodeDefinition?channelID=${channelID}&chaincodeName=${ccName}-${ChaincodeOpsSteps.SUFFIX}`);
        committed = response.data;
        // console.log(committed); // For debug
        if (committed !== null && Number(committed.sequence) === sequence) {
          expect(committed).to.be.an('object');
          for (const approval of committed.approvals) {
            expect(approval).to.equals(true);
          }
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
      // eslint-disable-next-line no-console
      console.log('.');
    }
    expect.fail(`The chaincode definition has not been committed.\nLast acquired committed: ${JSON.stringify(committed, null, 2)}`);
  }

  @then(/chaincode \(name: (.+), channel: (.+)\) should be set the collections for private template/)
  public async checkCollectionsForAssetTransferPrivate(ccName: string, channelID: string) {
    const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/chaincode/queryChaincodeDefinition?channelID=${channelID}&chaincodeName=${ccName}-${ChaincodeOpsSteps.SUFFIX}`);
    expect(response.data.collections).to.not.equals(null);
    expect(response.data.collections).to.deep.equals(this.createExpectedCollections());
  }

  @then(/(\d+) chaincodes should be committed on (.+)/)
  public async checkCommittedChaincodeNum(num: number, channelID: string) {

    let result;
    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/chaincode/queryChaincodeDefinitions?channelID=${channelID}`);
        result = response.data;
        // console.log(committed); // For debug
        if (result !== null && result.chaincode_definitions.length === num) {
          return;
        }
      } catch (error) {
        // console.log(error.message); // For debug
      }
      // eslint-disable-next-line no-console
      console.log('.');
    }
    expect.fail(`The chaincode definitions is not ${num}.\nLast acquired: ${JSON.stringify(result, null, 2)}`);
  }

  @then(/chaincode \(name: (.+), channel: (.+)\) based on basic should be able to register the asset \(ID: (.+)\) by invoking CreateAsset func/)
  public async canInvokeAssetTransferBasic(ccName: string, channelID: string, assetID: string) {
    const response = await axios.post(`${this.getServiceEndpoint(this.environment) }/api/v1/utils/invokeTransaction`,
      {
        channelID: channelID,
        ccName: `${ccName}-${ChaincodeOpsSteps.SUFFIX}`,
        func: 'CreateAsset',
        args: [assetID, 'blue', '5', 'Tomoko', '300'],
      },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );
    expect(response.status).to.equals(200);
  }

  @when(/consortium sets max malicious orgs in votes \(number: (\d+)\)/)
  public async setMaxMaliciousOrgsInVotes(number: number) {
    const status = await this.invokeChaincodeOpsFunc('SetMaxMaliciousOrgsInVotes', [`${number}`]);
    expect(status).to.equals(200);
  }

  @when(/consortium unsets max malicious orgs in votes/)
  public async unsetMaxMaliciousOrgsInVotes() {
    const status = await this.invokeChaincodeOpsFunc('UnsetMaxMaliciousOrgsInVotes', []);
    expect(status).to.equals(200);
  }

  @when(/(start|stop) opssc-agents for org(1|2|3|4)/, 'on-k8s')
  public async controlOpsSCAgents(action: 'start'|'stop', orgIndex: number) {
    const replicas = action === 'start' ? 1 : 0;
    const commands = `kubectl scale -n test-network --replicas=${replicas} deployment/org${orgIndex}-opssc-agent`;
    execSync(commands);

    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay(10000);
      try {
        const response = await axios.get(`${this.getServiceEndpoint('k8s', `org${orgIndex}`, 'agent')}/healthz`);
        switch (action) {
          case 'start':
            if (response.status === 200) {
              return;
            }
            break;
          case 'stop':
            if (response.status !== 200) {
              return;
            }
            break;
          default:
            expect.fail(`Unexpected action: ${action}`);
        }
      } catch (error) {
        if (action === 'stop') {
          const commands = `kubectl -n test-network get pods | grep org${orgIndex}-opssc-agent`;
          let result;
          try {
            result = execSync(commands).toString();
          } catch (err) {
            result = err.stderr.toString();
          }
          // eslint-disable-next-line no-console
          console.log(result);
          if (result.length === 0) {
            return;
          }
        }
        // console.log(error.message); // For debug
      }
    }
    expect.fail(`Fail to ${action} opssc-agent`);
  }

  @then(/(.+) fails to approve the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with an error \((.+)\)/)
  public async failToApproveChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, errorMessage: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    try {
      const _response = await axios.post(`${this.getServiceEndpoint(this.environment,org) }/api/v1/chaincode/proposals/${proposalID}/vote`,
        {
          updateRequest: {
          }
        },
        {
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );
    } catch (error) {
      expect(error.response).to.not.equals(null);
      expect(error.response.status).to.equals(500);
      expect(error.response.data.message).to.includes(errorMessage);
      return;
    }
    expect.fail('the request should fail');
  }

  @then(/(.+) fails to withdraw the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with an error \((.+)\)/)
  public async failToWithdrawChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, errorMessage: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    try {
      const _response = await axios.post(`${this.getServiceEndpoint(this.environment,org) }/api/v1/chaincode/proposals/${proposalID}/withdraw`,
        {
        },
        {
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );
    } catch (error) {
      expect(error.response).to.not.equals(null);
      expect(error.response.status).to.equals(500);
      expect(error.response.data.message).to.includes(errorMessage);
      return;
    }
    expect.fail('the request should fail');
  }

  @then(/chaincode \(name: (.+), channel: (.+)\) based on basic (golang|javascript|typescript|ccaas|k8s) should be able to get the asset \(ID: (.+)\) by querying ReadAsset func/)
  public async canQueryAssetTransferBasic(ccName: string, channelID: string, lang: string, assetID: string) {
    const response = await axios.get(`${this.getServiceEndpoint(this.environment) }/api/v1/utils/queryTransaction?channelID=${channelID}&ccName=${ccName}-${ChaincodeOpsSteps.SUFFIX}&func=ReadAsset&args=["${assetID}"]`);
    expect(response.status).to.equals(200);
    expect(response.data).to.not.equals(null);
    switch (lang) {
      case 'golang':
      case 'typescript':
        expect(response.data.Color).to.equals('blue');
        expect(response.data.Size).to.equals(5);
        expect(response.data.Owner).to.equals('Tomoko');
        expect(response.data.AppraisedValue).to.equals(300);
        break;
      case 'javascript':
        expect(response.data.Color).to.equals('blue');
        expect(response.data.Size).to.equals('5');
        expect(response.data.Owner).to.equals('Tomoko');
        expect(response.data.AppraisedValue).to.equals('300');
        break;
      case 'ccaas':
      case 'k8s':
        expect(response.data.color).to.equals('blue');
        expect(response.data.size).to.equals(5);
        expect(response.data.owner).to.equals('Tomoko');
        expect(response.data.appraisedValue).to.equals(300);
        break;
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }
  }

  private createExpectedCollections(): any {
    return {
      'config': [
        {
          'static_collection_config': {
            'name': 'assetCollection',
            'member_orgs_policy': {
              'signature_policy': {
                'rule': {
                  'n_out_of': {
                    'n': 1,
                    'rules': [
                      {
                        'signed_by': 0
                      },
                      {
                        'signed_by': 1
                      }
                    ]
                  }
                },
                'identities': [
                  {
                    'principal': 'CgdPcmcxTVNQEAA='
                  },
                  {
                    'principal': 'CgdPcmcyTVNQEAA='
                  }
                ]
              }
            },
            'required_peer_count': 1,
            'maximum_peer_count': 1,
            'block_to_live': '1000000',
            'member_only_read': true,
            'member_only_write': true,
            'endorsement_policy': {}
          }
        },
        {
          'static_collection_config': {
            'name': 'Org1MSPPrivateCollection',
            'member_orgs_policy': {
              'signature_policy': {
                'rule': {
                  'n_out_of': {
                    'n': 1,
                    'rules': [
                      {
                        'signed_by': 0
                      }
                    ]
                  }
                },
                'identities': [
                  {
                    'principal': 'CgdPcmcxTVNQEAA='
                  }
                ]
              }
            },
            'maximum_peer_count': 1,
            'block_to_live': '3',
            'member_only_read': true,
            'endorsement_policy': {
              'signature_policy': {
                'rule': {
                  'n_out_of': {
                    'n': 1,
                    'rules': [
                      {
                        'signed_by': 0
                      }
                    ]
                  }
                },
                'identities': [
                  {
                    'principal': 'CgdPcmcxTVNQEAA='
                  }
                ]
              }
            }
          }
        },
        {
          'static_collection_config': {
            'name': 'Org2MSPPrivateCollection',
            'member_orgs_policy': {
              'signature_policy': {
                'rule': {
                  'n_out_of': {
                    'n': 1,
                    'rules': [
                      {
                        'signed_by': 0
                      }
                    ]
                  }
                },
                'identities': [
                  {
                    'principal': 'CgdPcmcyTVNQEAA='
                  }
                ]
              }
            },
            'maximum_peer_count': 1,
            'block_to_live': '3',
            'member_only_read': true,
            'endorsement_policy': {
              'signature_policy': {
                'rule': {
                  'n_out_of': {
                    'n': 1,
                    'rules': [
                      {
                        'signed_by': 0
                      }
                    ]
                  }
                },
                'identities': [
                  {
                    'principal': 'CgdPcmcyTVNQEAA='
                  }
                ]
              }
            }
          }
        }
      ]
    };
  }
}
