# API specification

------------------------------------------------------------------------------
## Chaincode category

### Get all update proposals

* **URL**

  `/api/v1/chaincode/proposals`

* **Method:**

  `GET`

* **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** ``

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Request a new update proposal

* **URL**

  `/api/v1/chaincode/proposals/:id`

* **Method:**

  `POST`

* **URL Params**

  None

* **Data Params**

  **Required:**

   ```json
    {
    "proposal": {
      "channelID": "mychannel",
      "chaincodeName": "basic",
      "chaincodePackage": {
        "repository": "github.com/hyperledger-labs/fabric-opssc",
        "pathToSourceFiles": "sample-environments/fabric-samples/asset-transfer-basic/chaincode-go",
        "commitID": "main",
        "type": "golang"
      },
      "chaincodeDefinition": {
        "sequence": 1,
        "initRequired": false,
        "validationParameter": "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="
      }
    }
  }
  ```

`validationParameter` should be base64 encoded.

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The requested proposal
    ```json
    {
      "docType": "proposal",
      "ID": "deploy_basic",
      "creator": "Org1MSP",
      "channelID": "mychannel",
      "chaincodeName": "basic",
      "chaincodePackage": {
      },
      "chaincodeDefinition": {
      },
      "status": "proposed",
      "time": "2020-..."
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Get the proposal with the given ID

* **URL**

  `/api/v1/chaincode/proposals/:id`

* **Method:**

  `GET`

* **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The proposal with `id`
    ```json
    {
      "docType": "proposal",
      "ID": "deploy_basic",
      "creator": "Org1MSP",
      "channelID": "mychannel",
      "chaincodeName": "basic",
      "chaincodePackage": {
      },
      "chaincodeDefinition": {
      },
      "status": "proposed",
      "time": "2020-..."
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Vote for/against the proposal

* **URL**

  `/api/v1/chaincode/proposals/:id/vote`

* **Method:**

  `POST`

* **URL Params**

  None

* **Data Params**

  By default (if No data params are specified), the request votes for the proposal.

  * **Optional:**

    ```json
    {
      "updateRequest": {
        "status": "agreed", // ["agreed"|"disagreed"]`
        "data": "messages"
      }
    }
    ```

* **Success Response**

  * **Code:** 200 <br />
    **Content:** None

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Withdraw the proposal

* **URL**

  `/api/v1/chaincode/proposals/:id/withdraw`

* **Method:**

  `POST`

* **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** None

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Get the task histories with the given proposal

* **URL**

  `/api/v1/chaincode/proposals/:id/histories`

* **Method:**

  `GET`

*  **URL Params**

    * **Optional:**
      `taskID=["vote"|"acknowledge"|"commit"]`

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The task histories with `id` and `taskID`
    ```json
    {
      "\u0000history\u0000deploy_basic\u0000vote\u0000Org1MSP\u0000": {
        "docType": "history",
        "proposalID": "deploy_basic",
        "taskID": "vote",
        "orgID": "Org1MSP",
        "status": "agreed",
        "data": "",
        "time": "2020-..."
      },
      "\u0000history\u0000deploy_basic\u0000vote\u0000Org2MSP\u0000": {
        "docType": "history",
        "proposalID": "deploy_basic",
        "taskID": "vote",
        "orgID": "Org2MSP",
        "status": "agreed",
        "data": "",
        "time": "2020-..."
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Get the list of installed chaincodes

* **URL**

  `/api/v1/chaincode/getInstalledChaincodes`

* **Method:**

  `GET`

* **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The installed chaincodes
    ```json
    {
      "installed_chaincodes": [
        {
          "package_id": "chaincode_ops_1:bcc47dda45d52541f753c1c741af1d9da9fe6225ae0aaac24a867138c4fda469",
          "label": "chaincode_ops_1",
          "references": {
            "ops-channel": {
              "chaincodes": [
                {
                  "name": "chaincode_ops",
                  "version": "1"
                }
              ]
            }
          }
        },
      ]
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Get the chaincode definition with the given channel and chaincode name

* **URL**

  `/api/v1/chaincode/queryChaincodeDefinition`

* **Method:**

  `GET`

* **URL Params**

    * **Required:**
      `channelID=[string]`</br>
      `chaincodeName=[string]`

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The chaincode definition with `channelID` and `chaincodeName`
    ```json
    {
      "sequence": "1",
      "version": "1",
      "endorsement_plugin": "escc",
      "validation_plugin": "vscc",
      "validation_parameter": "EiAvQ2hhbm5lbC9BcHBsaWNhdGlvbi9FbmRvcnNlbWVudA==",
      "collections": {},
      "approvals": {
        "Org1MSP": true,
        "Org2MSP": true
      }
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Get the chaincode definitions with the given channel

* **URL**

  `/api/v1/chaincode/queryChaincodeDefinitions`

* **Method:**

  `GET`

*  **URL Params**

    * **Required:**
      `channelID=[string]`

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The list of the chaincode definitions with `channelID`
    ```json
    {
      "chaincode_definitions": [
        {
          "name": "basic",
          "sequence": "1",
          "version": "1",
          "endorsement_plugin": "escc",
          "validation_plugin": "vscc",
          "validation_parameter": "EiAvQ2hhbm5lbC9BcHBsaWNhdGlvbi9FbmRvcnNlbWVudA==",
          "collections": {}
        }
      ]
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

------------------------------------------------------------------------------
## Channel category

### Get all update proposals

* **URL**

  `/api/v1/channel/proposals`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The map of all the proposals
    ```json
    {
      "create_mychannel": {
        "docType": "proposal",
        "ID": "create_mychannel",
        "channelID": "mychannel",
        "description": "Create mychannel",
        "creator": "Org1MSP",
        "action": "create",
        "status": "committed",
        "opsProfile": {
        },
        "artifacts": {
          "configUpdate": "CglteWNoYW5uZWwSOxIpCgt(...)",
          "signatures": {
            "Org1MSP": "CoYICukHCgd(...)",
            "Org2MSP": "CoIICuUHCgd(...)"
          }
        }
      }
    }
    ```
  A key of the map is a proposal ID.

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Request a new update proposal

This is API to request a uew update proposal.
The server internally creates the config update and the signature by using `opsProfile` parameter.

* **URL**

  `/api/v1/channel/proposals/:id`

* **Method:**

  `POST`

*  **URL Params**

  None

* **Data Params**

  * **Required:**

    ```json
    "proposal": {
      "channelID": "mychannel",
      "description": "Create mychannel",
      "action": "create",
      "opsProfile": {
      }
    }
    ```

    `action` should be `create` if creating a channel or should be `update` if updating a channel.<br/>
    `opsProfile` for `create` should be [the format](../configtx-cli/ops/testdata/create_mychannel2_profile.yaml) for `fabric-configtx-cli create-channel`.<br/>
    `opsProfile` for `update` should be [the format](../configtx-cli/ops/testdata/multiple_ops_profile_for_mychannel_without_reading_other_files.yaml) for `fabric-configtx-cli execute-multiple-ops`.<br/>

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The proposal ID<br />
    `"create_mychannel"`

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Get the proposal with the given ID

* **URL**

  `/api/v1/channel/proposals/:id`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The the proposal with ``
    ```json
    {
      "docType": "proposal",
      "ID": "create_mychannel",
      "channelID": "mychannel",
      "description": "Create mychannel",
      "creator": "Org1MSP",
      "action": "create",
      "status": "committed",
      "opsProfile": {
      },
      "artifacts": {
        "configUpdate": "CglteWNoYW5uZWw...",
        "signatures": {
          "Org1MSP": "CoYICukHCgdPcmcx...",
          "Org2MSP": "CoIICuUHCgdPcmcy..."
        }
      }
    }
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Vote for the proposal

This is API to vote for the proposal.
The server internally creates the signature by using the content of the proposal.

* **URL**

  `/api/v1/channel/proposals/:id/vote`

* **Method:**

  `POST`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** None

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Get the list of all channels

* **URL**

  `/api/v1/channel/proposals/getChannels`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The list of all the channel information
    ```json
    [
      {
        "docType": "channel",
        "ID": "mychannel",
        "channelType": "application",
        "organizations": {
          "Org1MSP": "",
          "Org2MSP": ""
        }
      },
      {
        "docType": "channel",
        "ID": "ops-channel",
        "channelType": "ops",
        "organizations": {
          "Org1MSP": "",
          "Org2MSP": ""
        }
      },
      {
        "docType": "channel",
        "ID": "system-channel",
        "channelType": "system",
        "organizations": {
          "Org1MSP": "",
          "Org2MSP": ""
        }
      }
    ]
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Get the system config block

* **URL**

  `/api/v1/channel/proposals/systemConfigBlock`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The system channel config block <br />
    `"CiIaIO26iOmkIsGkYxJKmTdOb0KTu+zOKkdn6FlIlWvZ5e2NEoTSA..."`

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

------------------------------------------------------------------------------
## Utility category

### Invoke a chaincode (for test)

* **URL**

  `/api/v1/chaincode/invokeTransaction`

* **Method:**

  `POST`

*  **URL Params**

    * **Required:**
      `channelID=[string]`<br/>
      `ccName=[string]`<br/>
      `func=[string]`<br/>
      `args=[string array]`

* **Data Params**

  ```json
  {
    "channelID": "mychannel",
    "ccName": "basic",
    "func": "createCar",
    "args": ["CAR0", "AAA", "BBB", "Blue", "Mary"]
  }
  ```

* **Success Response**

  * **Code:** 200 <br />
    **Content:** None

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Query a chaincode (for test)

* **URL**

  `/api/v1/chaincode/queryTransaction`

* **Method:**

  `GET`

*  **URL Params**

    * **Required:**
      `channelID=[string]`<br/>
      `ccName=[string]`<br/>
      `channelID=[string]`<br/>

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The result of querying the chaincode
    ```json
    [
      {
        "Key": "CAR0",
        "Record": {
          "make": "AAA",
          "model": "BBB",
          "colour": "Blue",
          "owner": "Mary"
        }
      }
    ]
    ```

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`


### Version check

* **URL**

  `/api/v1/version`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** The version of the API server <br />
    `"0.2.0"`

* **Error Response:**

  * **Code:** 500 Internal Server Error <br />
    **Content:** `{ "message" : "..." }`

### Health check

* **URL**

  `/healthz`

* **Method:**

  `GET`

*  **URL Params**

  None

* **Data Params**

  None

* **Success Response**

  * **Code:** 200 <br />
    **Content:** `"OK"`

* **Error Response:**

  * **Code:**  <br />
    **Content:**
