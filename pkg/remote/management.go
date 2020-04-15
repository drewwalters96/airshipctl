/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package remote manages baremetal hosts.
package remote

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// Client is a set of functions that clients created for out-of-band power management and control should implement. The
// functions within client are used by power management commands and remote direct functionality.
type Client interface {
	RebootSystem(context.Context) error

	SystemPowerOff(context.Context) error

	// TODO(drewwalters96): Should this be a string forever? We may want to define our own custom type, as the
	// string format will be client dependent when we add new clients.
	SystemPowerStatus(context.Context) (string, error)

	NodeID() string

	SetBootSourceByType(context.Context) error

	// TODO(drewwalters96): This function is tightly coupled to Redfish. It should be combined with the
	// SetBootSource operation and removed from the client interface.
	SetVirtualMedia(context.Context, string) error
}

// Adapter bridges the gap between out-of-band clients. It can hold any type of OOB client, e.g. Redfish. An adapter
// can only communicate with one host.
type Adapter struct {
	config.ManagementConfiguration
	Hosts   []baremetalHost
}

// baremetalHost is an airshipctl representation of a baremetal host, defined by a baremetal host document, that embeds
// actions an out-of-band client can perform. Once instantiated, actions can be performed on a baremetal host.
type baremetalHost struct {
	Client
	Context context.Context
	BMCAddress string
	username   string
	password   string
}

type HostSelector func(*Adapter, config.ManagementConfiguration, document.Bundle) error

// ByLabel adds all hosts to an adapter whose documents match a supplied label selector.
func ByLabel(label string) HostSelector {
	return func(a *Adapter, mgmtCfg config.ManagementConfiguration, docBundle document.Bundle) error {
		selector := document.NewSelector().ByKind(document.BareMetalHostKind).ByLabel(label)
		docs, err := docBundle.Select(selector)
		if err != nil {
			return err
		}

		for _, doc := range docs {
			host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
			if err != nil {
				return err
			}

			a.Hosts = append(a.Hosts, host)
		}

		return nil
	}
}

// ByName adds the host to an adapter whose document meets the specified name.
func ByName(name string) HostSelector {
	return func(a *Adapter, mgmtCfg config.ManagementConfiguration, docBundle document.Bundle) error {
		selector := document.NewSelector().ByKind(document.BareMetalHostKind).ByName(name)
		doc, err := docBundle.SelectOne(selector)
		if err != nil {
			return err
		}

		host, err := newBaremetalHost(mgmtCfg, doc, docBundle)
		if err != nil {
			return err
		}

		a.Hosts = append(a.Hosts, host)

		return nil
	}
}

// NewAdapter provides an adapter that exposes the capability to perform remote direct functionality and remote
// management on multiple hosts.
//
// Examples:
//
//     Create an adapter to interact with hosts matching a label selector "label=example-label":
//             a := NewAdapter(ByLabel("label=example-label"))
//
//     Create an adapter to interact with a host matching the name "air-ephemeral":
//             a := NewAdapter(ByName("air-ephemeral"))
//
//     Reboot all hosts in an adapter:
//         for _, host := range a.hosts {
//             host.reboot(a.context)
//         }
//
//     You may also select by name and label by passing both functions.
//             a := NewAdapter(ByLabel("label=example-label"), ByName("air-epehemeral"))
func NewAdapter(settings *environment.AirshipCTLSettings, hostSelections ...HostSelector) (*Adapter, error) {
	managementCfg, err := settings.Config.CurrentContextManagementConfig()
	if err != nil {
		return nil, err
	}

	a := &Adapter{
		*managementCfg,
		make([]baremetalHost, 1),
	}

	docBundle, err := document.NewBundleByPath("/")
	if err != nil {
		return nil, err
	}

	// Each function in hostSelections modifies the list of hosts for the new adapter based on selection criteria
	// provided by CLI arguments and airshipctl settings.
	for _, addHost := range hostSelections {
		if err := addHost(a, *managementCfg, docBundle); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// newBaremetalHost creates a representation of a baremetal host that is configured to perform management actions by
// invoking its client methods (provided by the remote.Client interface).
func newBaremetalHost(
	mgmtCfg config.ManagementConfiguration,
	hostDoc document.Document,
	docBundle document.Bundle) (baremetalHost, error) {
	var host baremetalHost

	ip, err := document.GetBMHBMCAddress(hostDoc)
	if err != nil {
		return host, err
	}

	username, password, err := document.GetBMHBMCCredentials(hostDoc, docBundle)
	if err != nil {
		return host, err
	}

	// Select the client that corresponds to the management type specified in the airshipctl config.
	switch mgmtCfg.Type {
	case redfish.ClientType:
		ctx, client, err := redfish.NewClient(
				hostDoc.GetName(),
				ip,
				mgmtCfg.Insecure,
				mgmtCfg.UseProxy,
				username,
				password)

		if err != nil {
			return host, err
		}

		host = baremetalHost{client, ctx, ip, username, password}
	default:
		return host, ErrUnknownManagementType{Type: string(mgmtCfg.Type)}
	}

	return host, nil
}
