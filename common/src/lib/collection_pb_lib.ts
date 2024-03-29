/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
*/
// This code is based on https://github.com/hyperledger-labs/fabric-operations-console/blob/main/packages/stitch/src/libs/proto_handlers/collection_pb_lib.ts

import { protos } from 'fabric-protos';
import { logger } from '../logger';
import { PolicyLib } from './policy_pb_lib';
import { conformPolicySyntax } from './sig_policy_syntax_lib';
import { pp } from './misc';

export { Scc, MixedPolicySyntax };
export class CollectionLib {
  policyLib = new PolicyLib;

  // --------------------------------------------------------------------------------
  // build a CollectionConfigPackage protobuf - returns message
  // --------------------------------------------------------------------------------
  __build_collection_config_package(configurations: Scc[]) {
    const configs = [];
    for (const i in configurations) {
      const static_collection_config = this.__build_static_collection_config(configurations[i]);
      if (static_collection_config) {
        logger.debug('[protobuf-handler] static_collection_config:', pp(static_collection_config));
        const collection_config = this.__build_collection_config(static_collection_config);
        configs.push(collection_config);
      }
    }

    if (configs.length === 0) {
      logger.error('[protobuf-handler] 0 collection configs were built, cannot build collection package');
      return null;
    } else {
      const message = protos.CollectionConfigPackage.create({ config: configs });
      logger.debug('[protobuf-handler] collection config package:', protos.CollectionConfigPackage.toObject(message, { defaults: true }));
      return message;
    }
  }

  // --------------------------------------------------------------------------------
  // build a CollectionConfigPackage protobuf - return bin
  // --------------------------------------------------------------------------------
  __b_build_collection_config_package(configurations: Scc[]) {
    const message = this.__build_collection_config_package(configurations);
    if (!message) {
      return null;
    } else {
      return <Uint8Array>protos.CollectionConfigPackage.encode(message).finish();
    }
  }

  // --------------------------------------------------------------------------------
  // build a CollectionConfig protobuf - returns obj -
  // --------------------------------------------------------------------------------
  __build_collection_config(static_collection_config: any) {
    const message = protos.CollectionConfig.create({ static_collection_config: static_collection_config });
    return protos.CollectionConfig.toObject(message, { defaults: true });
  }

  // --------------------------------------------------------------------------------
  // build a CollectionPolicyConfig protobuf - returns obj
  // --------------------------------------------------------------------------------
  __build_collection_policy_config(signature_policy: any) {
    const message = protos.CollectionPolicyConfig.create({ signature_policy: signature_policy });
    return protos.CollectionPolicyConfig.toObject(message, { defaults: true });
  }

  // --------------------------------------------------------------------------------
  // build a StaticCollectionConfig protobuf - returns obj
  // --------------------------------------------------------------------------------
  __build_static_collection_config(config: Scc) {
    if (config.policy) {
      config.member_orgs_policy = JSON.parse(JSON.stringify(config.policy));          // copy sdk's field to fabric's
      delete config.policy;
    }
    if (config.maxPeerCount || config.max_peer_count) {
      config.maximum_peer_count = config.maxPeerCount || config.max_peer_count;        // copy sdk's field to fabric's
      delete config.maxPeerCount;
      delete config.max_peer_count;
    }

    // validation
    if (!config.member_orgs_policy) {
      logger.error('[protobuf-handler] collection config policy is missing "member_orgs_policy" aka "policy" field', pp(config));
      return null;
    } else {
      const private_data_fmt = conformPolicySyntax(config.member_orgs_policy);    // accepts fabric's format, sdk's format, or peer cli format
      const endorsement_fmt = (config.endorsement_policy && config.endorsement_policy.signature_policy) ?
        conformPolicySyntax(config.endorsement_policy.signature_policy) : null;

      const private_data_spe = private_data_fmt ? this.policyLib.__build_signature_policy_envelope_alt(private_data_fmt) : null;
      const endorsement_spe = endorsement_fmt ? this.policyLib.__build_signature_policy_envelope_alt(endorsement_fmt) : null;

      const opts = {
        name: config.name,
        member_orgs_policy: this.__build_collection_policy_config(private_data_spe),  // build message for field
        required_peer_count: config.required_peer_count,
        maximum_peer_count: config.maximum_peer_count,
        block_to_live: config.block_to_live,
        member_only_read: config.member_only_read,
        member_only_write: config.member_only_write,
        endorsement_policy: this.policyLib.__build_application_policy(        // build message for field
          endorsement_spe,
          config.endorsement_policy ? config.endorsement_policy.channel_config_policy_reference : null
        ),
      };
      const message = protos.StaticCollectionConfig.create(opts);
      return protos.StaticCollectionConfig.toObject(message, { defaults: true });
    }
  }
}

interface MixedPolicySyntax {
  version: number;        // [optional] - fabric syntax uses this field
  rule: any;            // fabric syntax uses this field
  identities: any;        // fabric & fabric-sdk syntax uses this field
  policy: any;          // [legacy] "fabric-sdk syntax" format uses this field
}

interface Scc {
  name: string;            // name of your collection
  required_peer_count: number;     // min number of peers that must get the private data to successfully endorse

  member_orgs_policy: MixedPolicySyntax | string | null; // sig policy of which orgs have access to the private data - supports peer-cli syntax using string
  policy: any | null;          // [legacy] fabric-sdk syntax uses this field... convert to "member_orgs_policy"

  maximum_peer_count: number | null | undefined;  // max number of peers the endorsing peer can try to send private data to
  maxPeerCount: number | null | undefined;     // [legacy] fabric sdk" format uses this... convert to "maximum_peer_count"
  max_peer_count: number | null | undefined;    // [legacy 2] fabric sdk" format uses this... convert to "maximum_peer_count" - this version shouldn't exist

  block_to_live: number;        // when to expire private data, after this number of blocks the private data is deleted, 0 = unlimited
  member_only_read: boolean;      // if true, only collection member clients can read private data
  member_only_write: boolean;     // if true, only collection member clients can write private data


  // proto file: /v2.0/peer/policy.proto
  endorsement_policy: {

    // [optional, set 1] "signature_policy" = { "version" : <number>, "rule": <SignaturePolicy>, "identities": <MSPPrincipal[]> }
    signature_policy: MixedPolicySyntax | string | null;  // peer-cli syntax uses string

    // [optional, set 1]  should identify a policy defined in the channel's configuration block
    channel_config_policy_reference: string | null;
  };
}