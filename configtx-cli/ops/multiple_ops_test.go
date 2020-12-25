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

func TestOutputConfigTXToMultipleOps(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-multipleops-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseBlockPath := "testdata/mychannel_config.block"
	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseChannelID := "mychannel"
	baseProfilePath := "testdata/multiple_ops_profile_for_mychannel.yaml"

	// Case: Positive test cases
	cases := []struct {
		channelID string
	}{
		{channelID: "mychannel"},
		{channelID: "system-channel"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToDoMultipleOps(fmt.Sprintf("testdata/%s_config.block", c.channelID), tempDir, outputFileName, baseOutputFormat,
			c.channelID, fmt.Sprintf("testdata/multiple_ops_profile_for_%s.yaml", c.channelID))
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/multiple_ops_to_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)

		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to do multiple ops to %s", c.channelID)
	}

	// Case: Using wrong block path
	err = OutputConfigTXToDoMultipleOps("testdata/wrong-path.block", tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get configtx from block")

	// Case: Using wrong profile path
	err = OutputConfigTXToDoMultipleOps(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/wrong-path.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load multiple ops profile")

	// Case: Using wrong output format
	err = OutputConfigTXToDoMultipleOps(baseBlockPath, tempDir, baseOutputFileName, "wrong format", baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
