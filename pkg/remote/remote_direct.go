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
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// DoRemoteDirect bootstraps the ephemeral node.
func DoRemoteDirect(settings *environment.AirshipCTLSettings) error {
	cfg := settings.Config
	bootstrapSettings, err := cfg.CurrentContextBootstrapInfo()
	if err != nil {
		return err
	}

	remoteConfig := bootstrapSettings.RemoteDirect
	if remoteConfig == nil {
		return config.ErrMissingConfig{What: "RemoteDirect options not defined in bootstrap config"}
	}

	ephemeralNode, err := getEphemeralNode(settings)
	if err != nil {
		return err
	}

	log.Debug("Found ephemeral node with BMCAddress: %s", ephemeralNode.BMCAddress)

	// Perform remote direct operations
	if remoteConfig.IsoURL == "" {
		return ErrMissingBootstrapInfoOption{What: "isoURL"}
	}

	err = ephemeralNode.SetVirtualMedia(ephemeralNode.Context, remoteConfig.IsoURL)
	if err != nil {
		return err
	}

	log.Debugf("Successfully loaded virtual media: %q", remoteConfig.IsoURL)

	err = ephemeralNode.SetBootSourceByType(ephemeralNode.Context)
	if err != nil {
		return err
	}

	err = ephemeralNode.RebootSystem(ephemeralNode.Context)
	if err != nil {
		return err
	}

	log.Debug("Restarted ephemeral host")

	return nil
}

func getEphemeralNode(settings *environment.AirshipCTLSettings) (baremetalHost, error) {
	adapter, err := NewAdapter(settings, ByLabel(document.EphemeralHostSelector))
	if err != nil {
		return baremetalHost{}, nil
	}

	if len(adapter.Hosts) == 1 {
		return baremetalHost{}, NewRemoteDirectErrorf("more than one node defined as the ephemeral node")
	}

	return adapter.Hosts[0], nil
}
