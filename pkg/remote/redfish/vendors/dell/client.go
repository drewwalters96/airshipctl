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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	ClientType = "redfish-dell"

	configurationVirtualCD = `<SystemConfiguration><Component FQDD=\"iDRAC.Embedded.1\"><Attribute Name=\"ServerBoot.1#BootOnce\">%s</Attribute><Attribute Name=\"ServerBoot.1#FirstBootDevice\">VCD-DVD</Attribute></Component></SystemConfiguration>` // nolint
	endpointImportSysCFG   = "%s/redfish/v1/Managers/%s/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"
)

// Client is a wrapper around the standard airshipctl Redfish client. This allows vendor specific Redfish clients to
// override methods without duplicating the entire client.
type Client struct {
	client redfish.Client
}

type iDRACAPIRespErr struct {
	Err iDRACAPIErr `json:"error"`
}

type iDRACAPIErr struct {
	ExtendedInfo []iDRACAPIExtendedInfo `json:"@Message.ExtendedInfo"`
	Code         string                 `json:"code"`
	Message      string                 `json:"message"`
}

type iDRACAPIExtendedInfo struct {
	Message    string `json:"Message"`
	Resolution string `json:"Resolution,omitempty"`
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
	managerID, err := redfish.GetManagerID(ctx, c.client.RedfishAPI, c.client.EphemeralNodeID())
	if err != nil {
		return err
	}

	// NOTE(drewwalters96): Setting the boot device to a virtual media type requires an API request to the iDRAC
	// actions API. The request is made below using the same HTTP client used by the Redfish API and exposed by the
	// standard airshipctl Redfish client. Only iDRAC 9 >= 3.3 is supports this endpoint.
	url := fmt.Sprintf(endpointImportSysCFG, c.client.RedfishCFG.BasePath, managerID)
	body := fmt.Sprintf(`{"ShareParameters":{"Target": "ALL"},"ImportBuffer": %s"}`, configurationVirtualCD)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	if auth, ok := ctx.Value(redfishClient.ContextBasicAuth).(redfishClient.BasicAuth); ok {
		req.SetBasicAuth(auth.UserName, auth.Password)
	}

	httpResp, err := c.client.RedfishCFG.HTTPClient.Do(req)
	if httpResp.StatusCode != http.StatusAccepted {
		body, ok := ioutil.ReadAll(httpResp.Body)
		if ok != nil {
			log.Debugf("Malformed iDRAC response: %s", body)
			return redfish.ErrRedfishClient{Message: "Unable to set boot device. Malformed iDRAC response."}
		}

		var iDRACResp iDRACAPIRespErr
		ok = json.Unmarshal(body, &iDRACResp)
		if ok != nil {
			log.Debugf("Malformed iDRAC response: %s", body)
			return redfish.ErrRedfishClient{Message: "Unable to set boot device. Malformed iDrac response."}
		}

		return redfish.ErrRedfishClient{
			Message: fmt.Sprintf("Unable to set boot device. %s", iDRACResp.Err.ExtendedInfo[0]),
		}
	} else if err != nil {
		return redfish.ErrRedfishClient{Message: fmt.Sprintf("Unable to set boot device. %v", err)}
	}

	defer httpResp.Body.Close()

	return nil
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
