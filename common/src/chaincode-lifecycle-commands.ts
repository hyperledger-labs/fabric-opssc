/*
 * Copyright 2020-2021 Hitachi America, Ltd. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { protos as fabric_common_protos, lifecycle as lifecycle_protos } from 'fabric-protos';
import { Contract, DefaultQueryHandlerStrategies, Gateway, GatewayOptions, Identity } from 'fabric-network';
// eslint-disable-next-line @typescript-eslint/no-var-requires
const { buildPolicy } = require('./lib/Policy');
import { finalPackage as finalPackageChaincode, package as packageChaincode } from './lib/Packager';
import { Channel, Endorser } from 'fabric-common';
import crypto from 'crypto';
import { logger } from './logger';

export interface ChaincodeDefinition {
  name: string;
  sequence: number | Long;
  version: string;
  endorsement_plugin?: string;
  validation_plugin?: string;
  validation_parameter?: any;
  package_id?: string;
  collections?: any;
  init_required?: boolean;
}

export interface QueryChaincodeDefinitionResult {
  name?: string;
  sequence: number | Long;
  version: string;
  endorsement_plugin?: string;
  validation_plugin?: string;
  validation_parameter?: any;
  package_id?: string;
  collections?: any;
  init_required?: boolean;
  approvals?: { [key: string]: boolean };
}

export interface QueryChaincodeDefinitionsResult {
  chaincode_definitions?: QueryChaincodeDefinitionResult[];
}

export interface QueryInstalledChaincodeResult {
  package_id?: string;
  label?: string;
  references?: {
    [k: string]: {
      chaincodes?: {
        name?: string;
        version?: string;
      }
    }
  };
}

export interface QueryInstalledChaincodesResult {
  installed_chaincodes?: QueryInstalledChaincodeResult[];
}

export interface PackageRequest {
  lang: string;
  chaincodePath: string;
  label: string;
  metadataPath?: string;
  goPath?: string;
}

export interface InstallRequest {
  package: Buffer
}

export interface ChaincodeRequest {
  chaincode: ChaincodeDefinition;
}


const LIFECYCLE_SCC_NAME = '_lifecycle';
const DEFAULT_QUERY_TIMEOUT_MINUTES = 300;

/**
 * <p> ChaincodeLifecycleCommands is a class to perform _lifecycle operations.
 * This class internally calls _lifecycle system chaincode instead of using "peer" binary.
 * </p>
 *
 */
/*
 * NOTE: Currently, this class uses Gateway in fabric-network instead of low-level APIs in fabric-common directly to simplify the implementation.
 * In the future, it may need to be implemented based on fabric-common for finer control.
 */
export class ChaincodeLifecycleCommands {

  private readonly connectionProfile: any;
  private readonly identity: Identity;
  private readonly discoverAsLocalhost: boolean;
  private channelID: string;

  private gateway: Gateway | null;
  private lifecycleSCC: Contract | null;
  private channel: Channel | null;

  /**
   * ChaincodeLifecycleCommands constructor
   *
   * @param {string} channelID the channel ID on which _lifecycle operations should be executed
   * @param {Identity} identity the client identity to call _lifecycle operations
   * @param {any} connectionProfile the connection profile that provides the necessary connection information for the client organization
   * @param {boolean} discoverAsLocalhost whether to discover the target nodes as localhost
   */
  constructor(channelID: string, identity: Identity, connectionProfile: any, discoverAsLocalhost?: boolean) {
    this.connectionProfile = connectionProfile;
    this.identity = identity;
    this.channelID = channelID;
    this.discoverAsLocalhost = (discoverAsLocalhost === true);

    this.gateway = null;
    this.lifecycleSCC = null;
    this.channel = null;
  }

  /**
   * Package a chaincode.
   *
   * @async
   * @param {PackageRequest} request the request to package a chaincode
   * @returns {Promise<Buffer>} the package binary
   */
  async package(request: PackageRequest): Promise<Buffer> {
    const inner_tarball = await packageChaincode(request.chaincodePath, request.lang, request.metadataPath, request.goPath);
    return await finalPackageChaincode(request.label, request.lang, inner_tarball, request.chaincodePath);
  }

  /**
   * <p> Install a chaincode to all peers for the target client organization.
   * This ensures that all peers has the chaincode installed, even if the chaincode are already installed on some of peers.</p>
   *
   * @async
   * @param {InstallRequest} request the request to install a chaincode
   * @returns {Promise<string|null>} the installed package ID (null if the chaincode are already installed on all peers)
   */
  async install(request: InstallRequest): Promise<string|null> {
    await this.prepareLifecycleContract();

    let preferredPackageID: string | null = null;

    // Install to multiple peers for an org in parallel (Make the chaincode to be installed on all peers)
    const args = [lifecycle_protos.InstallChaincodeArgs.encode({ chaincode_install_package: request.package }).finish()];
    const peers = this.channel!.getEndorsers(this.identity.mspId);
    await Promise.all(peers.map(async (peer) => {
      try {
        const result = await this.queryChaincode(LIFECYCLE_SCC_NAME, 'InstallChaincode', args, peer);
        const { package_id, label } = lifecycle_protos.InstallChaincodeResult.decode(result);
        logger.debug(`Install chaincode result: ${package_id}, ${label}, ${peer.name}`);
        if (!preferredPackageID) {
          preferredPackageID = package_id;
        }
        else if (preferredPackageID != package_id) {
          throw new Error(`Package IDs are inconsistent (preferred: ${preferredPackageID}, current peer: ${package_id})`);
        }
      } catch (error) {
        if (error.message == null || !(error.message as string).includes('chaincode already successfully installed')) {
          throw error;
        }
        logger.warn(`Chaincode already successfully installed for ${peer.name}`);
      }
    }));
    return preferredPackageID;
  }

  /**
   * Approve a chaincode definition for the target client organization.
   *
   * @async
   * @param {ChaincodeRequest} request the request to approve a chaincode definition
   * @returns {Promise<void>}
   */
  async approve(request: ChaincodeRequest): Promise<void> {
    await this.prepareLifecycleContract();

    const source = new lifecycle_protos.ChaincodeSource();
    if (request.chaincode.package_id) {
      const localPackage = new lifecycle_protos.ChaincodeSource.Local();
      localPackage.package_id = request.chaincode.package_id;
      source.local_package = localPackage;
      source.Type = 'local_package';
    } else {
      source.unavailable = new lifecycle_protos.ChaincodeSource.Unavailable();
      source.Type = 'unavailable';
    }

    const approveChaincodeDefinitionForMyOrgArgs: lifecycle_protos.IApproveChaincodeDefinitionForMyOrgArgs = {
      sequence: request.chaincode.sequence,
      name: request.chaincode.name,
      version: request.chaincode.version,
      source: source
    };
    if (request.chaincode.validation_parameter) {
      approveChaincodeDefinitionForMyOrgArgs.validation_parameter = createEndorsementPolicyDefinition(request.chaincode.validation_parameter);
    }
    if (request.chaincode.endorsement_plugin) {
      approveChaincodeDefinitionForMyOrgArgs.endorsement_plugin = request.chaincode.endorsement_plugin;
    }
    if (request.chaincode.validation_plugin) {
      approveChaincodeDefinitionForMyOrgArgs.validation_plugin = request.chaincode.validation_plugin;
    }
    if (request.chaincode.init_required) {
      approveChaincodeDefinitionForMyOrgArgs.init_required = request.chaincode.init_required;
    }
    if (request.chaincode.collections) {
      approveChaincodeDefinitionForMyOrgArgs.collections = request.chaincode.collections;
    }

    const args = [lifecycle_protos.ApproveChaincodeDefinitionForMyOrgArgs.encode(approveChaincodeDefinitionForMyOrgArgs).finish()];

    const transaction = this.lifecycleSCC!.createTransaction('ApproveChaincodeDefinitionForMyOrg');
    transaction.setEndorsingPeers(this.channel!.getEndorsers(this.identity.mspId));

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore: To set bytes in args instead of a string array
    await transaction.submit(...args);
  }

  /**
   * Commit chaincode definition on the target channel.
   *
   * @async
   * @param {ChaincodeRequest} request the request to commit a chaincode definition
   * @returns {Promise<void>}
   */
  async commit(request: ChaincodeRequest): Promise<void> {
    await this.prepareLifecycleContract();

    const commitChaincodeDefinitionArgs: lifecycle_protos.ICommitChaincodeDefinitionArgs = {
      sequence: request.chaincode.sequence,
      name: request.chaincode.name,
      version: request.chaincode.version
    };
    if (request.chaincode.validation_parameter) {
      commitChaincodeDefinitionArgs.validation_parameter = createEndorsementPolicyDefinition(request.chaincode.validation_parameter);
    }
    if (request.chaincode.endorsement_plugin) {
      commitChaincodeDefinitionArgs.endorsement_plugin = request.chaincode.endorsement_plugin;
    }
    if (request.chaincode.validation_plugin) {
      commitChaincodeDefinitionArgs.validation_plugin = request.chaincode.validation_plugin;
    }
    if (request.chaincode.init_required) {
      commitChaincodeDefinitionArgs.init_required = request.chaincode.init_required;
    }
    if (request.chaincode.collections) {
      commitChaincodeDefinitionArgs.collections = request.chaincode.collections;
    }

    const args = [lifecycle_protos.CommitChaincodeDefinitionArgs.encode(commitChaincodeDefinitionArgs).finish()];

    const transaction = this.lifecycleSCC!.createTransaction('CommitChaincodeDefinition');

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore: To set bytes in args instead of a string array
    await transaction.submit(...args);
  }

  /**
   * <p> Query the installed chaincodes on a peer of the target client organization.
   * The target peer is randomly selected internally. </p>
   *
   * @async
   * @returns {Promise<QueryInstalledChaincodesResult>} the installed chaincodes
   */
  async queryInstalledChaincodes(): Promise<QueryInstalledChaincodesResult> {
    await this.prepareLifecycleContract();

    const transaction = this.lifecycleSCC!.createTransaction('QueryInstalledChaincodes');

    const args = [lifecycle_protos.QueryInstalledChaincodesArgs.encode({}).finish()];

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore: To set bytes in args instead of a string array
    const result = await transaction.evaluate(...args);
    return lifecycle_protos.QueryInstalledChaincodesResult.decode(result) as QueryInstalledChaincodesResult;
  }

  /**
   * Query the committed chaincode with the given chaincode name on the target channel.
   *
   * @async
   * @param {string} chaincodeName the target chaincode name
   * @returns {Promise<QueryChaincodeDefinitionResult>} the committed chaincode
   */
  async queryChaincodeDefinition(chaincodeName: string): Promise<QueryChaincodeDefinitionResult> {
    await this.prepareLifecycleContract();

    const transaction = this.lifecycleSCC!.createTransaction('QueryChaincodeDefinition');

    const args = [lifecycle_protos.QueryChaincodeDefinitionArgs.encode({ name: chaincodeName }).finish()];

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore: To set bytes in args instead of a string array
    const result = await transaction.evaluate(...args);
    return lifecycle_protos.QueryChaincodeDefinitionResult.decode(result) as QueryChaincodeDefinitionResult;
  }

  /**
   * Query all the committed chaincodes on the target channel.
   *
   * @async
   * @returns {Promise<QueryChaincodeDefinitionsResult>} the committed chaincodes
   */
  async queryChaincodeDefinitions(): Promise<QueryChaincodeDefinitionsResult> {
    await this.prepareLifecycleContract();

    const transaction = this.lifecycleSCC!.createTransaction('QueryChaincodeDefinitions');

    const args = [lifecycle_protos.QueryChaincodeDefinitionsArgs.encode({}).finish()];

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore: To set bytes in args instead of a string array
    const result = await transaction.evaluate(...args);
    return lifecycle_protos.QueryChaincodeDefinitionsResult.decode(result) as QueryChaincodeDefinitionsResult;
  }

  /*
   * Internal method to get a new Gateway to access _lifecycle chaincodes.
   * This gateway should use peers for the connected organization.
   */
  private async getFabricGateway(): Promise<Gateway> {

    const connectOpt: GatewayOptions = {
      identity: this.identity,
      discovery: {
        enabled: true,
        asLocalhost: this.discoverAsLocalhost
      },
      queryHandlerOptions: {
        timeout: DEFAULT_QUERY_TIMEOUT_MINUTES,
        strategy: DefaultQueryHandlerStrategies.MSPID_SCOPE_SINGLE
      }
    };

    const gateway = new Gateway();
    await gateway.connect(this.connectionProfile, connectOpt);
    return gateway;
  }

  /*
   * Internal method to query the given chaincode with the given peer.
   * NOTE: This method internally uses low-level APIs in fabric-common because high-level APIs in fabric-network does not seem to provide APIs
   * (1) with specifying query peers or (2) for querying all peers with fine controls.
   *
   * This method is used to install a chaincode to a specific peer.
   */
  private async queryChaincode(chaincodeName: string, functionName: string, args: Uint8Array[] | string[] | undefined, peer: Endorser): Promise<Buffer> {
    const query = this.channel!.newQuery(chaincodeName);
    const identityContext = this.gateway!.identityContext!;
    query.build(identityContext, {
      fcn: functionName,
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore: To set bytes in args instead of a string array
      args: args
    });
    query.sign(identityContext);
    const response = await query.send({
      targets: [peer],
      requestTimeout: DEFAULT_QUERY_TIMEOUT_MINUTES * 1000, // timeout in milliseconds
    });
    if (response.queryResults.length < 1) {
      throw new Error('Peer returned error: ' + response.responses[0].response.message);
    }

    return response.queryResults[0];
  }

  /*
   * Internal method to prepare _lifecycle contract.
   */
  private async prepareLifecycleContract(): Promise<void> {
    if (this.gateway !== null) return;

    this.gateway = await this.getFabricGateway();
    const network = await this.gateway.getNetwork(this.channelID);
    this.lifecycleSCC = network.getContract(LIFECYCLE_SCC_NAME);
    this.channel = network.getChannel();

    // Workaround:
    // Until one or more user chaincodes are deployed (approved) for an organization,
    // the current discovery service does not seem detect to _lifecycle chaincode for the organization
    // and to provide the endorsement plans including the organization.
    // So, QueryHandler.evaluate() fails with an error the "The peer is not running chaincode _lifecycle"
    // As the workaround, explicitly add _lifecycle chaincode to the endorsers of the organization.
    // Alternative: Use fabric-common directly, instead of fabric-network (gateway)
    const endorsers = this.channel!.getEndorsers(this.identity.mspId);
    for (const endorser of endorsers) {
      if (!endorser.hasChaincode(LIFECYCLE_SCC_NAME)) {
        endorser.addChaincode(LIFECYCLE_SCC_NAME);
      }
    }
  }

  /**
   * Close the Fabric network connection to execute _lifecycle chaincodes.
   */
  close() {
    if (this.gateway !== null) {
      this.gateway.disconnect();
      this.gateway = null;
    }
  }
}

/*
 * createEndorsementPolicyDefinition is an internal method to create ProtoBuffer encoded endorsement policy.
 * This is based on the implementation of fabric-client/lib/Chaincode.js in v2.0.0-beta.2
 * in https://github.com/hyperledger/fabric-sdk-node.
 */
function createEndorsementPolicyDefinition(policy: any): Uint8Array {

  // Case: Return as is if already encoded
  if (policy instanceof Uint8Array) {
    return policy as Uint8Array;
  }

  // Case: Build and encode the policy if string or json object
  const application_policy = new fabric_common_protos.ApplicationPolicy();

  if (typeof policy === 'string') {
    application_policy.channel_config_policy_reference = policy;
  } else if (policy instanceof Object) {
    const signature_policy = buildPolicy(null, policy, true);
    application_policy.signature_policy = signature_policy;
  } else {
    throw new Error('The endorsement policy is not valid');
  }

  return fabric_common_protos.ApplicationPolicy.encode(application_policy).finish();
}

/**
 * computePackageID is an utility function to compute the package ID without installing the chaincode.
 *
 * @param {string} label the chaincode label
 * @param {Buffer} packageBytes the package bytes of the chaincode
 * @returns {string} the package ID
 */
export function computePackageID(label: string, packageBytes: Buffer): string {
  const sha256 = crypto.createHash('sha256');
  const packageHash = sha256.update(packageBytes).digest().toString('hex');

  return `${label}:${packageHash}`;
}