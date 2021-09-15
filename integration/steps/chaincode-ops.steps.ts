/*
 * Copyright 2020-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { binding, then, when } from 'cucumber-tsflow';
import { expect } from 'chai';
import axios from 'axios';
import BaseStepClass from '../utils/base-step-class';

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

  @when(/(.+) requests a proposal to deploy the chaincode \(name: (.+), seq: (\d+), channel: (.+)\) based on basic (golang|javascript|typescript) template via opssc-api-server/)
  public async requestChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, lang: string) {
    const repository = process.env.IT_REMOTE_CC_REPO || 'github.com/hyperledger-labs/fabric-opssc';
    const commitID = process.env.IT_REMOTE_COMMIT_ID || 'main';
    let pathToSourceFiles;
    switch (lang) {
      case 'golang':
        pathToSourceFiles = process.env.IT_REMOTE_BASIC_GO_CC_PATH || 'sample-environments/fabric-samples/asset-transfer-basic/chaincode-go';
        break;
      case 'javascript':
        pathToSourceFiles = process.env.IT_REMOTE_BASIC_JS_CC_PATH || 'sample-environments/fabric-samples/asset-transfer-basic/chaincode-javascript';
        break;
      case 'typescript':
        pathToSourceFiles = process.env.IT_REMOTE_BASIC_TS_CC_PATH || 'sample-environments/fabric-samples/asset-transfer-basic/chaincode-typescript';
        break;
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }

    const proposal = {
      ID: `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`,
      channelID: channelID,
      chaincodeName: `${ccName}_${ChaincodeOpsSteps.SUFFIX}`,
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
        validationParameter: Buffer.from('/Channel/Application/Endorsement').toString('base64'),
      }
    };
    // console.log(proposal) // For debug
    const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/chaincode/proposals/${proposal.ID}`,
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

  @when(/(.+) votes (for|against) the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with opssc-api-server/)
  public async voteChaincodeDeploymentProposal(org: string, vote: string, ccName: string, sequence: number, channelID: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/chaincode/proposals/${proposalID}/vote`,
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
    const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/chaincode/proposals/${proposalID}/withdraw`,
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
        const response = await axios.get(`${this.getAPIEndpoint()}/api/v1/chaincode/proposals/${proposalID}`);
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
      await this.delay();
      try {
        const response = await axios.get(`${this.getAPIEndpoint()}/api/v1/chaincode/proposals/${proposalID}/histories?taskID=${TaskTypeFuncs.toSingular(taskPP)}`);
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
        const response = await axios.get(`${this.getAPIEndpoint()}/api/v1/chaincode/queryChaincodeDefinition?channelID=${channelID}&chaincodeName=${ccName}_${ChaincodeOpsSteps.SUFFIX}`);
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

  @then(/(\d+) chaincodes should be committed on (.+)/)
  public async checkCommittedChaincodeNum(num: number, channelID: string) {

    let result;
    for (let n = ChaincodeOpsSteps.RETRY; n >= 0; n--) {
      await this.delay();
      try {
        const response = await axios.get(`${this.getAPIEndpoint()}/api/v1/chaincode/queryChaincodeDefinitions?channelID=${channelID}`);
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
    const response = await axios.post(`${this.getAPIEndpoint()}/api/v1/utils/invokeTransaction`,
      {
        channelID: channelID,
        ccName: `${ccName}_${ChaincodeOpsSteps.SUFFIX}`,
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

  @then(/(.+) fails to approve the proposal for chaincode \(name: (.+), seq: (\d+), channel: (.+)\) with an error \((.+)\)/)
  public async failToApproveChaincodeDeploymentProposal(org: string, ccName: string, sequence: number, channelID: string, errorMessage: string) {
    const proposalID = `proposal_cc_deployment_${ccName}_${ChaincodeOpsSteps.SUFFIX}_on_${channelID}_seq_${sequence}`;
    try {
      const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/chaincode/proposals/${proposalID}/vote`,
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
      const _response = await axios.post(`${this.getAPIEndpoint(org)}/api/v1/chaincode/proposals/${proposalID}/withdraw`,
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

  @then(/chaincode \(name: (.+), channel: (.+)\) based on basic (golang|javascript|typescript) should be able to get the asset \(ID: (.+)\) by querying ReadAsset func/)
  public async canQueryAssetTransferBasic(ccName: string, channelID: string, lang: string, assetID: string) {
    const response = await axios.get(`${this.getAPIEndpoint()}/api/v1/utils/queryTransaction?channelID=${channelID}&ccName=${ccName}_${ChaincodeOpsSteps.SUFFIX}&func=ReadAsset&args=["${assetID}"]`);
    expect(response.status).to.equals(200);
    expect(response.data).to.not.equals(null);
    switch (lang) {
      case 'golang':
        expect(response.data.color).to.equals('blue');
        expect(response.data.size).to.equals(5);
        expect(response.data.owner).to.equals('Tomoko');
        expect(response.data.appraisedValue).to.equals(300);
        break;
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
      default:
        expect.fail(`currently, ${lang} is not supported`);
    }
  }
}
