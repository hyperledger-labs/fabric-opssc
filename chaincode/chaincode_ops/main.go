/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger-labs/fabric-opssc/chaincode/chaincode_ops/core"
)

func main() {
	s, err := contractapi.NewChaincode(&core.SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode operation smart contract: %v", err)
	}

	if err := s.Start(); err != nil {
		log.Panicf("Error starting chaincode operation smart contract: %v", err)
	}
}
