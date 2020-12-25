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

func TestMultipleOpsCmd(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "configtx-execute-multiple-ops-cmd-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	// cases
	cases := []struct {
		command         string
		expectedMessage string
		checkOutputFile bool
	}{
		{command: "fabric-configtx-cli execute-multiple-ops", expectedMessage: "the required parameter 'channelID' is empty"},
		{command: "fabric-configtx-cli execute-multiple-ops --channelID mychannel", expectedMessage: "the required parameter 'blockPath' is empty"},
		{command: "fabric-configtx-cli execute-multiple-ops --channelID mychannel --blockPath test", expectedMessage: "the required parameter 'profile' is empty"},
		{command: fmt.Sprintf("fabric-configtx-cli execute-multiple-ops --channelID mychannel --blockPath ../ops/testdata/mychannel_config.block --profile ../ops/testdata/wrong_profile.yaml --outputDir %s", tempDir),
			expectedMessage: "failed to output configtx to execute multiple ops"},
		{command: fmt.Sprintf("fabric-configtx-cli execute-multiple-ops --channelID mychannel --blockPath ../ops/testdata/mychannel_config.block --profile ../ops/testdata/multiple_ops_profile_for_mychannel_without_reading_other_files.yaml --outputDir %s", tempDir),
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
