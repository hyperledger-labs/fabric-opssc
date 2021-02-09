/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger-labs/fabric-opssc/chaincode/channel_ops/chaincode"
)

func main() {
	s, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating channel operation smart contract: %v", err)
	}

	if err := s.Start(); err != nil {
		log.Panicf("Error starting channel operation smart contract: %v", err)
	}
}
