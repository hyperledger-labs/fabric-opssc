# Voting Config Option

`chaincode_ops` chaincode provides a voting config option.
This allows OpsSC users to configure the maximum number of malicious organizations (`f`) in the voting process.
This config option is not yet supported for channel_ops.
- If the option is set, `2f + 1` is required to judge a proposal gets `Approved`.
- If the option is not set, a majority of all participating organizations is required to judge a proposal gets `Approved`.

You can use this to call the following CC functions in `chaincode_ops` chaincode.
- `SetMaxMaliciousOrgsInVotes()`: sets number of max malicious orgs in votes.
- `UnsetMaxMaliciousOrgsInVotes()`: unsets number of max malicious orgs in votes.
- `GetVotingConfig()`: returns the voting config.

## Example: Skip the Voting Process

By using this, it is possible to configure to skip the voting process from other organizations for chaincode proposals.
With this configuration, when a proposal request is sent by one organization via the REST API of the OpsSC API server,
the deployment of the proposed chaincode is automatically executed immediately.

The following command is an example by invoking the SC function via an OpsSC API server:

```bash
curl -X POST "http://localhost:3000/api/v1/utils/invokeTransaction" \
-H "Expect:" \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<EOF
{
  "channelID": "ops-channel",
  "ccName": "chaincode_ops",
  "func": "SetMaxMaliciousOrgsInVotes",
  "args": ["0"]
}
EOF
```

_NOTE:_  In this configuration, chaincode is deployed by a single organization's decision (it means a centralized manner).
Note that this means that OpsSC is not used for the decentralized operations that it originally envisioned, which partially undermines the value it provides.
Of course, it is useful for testing or developing phase, etc.