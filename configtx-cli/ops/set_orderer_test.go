/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-config/protolator"
	"github.com/stretchr/testify/require"
)

func TestOutputConfigTXToSetOrderer(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-set-orderer-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseBlockPath := "testdata/mychannel_config.block"
	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseChannelID := "mychannel"
	baseProfilePath := "testdata/updated_orderer_profile.yaml"

	// Case: Positive test cases
	cases := []struct {
		channelID string
	}{
		{channelID: "mychannel"},
		{channelID: "system-channel"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToSetOrderer(fmt.Sprintf("testdata/%s_config.block", c.channelID), tempDir, outputFileName, baseOutputFormat, c.channelID, baseProfilePath)
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/set_orderer_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)

		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to set %s", c.channelID)
	}

	// Case: Using wrong block path
	err = OutputConfigTXToSetOrderer("testdata/wrong-path.block", tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get configtx from block")

	// Case: Using wrong profile path
	err = OutputConfigTXToSetOrderer(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/wrong-path.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load orderer profile")

	// Case: Using wrong output format
	err = OutputConfigTXToSetOrderer(baseBlockPath, tempDir, baseOutputFileName, "wrong format", baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
