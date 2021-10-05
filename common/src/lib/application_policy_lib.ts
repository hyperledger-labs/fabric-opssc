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
// This code is based on https://github.com/hyperledger-labs/fabric-operations-console/blob/main/packages/stitch/src/libs/lifecycle_pb_lib.ts

import { logger } from '../logger';
import { common, protos } from 'fabric-protos';
import { conformPolicySyntax } from  './sig_policy_syntax_lib';

/*
 * createEndorsementPolicyDefinition is an internal method to create ProtoBuffer encoded endorsement policy.
 * This is based on the implementation of __b_build_application_policy().
 * in https://github.com/hyperledger-labs/fabric-operations-console/blob/main/packages/stitch/src/libs/lifecycle_pb_lib.ts.
 */
function createEndorsementPolicyDefinition(policy: any): Uint8Array | null {
  const app_policy_opts = {
    channel_config_policy_reference: <string | undefined>'/Channel/Application/Endorsement',			// default endorsement policy name
    signature_policy: <any>{}
  };
  if (!policy) {												// use default
    logger.debug('protobuf-handler] ccd - using the default channel policy reference for the endorsement policy');
    delete app_policy_opts.signature_policy;										// delete the other option, only 1 should be set
  } else if (policy[0] === '/') {
    logger.debug('protobuf-handler] ccd - using the provided channel policy reference for the endorsement policy');
    app_policy_opts.channel_config_policy_reference = policy;	// set provided channel policy name
    delete app_policy_opts.signature_policy;
  } else {
    logger.debug('protobuf-handler] ccd - parsing signature policy syntax for the endorsement policy');
    const endorsement_fmt = conformPolicySyntax(policy);		// conform it to fabric's structure
    if (!endorsement_fmt) {
      return null;
    } else {
      app_policy_opts.signature_policy = __build_MSPPrincipal_bin_inside(endorsement_fmt);
      delete app_policy_opts.channel_config_policy_reference;
    }
  }
  const message = protos.ApplicationPolicy.create(app_policy_opts);
  const b_applicationPolicy = <Uint8Array>protos.ApplicationPolicy.encode(message).finish();
  return b_applicationPolicy;
}

function __build_MSPPrincipal_bin_inside(obj: any) {
  if (obj) {
    for (const i in obj.identities) {
      if (obj.identities[i].principal) {
        const message = common.MSPRole.create(obj.identities[i].principal);
        const b_MSPRole = <Uint8Array>common.MSPRole.encode(message).finish();
        obj.identities[i].principal = b_MSPRole;							// overwrite with binary representation
      }
    }
  }
  return obj;
}

// exports
export { createEndorsementPolicyDefinition };
