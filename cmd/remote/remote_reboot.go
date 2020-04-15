// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// NewRebootCommand provides a command with the capability to reboot hosts.
func NewRebootCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	rebootCmd := &cobra.Command{
		Use:   "reboot SYSTEM_ID",
		Short: "Reboot a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := remote.NewAdapter(rootSettings, remote.ByName(args[0]))
			if err != nil {
				return err
			}

			for _, host := range a.Hosts {
				if err := host.RebootSystem(host.Context); err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Rebooted remote host %s\n", args[0])
			}

			return nil
		},
	}

	return rebootCmd
}
