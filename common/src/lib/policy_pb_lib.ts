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
// This code is based on https://github.com/hyperledger-labs/fabric-operations-console/blob/main/packages/stitch/src/libs/proto_handlers/policy_pb_lib.ts

import { common } from 'fabric-protos';
import { logger } from '../logger';
import { pp } from './misc';

import Policy_SignaturePolicyEnvelope = common.SignaturePolicyEnvelope;
import Policies_ImplicitMetaPolicy = common.ImplicitMetaPolicy;
import MSP_Principle_MSPPrincipal = common.MSPPrincipal;
import Policy_SignaturePolicy = common.SignaturePolicy;
import MSP_Principle_MSPRole = common.MSPRole;

export class PolicyLib {

  // --------------------------------------------------------------------------------
  // make a signature policy envelope protobuf
  // --------------------------------------------------------------------------------
  /*
		opts:{
			version: <number>
			p_rule: <protobuf SignaturePolicy>
			p_identities: [<protobuf MSPPrincipal>]
		}
	*/
  p_build_signature_policy_envelope(opts: Bse) {
    const p_signaturePolicyEnvelope = new Policy_SignaturePolicyEnvelope();
    p_signaturePolicyEnvelope.version = opts.version;
    p_signaturePolicyEnvelope.rule = opts.p_rule;
    p_signaturePolicyEnvelope.identities = opts.p_identities;
    return p_signaturePolicyEnvelope;
  }

  // --------------------------------------------------------------------------------
  // make a signature policy protobuf
  // --------------------------------------------------------------------------------
  /*
		opts: {
			n_out_of: <protobuf>,		// only one of these should be set! else use `null`
			signed_by: <number>			// only one of these should be set! else use `null`
		}
	*/
  p_build_signature_policy(opts: Bsp) {
    const p_signaturePolicy = new Policy_SignaturePolicy();
    if (opts.n_out_of !== null) {
      p_signaturePolicy.n_out_of = opts.n_out_of;
    } else if (opts.signed_by !== null) {					// remember this is a number, so 0 is possible/valid
      p_signaturePolicy.signed_by = opts.signed_by;
    }

    logger.debug('[protobuf-handler] policy - has n_out_of:', !p_signaturePolicy, 'has signed_by:', !p_signaturePolicy, opts.signed_by);
    return p_signaturePolicy;
  }

  // --------------------------------------------------------------------------------
  // make a n out of any protobuf
  // --------------------------------------------------------------------------------
  /*
		opts: {
			n: <number>;							// number of signatures
			rules_list: Policy_SignaturePolicy[];	// array of rule lists...
		}
	*/
  p_build_n_out_of(opts: B1o) {
    const p_n_out_of = new Policy_SignaturePolicy.NOutOf();		// this is how to create a nested message, nested protoc
    if (isNaN(opts.n)) {
      logger.error('cannot build "n-of" policy b/c "n" is not a number', opts.n);
    } else {
      p_n_out_of.n = opts.n;
      p_n_out_of.rules = opts.rules_list;
    }
    return p_n_out_of;
  }

  // --------------------------------------------------------------------------------
  // build a default endorsement policy - any 1 member may sign (msp ids must be provided)
  // --------------------------------------------------------------------------------
  /*
		opts: {
			msp_ids: [ "PeerOrg1"]
		}
	*/
  p_build_default_e_policy_envelope(opts: Bdp) {
    const signed_by_list = [], principal_list = [];
    for (const i in opts.msp_ids) {

      const p_MSPRole = new MSP_Principle_MSPRole();
      p_MSPRole.role = MSP_Principle_MSPRole.MSPRoleType.MEMBER;
      p_MSPRole.msp_identifier = opts.msp_ids[i];

      const p_MSPPrincipal = new MSP_Principle_MSPPrincipal();
      p_MSPPrincipal.principal_classification = MSP_Principle_MSPPrincipal.Classification.ROLE;
      p_MSPPrincipal.principal = MSP_Principle_MSPRole.encode(p_MSPRole).finish();

      const p_signaturePolicy = this.p_build_signature_policy({ signed_by: Number(i), n_out_of: null });
      signed_by_list.push(p_signaturePolicy);		// its import that these two arrays get built together
      principal_list.push(p_MSPPrincipal);		// position x of each array should refer to the same msp
    }

    const p_oneOfAny = this.p_build_n_out_of({ n: 1, rules_list: signed_by_list });
    const p_signaturePolicy2 = this.p_build_signature_policy({ signed_by: null, n_out_of: p_oneOfAny });
    const bsp_opts = {
      version: 0,
      p_rule: p_signaturePolicy2,
      p_identities: principal_list
    };
    const p_signaturePolicyEnvelope = this.p_build_signature_policy_envelope(bsp_opts);
    return p_signaturePolicyEnvelope;
  }

  // --------------------------------------------------------------------------------
  // build a "MSPRole" protobuf - returns message
  // --------------------------------------------------------------------------------
  p_build_msp_role(opts: Bmr) {
    const role_name = (opts && opts.role) ? opts.role.toUpperCase() : null;
    if (!role_name || roles_map[role_name] === undefined) {
      logger.error('[protobuf-handler] cannot find or invalid "role". cannot build MSPRole for signature policy. role:', role_name);
      return null;
    } else if (!opts.msp_identifier) {
      logger.error('[protobuf-handler] undefined msp id. cannot build MSPRole for signature policy');
      return null;
    } else {
      const p_MSPRole = new MSP_Principle_MSPRole();
      p_MSPRole.role = roles_map[role_name];
      p_MSPRole.msp_identifier = opts.msp_identifier;
      return p_MSPRole;
    }
  }

  // --------------------------------------------------------------------------------
  // build a MSPRole msg - returns binary
  // --------------------------------------------------------------------------------
  __b_build_msp_role(opts: { msp_identifier: string, role: string }) {
    const role_name = (opts && opts.role) ? opts.role.toUpperCase() : null;

    if (!role_name || roles_map[role_name] === undefined) {
      logger.error('[protobuf-handler] cannot find or invalid "role". cannot build MSPRole for signature policy. role:', role_name);
      return null;
    } else if (!opts.msp_identifier) {
      logger.error('[protobuf-handler] undefined msp id. cannot build MSPRole for signature policy');
      return null;
    }

    // use fromObject instead of create b/c role_name will be a string, not enum
    const message = MSP_Principle_MSPRole.create({ role: roles_map[role_name], msp_identifier: opts.msp_identifier });
    const b_MSPRole = <Uint8Array>MSP_Principle_MSPRole.encode(message).finish();
    return b_MSPRole;
  }

  // --------------------------------------------------------------------------------
  // build a MSPPrincipal msg - returns message
  // --------------------------------------------------------------------------------
  __build_msp_principal(opts: { principal_classification: number, b_principal: Uint8Array }) {
    const p_opts = {
      principal_classification: opts.principal_classification,
      principal: opts.b_principal
    };
    const message = MSP_Principle_MSPPrincipal.create(p_opts);
    return message;
  }

  // --------------------------------------------------------------------------------
  // build a custom cc policy - using Fabric SDK's format - :( (alt format in __build_signature_policy_envelope)
  // --------------------------------------------------------------------------------
  /*
		opts: {
			identities: [					// put all identities that will have the ability to sign here, 1 for each
				{							// first identity that could sign
					role: {
						name: 'member',		// should match roles in channel, typically 'member' or 'admin'
						mspId: 'PeerOrg1'
					}
				},
				{							// second identity that could sign
					role: {
						name: 'member',		// should match roles in channel, typically 'member' or 'admin'
						mspId: 'PeerOrg2'
					}
				}
			],
			policy: {
				"<number>-of" : [			// <number> is the amount of signatures to get. eg "2-of" for 2 signatures
					{ 'signed-by': 0 },		// this is the array position of the identity in "identities" that can sign
					{ 'signed-by': 1 },		// this is the array position of the identity in "identities" that can sign
				]
			}
		}
	*/
  /* removed 05/04/2020 - use __b_build_signature_policy_envelope instead
	p_build_custom_e_policy_envelope_fabricSDK(opts: Bcp) {
		const principal_list = [];
		if (!opts || !opts.identities || !opts.policy) {			// basic input check
			logger.error('[protobuf-handler] cannot find "identities" or "policy" field in endorsement policy. :', opts);
			return null;
		}

		for (let i in opts.identities) {							// create principal list from identities
			const role_name = (opts.identities[i].role && opts.identities[i].role.name) ? opts.identities[i].role.name.toLowerCase() : null;
			const msp_id = (opts.identities[i].role) ? opts.identities[i].role.mspId : null;
			const p_MSPRole = this.p_build_msp_role({ role: role_name, msp_identifier: msp_id });

			if (p_MSPRole) {
				const p_MSPPrincipal = new MSP_Principle_MSPPrincipal();
				p_MSPPrincipal.setPrincipalClassification(MSP_Principle_MSPPrincipal.Classification.ROLE);
				p_MSPPrincipal.setPrincipal(p_MSPRole.serializeBinary());
				principal_list.push(p_MSPPrincipal);
			}
		}

		const p_signaturePolicy = this.p_build_policy(opts.policy, 0);	// create signature policy from policy
		if (!p_signaturePolicy) {
			logger.error('[protobuf-handler] policy field could not be built. the provided endorsement policy is not understood.');
			return null;
		} else {
			const bsp_opts = {
				version: 0,
				p_rule: p_signaturePolicy,
				p_identities: principal_list
			};
			const p_signaturePolicyEnvelope = this.p_build_signature_policy_envelope(bsp_opts);
			logger.debug('[protobuf-handler] p_signaturePolicyEnvelope?:', p_signaturePolicyEnvelope.toObject());
			return p_signaturePolicyEnvelope;
		}
	}*/

  // recursive function to format a policy
  p_build_policy(policy: any, depth: number) {
    if (depth >= 10000) {
      logger.error('[protobuf-handler] policy - field is too deeply nested, might be circular? aborting', depth);
      return null;
    } else if (!policy || Object.keys(policy).length === 0) {
      return null;						// sub policy does not exist - this is okay
    } else {

      // "signed-by" type of policy
      if (policy['signed-by'] >= 0) {
        const bsp_opts2 = {
          signed_by: Number(policy['signed-by']),
          n_out_of: null
        };
        logger.debug('[protobuf-handler] policy - making a signed_by', bsp_opts2);
        const p_signaturePolicy = this.p_build_signature_policy(bsp_opts2);
        return p_signaturePolicy;
      } else {

        // "n-of" type of policy
        const policy_name = Object.keys(policy)[0];
        const matches = policy_name.match(/^(\d+)-of/);								// parse "<number>-of" and get <number> out
        const signature_number = (matches && matches[1]) ? Number(matches[1]) : 1;		// make it a number

        const subPolicies = [];
        for (const i in policy[policy_name]) {											// build each sub policy
          const sub_policy = this.p_build_policy(policy[policy_name][i], ++depth);	// recursive!
          if (sub_policy) {
            subPolicies.push(sub_policy);
          }
        }

        const p_outOfAny = this.p_build_n_out_of({ n: signature_number, rules_list: subPolicies });
        logger.debug('[protobuf-handler] policy - signature_number', signature_number, 'p_outOfAny: ', (p_outOfAny) ? pp(p_outOfAny) : null);

        // final step, build the signature policy
        const bsp_opts2 = {
          signed_by: null,
          n_out_of: p_outOfAny
        };
        logger.debug('[protobuf-handler] policy - making a n_out_of', bsp_opts2);
        const p_signaturePolicy = this.p_build_signature_policy(bsp_opts2);
        return p_signaturePolicy;
      }
    }
  }

  // --------------------------------------------------------------------------------
  // build a implicit meta policy protobuf
  // --------------------------------------------------------------------------------
  p_build_implicit_meta_policy(opts: Bmp) {
    const rule_name = (opts && opts.rule) ? opts.rule.toUpperCase() : null;
    if (!rule_name || rules_map[rule_name] === undefined) {
      logger.error('[protobuf-handler] cannot find or invalid "rule". cannot build implicitMetaPolicy for signature policy. rule:', rule_name);
      return null;
    }
    const	rule = rules_map[rule_name];

    const p_implicitMetaPolicy = new Policies_ImplicitMetaPolicy();
    p_implicitMetaPolicy.rule = rule;
    if (opts.sub_policy) {
      p_implicitMetaPolicy.sub_policy = opts.sub_policy;
    } else if (opts.subPolicy) {
      p_implicitMetaPolicy.sub_policy = opts.subPolicy;
    } else {
      logger.warn('[protobuf-handler] there is no "subPolicy" field set for your implicit meta policy');
    }

    return p_implicitMetaPolicy;
  }

  // --------------------------------------------------------------------------------
  // build a signature policy protobuf as binary
  // --------------------------------------------------------------------------------
  __b_build_signature_policy_envelope(json: Bs2) {
    const envelope_alt = this.__build_signature_policy_envelope_alt(json);
    return envelope_alt ? <Uint8Array>Policy_SignaturePolicyEnvelope.encode(envelope_alt).finish() : null;
  }

  // --------------------------------------------------------------------------------
  // build a custom cc policy - using Fabric's format
  // --------------------------------------------------------------------------------
  // ! see docs/sig_policy_syntax.md for more information and examples !
  /*
	json: {
		version: 0,
		identities: [{
			principal_classification: 0,
			principal: {
				mspIdentifier: 'PeerOrg1',
				role: 'ADMIN'
			}
		}],
		rule: {												// rule contains either signedBy of nOutOf
			nOutOf: {
				n: 1,
				rules: [{									// rules contains either signedBy of nOutOf (recursive)
					signedBy: 0
				}]
			}
		}
	}*/

  /* removed 05/04/2020 - use __build_signature_policy_envelope_alt instead
	__build_signature_policy_envelope(json: Bs2) {
		if (!json.rule) {
			logger.error('[protobuf-handler] "rule" field not found in signature policy envelope');
			return null;
		} else if (!json.identities) {
			logger.error('[protobuf-handler] "identities" field not found in signature policy envelope');
			return null;
		} else {

			// convert principal field to binary [MSPRole]
			for (let i in json.identities) {
				if (json.identities[i].principal) {										// convert principal fields
					const opts = {
						role: json.identities[i].principal.role,
						msp_identifier: json.identities[i].principal.mspIdentifier		// rename msp id field
					};
					const p_MSPRole = this.p_build_msp_role(opts);
					if (p_MSPRole) {
						json.identities[i].principal = p_MSPRole.serializeBinary();		// overwrite json with binary of a msp role
					}
				}
			}

			const SignaturePolicyEnvelope = __pb_root.lookupType('common.SignaturePolicyEnvelope');
			let p_signaturePolicyEnvelope = SignaturePolicyEnvelope.fromObject(json);
			logger.debug('[protobuf-handler] p_signaturePolicyEnvelope?:', SignaturePolicyEnvelope.toObject(p_signaturePolicyEnvelope));
			return p_signaturePolicyEnvelope;
		}
	}*/

  // same as above, but it only use protobuf.js to build a signature policy envelope
  __build_signature_policy_envelope_alt(json: Bs2) {
    if (!json.rule) {
      logger.error('[protobuf-handler] "rule" field not found in signature policy envelope');
      return null;
    } else if (!json.identities) {
      logger.error('[protobuf-handler] "identities" field not found in signature policy envelope');
      return null;
    } else {
      const p_mspPrincipals = [];

      // convert the principal field to binary [MSPRole]
      for (const i in json.identities) {
        if (json.identities[i].principal) {
          const b_MSPRole = this.__b_build_msp_role(json.identities[i].principal);
          if (b_MSPRole) {
            const classification = json.identities[i].principal_classification || 0;	// default to 0, for the "ROLE" classification
            const p_mspPrincipal = this.__build_msp_principal({ principal_classification: classification, b_principal: b_MSPRole });
            p_mspPrincipals.push(p_mspPrincipal);				// create array of MSPPrincipal
          }
        }
      }

      const spe = {
        version: isNaN(json.version) ? 1 : Number(json.version), 	// int32
        identities: p_mspPrincipals, 								// repeated MSPPrincipal
        rule: json.rule, 											// SignaturePolicy
      };
      const p_signaturePolicyEnvelope = Policy_SignaturePolicyEnvelope.fromObject(spe);
      logger.debug('[protobuf-handler] p_signaturePolicyEnvelope?:', Policy_SignaturePolicyEnvelope.toObject(p_signaturePolicyEnvelope));
      return p_signaturePolicyEnvelope;
    }
  }

  // --------------------------------------------------------------------------------
  // decode
  // --------------------------------------------------------------------------------
  __decode_signature_policy_envelope(pb: Uint8Array, full: boolean) {
    const message = Policy_SignaturePolicyEnvelope.decode(pb);
    let obj = Policy_SignaturePolicyEnvelope.toObject(message, { defaults: false });

    if (obj && full === true) {				// fully decode is requested
      obj = this.decode_identities(obj.identities);
    }
    return obj;
  }

  decode_identities(identities: any) {
    if (identities) {
      for (const i in identities) {
        if (identities[i].principal) {
          const bin = identities[i].principal;
          identities[i].principal = this.__decode_msp_role(bin);

          if (identities[i].principal.role !== undefined) {
            for (const role in roles_map) {
              if (roles_map[role] === identities[i].principal.role) {	// convert the role back to the name not the integer
                identities[i].principal.role = role;
                break;
              }
            }
          }
        }
      }
    }
    return identities;
  }

  // --------------------------------------------------------------------------------
  // decode
  // --------------------------------------------------------------------------------
  __decode_msp_role(pb: Uint8Array) {
    const message = MSP_Principle_MSPRole.decode(pb);
    const obj = MSP_Principle_MSPRole.toObject(message, { defaults: false });
    return obj;
  }

  // --------------------------------------------------------------------------------
  // decode
  // --------------------------------------------------------------------------------
  __decode_implicit_policy(pb: Uint8Array) {
    const message = Policies_ImplicitMetaPolicy.decode(pb);
    const obj = Policies_ImplicitMetaPolicy.toObject(message, { defaults: true });		// must be true to work...
    return obj;
  }

  // -------------------------------------------------
  // build a protos.ApplicationPolicy message - returns message
  // -------------------------------------------------
  __build_application_policy(signature_policy: any, channel_config_policy_reference: any) {
    const opts: any = {};
    if (signature_policy) {														// message is of type "oneof", only 1 field should be set
      opts.signature_policy = signature_policy;
    } else {
      opts.channel_config_policy_reference = channel_config_policy_reference;
    }
    const message = common.ApplicationPolicy.create(opts);
    return message;
  }
}

const roles_map = <any>{
  'MEMBER': MSP_Principle_MSPRole.MSPRoleType.MEMBER,	// 0
  'ADMIN': MSP_Principle_MSPRole.MSPRoleType.ADMIN,	// 1
  'CLIENT': MSP_Principle_MSPRole.MSPRoleType.CLIENT,	// 2
  'PEER': MSP_Principle_MSPRole.MSPRoleType.PEER,		// 3
  'ORDERER': 4,	// 4
};

// find a role
function find_role(role_enum: any) {
  for (const key in roles_map) {
    if (roles_map[key] === role_enum) {
      return key;
    }
  }
  return 'MEMBER';								// default
}

const rules_map = <any>{
  'ANY': Policies_ImplicitMetaPolicy.Rule.ANY,
  'ALL': Policies_ImplicitMetaPolicy.Rule.ALL,
  'MAJORITY': Policies_ImplicitMetaPolicy.Rule.MAJORITY,
};

interface Bsp {
  n_out_of: Policy_SignaturePolicy.NOutOf | null;
	signed_by: number | null;
}

interface Bdp {
	msp_ids: string[];
}

interface Bse {
	version: number;
	p_rule: Policy_SignaturePolicy;
	p_identities: MSP_Principle_MSPPrincipal[];
}

interface B1o {
	n: number;
	rules_list: Policy_SignaturePolicy[];
}

/*
interface Bcp {
	identities: [
		{
			role: {
				name: string,
				mspId: string
			}
		}
	];
	policy: any;
}
*/

interface Bmp {
	rule: string;
	sub_policy: string | undefined;
	subPolicy: string | undefined;
}

interface Bs2 {
	version: number;
	rule: any;
	identities: any;
}

interface Bmr {
	role: string | null;
	msp_identifier: string | null;
}

export { rules_map, roles_map, Bse, find_role };