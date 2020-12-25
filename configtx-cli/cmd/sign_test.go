/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignCmd(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "configtx-sign-cmd-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseMSPID := "Org1MSP"
	baseOrgLabel := "org1"
	baseSignKeyPath := fmt.Sprintf("../ops/testdata/peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/keystore/priv_sk", baseOrgLabel, baseOrgLabel)
	baseSignCertPath := fmt.Sprintf("../ops/testdata/peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/signcerts/cert.pem", baseOrgLabel, baseOrgLabel)

	// cases
	cases := []struct {
		command         string
		expectedMessage string
		checkOutputFile bool
	}{
		{command: "fabric-configtx-cli sign", expectedMessage: "the required parameter 'mspID' is empty"},
		{command: fmt.Sprintf("fabric-configtx-cli sign --mspID %s", baseMSPID), expectedMessage: "the required parameter 'keyPath' is empty"},
		{command: fmt.Sprintf("fabric-configtx-cli sign --mspID %s --keyPath %s", baseMSPID, baseSignKeyPath), expectedMessage: "the required parameter 'certPath' is empty"},
		{command: fmt.Sprintf("fabric-configtx-cli sign --mspID %s --keyPath %s --certPath %s", baseMSPID, baseSignKeyPath, baseSignCertPath), expectedMessage: "the required parameter 'configTxPath' is empty"},
		{command: fmt.Sprintf("fabric-configtx-cli sign --mspID %s --keyPath %s --certPath %s --configTxPath ../ops/testdata/wrong.pb --outputDir %s", baseMSPID, baseSignKeyPath, baseSignCertPath, tempDir),
			expectedMessage: "failed to output configtx to sign"},
		{command: fmt.Sprintf("fabric-configtx-cli sign --mspID %s --keyPath %s --certPath %s --configTxPath ../ops/testdata/set_mychannel_expected.pb --outputDir %s", baseMSPID, baseSignKeyPath, baseSignCertPath, tempDir),
			expectedMessage: "", checkOutputFile: true},
	}

	for _, c := range cases {
		os.Remove(filepath.Join(tempDir, "output.pb"))

		outBuf := new(bytes.Buffer)
		cmd := NewRootCmd()
		cmd.SetOutput(outBuf)
		cmdArgs := strings.Split(c.command, " ")
		fmt.Printf("commands args: %v\n", cmdArgs)
		cmd.SetArgs(cmdArgs[1:])
		cmd.Execute()

		actual := outBuf.String()
		require.Contains(t, actual, c.expectedMessage)

		if c.checkOutputFile {
			_, err := os.Stat(filepath.Join(tempDir, "output.pb"))
			require.NoError(t, err)
		}
	}
}
