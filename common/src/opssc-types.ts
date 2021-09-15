/*
 * Copyright 2019-2021 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// Types for chaincode ops

export type ChaincodePackage = {
  repository: string;
  commitID: string;
  pathToSourceFiles?: string;
  type: string;
}

export type ChaincodeDefinition = {
  sequence: number;
  initRequired: boolean;
  validationParameter: string;
}

export type ChaincodeUpdateProposal = {
  ID: string;
  channelID: string;
  chaincodeName: string;
  creator: string;
  chaincodePackage: ChaincodePackage;
  chaincodeDefinition: ChaincodeDefinition;
  status: string;
  time: string;
}

export type ChaincodeUpdateProposalInput = {
  ID: string;
  channelID: string;
  chaincodeName: string;
  chaincodePackage: ChaincodePackage;
  chaincodeDefinition: ChaincodeDefinition;
}

export type ChaincodeDeploymentEventDetail = {
  proposal: ChaincodeUpdateProposal;
  operationTargets: string[];
}

// Types for channel ops

export interface ChannelUpdateProposal {
  ID: string;
  channelID: string;
  description?: string;
  creator: string;
  action: string;
  opsProfile: any;
  artifacts: Artifacts;
}

export interface Artifacts {
  configUpdate: string;
  signatures?: {[key: string]: string};
}

export interface ChannelUpdateProposalInput {
  ID: string;
  channelID: string;
  description?: string;
  action?: string;
  opsProfile: any;
  configUpdate: string;
  signature: string;
}

export type Channel = {
  ID: string;
  channelType: string;
  organizations: {
    [key: string]: string;
  } | null;
}

export type ChannelOpsEventDetail = {
  proposalID: string;
  operationTargets: string[];
}


// Common types

export interface History {
  proposalID: string;
  orgID?: string;
  taskID?: string;
  status: TaskStatus;
  data?: string;
  time?: string;
}

export interface TaskStatusUpdate {
  proposalID: string;
  status?: TaskStatus;
  data?: string;
}

export interface VoteTaskStatusUpdate extends TaskStatusUpdate{
  status?: VoteTaskStatus;
}

export type TaskStatus = VoteTaskStatus | AgentTaskStatus
export type VoteTaskStatus = 'agreed' | 'disagreed'
export type AgentTaskStatus = 'success' | 'failure'

export interface HistoryQueryParams {
	proposalID: string;
	taskID?:     string;
	orgID?:      string;
}