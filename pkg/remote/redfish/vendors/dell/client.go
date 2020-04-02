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

// Package dell wraps the standard Redfish client in order to provide additional functionality required to perform
// actions on iDRAC servers.
package dell

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType                 string = "redfish-dell"
)

// Client is a wrapper around the standard airshipctl Redfish client. This allows vendor specific Redfish clients to
// override methods without duplicating the entire client.
type Client struct {
	client redfish.Client
}

// EphemeralNodeID retrieves the ephemeral node ID.
func (c *Client) EphemeralNodeID() string {
	return c.client.EphemeralNodeID()
}

// RebootSystem power cycles a host by sending a shutdown signal followed by a power on signal.
func (c *Client) RebootSystem(ctx context.Context, systemID string) error {
	return c.client.RebootSystem(ctx, systemID)
}

// SetEphemeralBootSourceByType sets the boot source of the ephemeral node to a virtual CD, "VCD-DVD".
func (c *Client) SetEphemeralBootSourceByType(ctx context.Context) error {
	return c.client.SetEphemeralBootSourceByType(ctx)
}

// SystemPowerOff shuts down a host.
func (c *Client) SystemPowerOff(ctx context.Context, systemID string) error {
	return c.client.SystemPowerOff(ctx, systemID)
}

// SystemPowerStatus retrieves the power status of a host as a human-readable string.
func (c *Client) SystemPowerStatus(ctx context.Context, systemID string) (string, error) {
	return c.client.SystemPowerStatus(ctx, systemID)
}

// SetVirtualMedia injects a virtual media device to an established virtual media ID. This assumes that isoPath is
// accessible to the redfish server and virtualMedia device is either of type CD or DVD.
func (c *Client) SetVirtualMedia(ctx context.Context, isoPath string) error {
	return c.client.SetVirtualMedia(ctx, isoPath)
}

// NewClient returns a client with the capability to make Redfish requests.
func NewClient(ephemeralNodeID string,
	isoPath string,
	redfishURL string,
	insecure bool,
	useProxy bool,
	username string,
	password string) (context.Context, *Client, error) {
	ctx, genericClient, err := redfish.NewClient(
		ephemeralNodeID, isoPath, redfishURL, insecure, useProxy, username, password)
	if err != nil {
		return ctx, nil, err
	}

	c := &Client{*genericClient}

	return ctx, c, nil
}
