/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"testing"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestConfig(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-with-defaults",
			CmdLine: "",
			Cmd:     cmd.NewConfigCommand(nil),
		},
		{
			Name:    "config-cmd-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewConfigCommand(nil),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}
