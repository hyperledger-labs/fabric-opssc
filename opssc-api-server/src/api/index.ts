/*
 * Copyright 2019-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import express, { Router } from 'express';
import { ChaincodeLifecycleCommands } from 'opssc-common/chaincode-lifecycle-commands';
import { ChannelCommands } from 'opssc-common/channel-commands';
import { OpsSCConfig } from 'opssc-common/config';
import { FabricClient } from 'opssc-common/fabric-client';
import { ChaincodeUpdateProposalInput, ChannelUpdateProposalInput, HistoryQueryParams, VoteTaskStatusUpdate } from 'opssc-common/opssc-types';
import { logger } from '../logger';

export default function router(fabricClient: FabricClient, opsSCConfig: OpsSCConfig): Router {

  const router = express.Router();

  // ----- REST API to get version information
  router.get('/version', async (req, res) => {
    res.json(process.env.npm_package_version);
  });

  // ----- Internal functions to create command instances

  function createChaincodeLifecycleCommands(channelID: string): ChaincodeLifecycleCommands {
    return new ChaincodeLifecycleCommands(channelID, fabricClient.getIdentity(), fabricClient.config.connectionProfile);
  }

  function createChannelCommands(): ChannelCommands {
    return new ChannelCommands(fabricClient.config.adminMSPID, fabricClient.config.adminMSPConfigPath, fabricClient.config.connectionProfile);
  }

  // ----- Utility functions to query/invoke OpsSC chaincodes

  async function queryChaincodeOpsSC(func: string, ...args: string[]):Promise<string> {
    const request = {
      channelID: opsSCConfig.channelID,
      chaincodeName: opsSCConfig.chaincodes.chaincodeOpsCCName,
      func: func,
      args: args
    };
    return await fabricClient.evaluateTransaction(request);
  }

  async function invokeChaincodeOpsSC(func: string, ...args: string[]):Promise<string> {
    const request = {
      channelID: opsSCConfig.channelID,
      chaincodeName: opsSCConfig.chaincodes.chaincodeOpsCCName,
      func: func,
      args: args
    };
    return await fabricClient.submitTransaction(request);
  }

  async function queryChannelOpsSC(func: string, ...args: string[]):Promise<string> {
    const request = {
      channelID: opsSCConfig.channelID,
      chaincodeName: opsSCConfig.chaincodes.channelOpsCCName,
      func: func,
      args: args
    };
    return await fabricClient.evaluateTransaction(request);
  }

  async function invokeChannelOpsSC(func: string, ...args: string[]):Promise<string> {
    const request = {
      channelID: opsSCConfig.channelID,
      chaincodeName: opsSCConfig.chaincodes.channelOpsCCName,
      func: func,
      args: args
    };
    return await fabricClient.submitTransaction(request);
  }

  // ----- REST API to interact the OpsSC chaincode for operating chaincodes and to query information on chaincodes

  router.get('/chaincode/getInstalledChaincodes', async (req, res) => {
    let lifecycleCommands;
    try {
      const channelID = opsSCConfig.channelID; // Workaround: Specify ops-channel as the channel name to use _lifecycle system chaincode via fabric-sdk-node
      lifecycleCommands = createChaincodeLifecycleCommands(channelID);
      const installedChaincodes = await lifecycleCommands.queryInstalledChaincodes();
      res.json(installedChaincodes);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        lifecycleCommands?.close();
      } catch (e) {
        logger.error(`failed to close lifecycle commands ${e.message}`);
      }
    }
  });


  router.get('/chaincode/queryChaincodeDefinition', async (req, res) => {
    let lifecycleCommands;
    try {
      const channelID = String(req.query.channelID);
      const chaincodeName = String(req.query.chaincodeName);

      lifecycleCommands = createChaincodeLifecycleCommands(channelID);
      const chaincodeDefinition = await lifecycleCommands.queryChaincodeDefinition(chaincodeName);

      res.json(chaincodeDefinition);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        lifecycleCommands?.close();
      } catch (e) {
        logger.error(`failed to close lifecycle commands ${e.message}`);
      }
    }
  });


  router.get('/chaincode/queryChaincodeDefinitions', async (req, res) => {
    let lifecycleCommands;
    try {
      const channelID = String(req.query.channelID);

      lifecycleCommands = createChaincodeLifecycleCommands(channelID);
      const chaincodeDefinitions = await lifecycleCommands.queryChaincodeDefinitions();

      res.json(chaincodeDefinitions);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        lifecycleCommands?.close();
      } catch (e) {
        logger.error(`failed to close lifecycle commands ${e.message}`);
      }
    }
  });

  router.get('/chaincode/proposals', async (req, res) => {
    try {
      const proposals = JSON.parse(await queryChaincodeOpsSC('GetAllProposals'));
      res.json(proposals);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.get('/chaincode/proposals/:id', async (req, res) => {
    try {
      const proposalID = req.params.id;
      const proposal = JSON.parse(await queryChaincodeOpsSC('GetProposal', proposalID));
      res.json(proposal);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.get('/chaincode/proposals/:id/histories', async (req, res) => {
    try {
      const params: HistoryQueryParams = {
        proposalID: String(req.params.id),
        taskID: req.query.taskID !== undefined ? String(req.query.taskID) : undefined
      };
      const result = JSON.parse(await queryChaincodeOpsSC('GetHistories', JSON.stringify(params)));
      res.json(result);
    } catch (e) {
      if (e.message != null) {
        logger.error(e.message);
      }
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.post('/chaincode/proposals/:id', async (req, res) => {
    try {
      const input = req.body.proposal as ChaincodeUpdateProposalInput;
      input.ID = req.params.id;
      const result = JSON.parse(await invokeChaincodeOpsSC('RequestProposal', JSON.stringify(input)));
      res.json(result);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.post('/chaincode/proposals/:id/vote', async (req, res) => {
    try {
      let taskStatusUpdate:VoteTaskStatusUpdate = {
        proposalID: req.params.id
      };
      if (req.body.updateRequest) {
        taskStatusUpdate = req.body.updateRequest as VoteTaskStatusUpdate;
        taskStatusUpdate.proposalID = req.params.id;
      }
      const result = await invokeChaincodeOpsSC('Vote', JSON.stringify(taskStatusUpdate));
      res.json(result);
    } catch (e) {
      logger.error(e.message);
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.post('/chaincode/proposals/:id/withdraw', async (req, res) => {
    try {
      const result = await invokeChaincodeOpsSC('WithdrawProposal', req.params.id);
      res.json(result);
    } catch (e) {
      logger.error(e.message);
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  // ----- REST API to interact the OpsSC chaincode for operating channels and to query information on channels

  router.get('/channel/getChannels', async (req, res) => {
    try {
      const proposals = JSON.parse(await queryChannelOpsSC('GetAllChannels'));
      res.json(proposals);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.post('/channel/proposals/:id', async (req, res) => {
    let channelCommands;
    try {
      const proposalID = req.params.id;

      const proposal = req.body.proposal;
      const channelID = proposal.channelID;
      const profile = proposal.opsProfile;
      const description = proposal.description;
      let action = proposal.action;

      let deltaBase64 = '';
      channelCommands = createChannelCommands();

      // Create ConfigUpdate (delta) encoded by base64 for the proposal
      switch (action) {
        case 'create': {
          // Create ConfigUpdate to create the specified channel
          deltaBase64 = channelCommands.createDeltaToCreateChannel(channelID, profile, 'base64');
          break;
        }
        default: {
          action = 'update';

          // Fetch channel config block
          const configBlockFilePath = channelCommands.fetch(channelID, 'config', 'filePath') as string;

          // Create ConfigUpdate to do multiple operations to the specified channel
          deltaBase64 = channelCommands.createDeltaToUpdateChannelByMultipleOperations(channelID, configBlockFilePath, profile, 'base64');
          break;
        }
      }
      // Create ConfigSignature as base64 for the proposal
      const signBase64 = channelCommands.sign(deltaBase64, 'base64');

      const input: ChannelUpdateProposalInput = {
        ID: proposalID,
        channelID: channelID,
        description: description,
        action: action,
        opsProfile: proposal.opsProfile,
        configUpdate: deltaBase64,
        signature: signBase64,
      };

      // Send transaction for proposal
      const result = await invokeChannelOpsSC('RequestProposal', JSON.stringify(input));
      res.json(result);
    } catch (e) {
      logger.error(e.message);
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        channelCommands?.cleanUp();
      } catch (e) {
        logger.error(`fail to clean up channelCommands: ${e}`);
      }
    }
  });

  router.post('/channel/proposals/:id/vote', async (req, res) => {
    let channelCommands;
    try {
      const proposalID = req.params.id;

      // Get the proposal from channel_ops
      const proposal = JSON.parse(await queryChannelOpsSC('GetProposal', proposalID));

      // Create ConfigSignature as base64 from the proposal
      channelCommands = createChannelCommands();
      const signBase64 = channelCommands.sign(proposal.artifacts.configUpdate, 'base64');

      // Send transaction for vote
      const result = await invokeChannelOpsSC('Vote', proposalID, signBase64);
      res.json(result);
    } catch (e) {
      logger.error(e.message);
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        channelCommands?.cleanUp();
      } catch (e) {
        logger.error(`fail to clean up channelCommands: ${e}`);
      }
    }
  });

  router.get('/channel/proposals/:id', async (req, res) => {
    try {
      const proposalID = req.params.id;
      const proposal = JSON.parse(await queryChannelOpsSC('GetProposal', proposalID));

      res.json(proposal);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.get('/channel/proposals', async (req, res) => {
    try {
      const proposals = JSON.parse(await queryChannelOpsSC('GetAllProposals'));

      res.json(proposals);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.get('/channel/systemConfigBlock', async (req, res) => {
    let channelCommands;
    try {
      const systemChannelID = String(await queryChannelOpsSC('GetSystemChannelID'));

      // fetch channel config block (ConfigSignature) and encode ConfigSignature to base64
      channelCommands = createChannelCommands();
      const configBlockBase64 = (channelCommands.fetch(systemChannelID, 'config', 'buffer') as Buffer).toString('base64');
      channelCommands.cleanUp();

      res.json(configBlockBase64);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    } finally {
      try {
        channelCommands?.cleanUp();
      } catch (e) {
        logger.error(`fail to clean up channelCommands: ${e}`);
      }
    }
  });

  // ----- REST API to query and invoke chaincodes (for test)

  router.get('/utils/queryTransaction', async (req, res) => {
    try {
      const ccName = String(req.query.ccName);
      const func = String(req.query.func);
      const args = JSON.parse(String(req.query.args));
      const channelID = String(req.query.channelID);
      const request = {
        channelID: channelID,
        chaincodeName: ccName,
        func: func,
        args: args
      };
      const result = JSON.parse(await fabricClient.evaluateTransaction(request));
      res.json(result);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  router.post('/utils/invokeTransaction', async (req, res) => {
    try {
      const request = {
        channelID: req.body.channelID,
        chaincodeName: req.body.ccName,
        func: req.body.func,
        args: req.body.args
      };
      const result = await fabricClient.submitTransaction(request);
      res.json(result);
    } catch (e) {
      res.status(500).json({
        message: e.toString()
      });
    }
  });

  return router;
}
