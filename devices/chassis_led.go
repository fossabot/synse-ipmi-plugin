package devices

import (
	"fmt"
	"strings"

	"github.com/vapor-ware/synse-ipmi-plugin/protocol"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// BmcChassisLed is the handler for the bmc-boot-target device.
//
// This is really chassis identify, which according to the IPMI spec:
//
//   "This command causes the chassis to physically identify itself by a mechanism
//   chosen by the system implementation; such as turning on blinking user-visible
//   lights or emitting beeps via a speaker, LCD panel, etc" -- 28.5 Chassis Identify Command
//
// This was considered LED in Synse 1.4 so we will continue to consider it
// an LED device, even though it may not be.
var BmcChassisLed = sdk.DeviceHandler{
	Type:  "led",
	Model: "bmc-chassis-led",

	Read:  bmcChassisRead,
	Write: bmcChassisWrite,
}

func bmcChassisRead(device *sdk.Device) ([]*sdk.Reading, error) {

	state, err := protocol.GetChassisIdentify(device.Data)
	if err != nil {
		return nil, err
	}

	ret := []*sdk.Reading{
		sdk.NewReading("state", state),
	}

	return ret, nil
}

func bmcChassisWrite(device *sdk.Device, data *sdk.WriteData) error {

	action := data.Action
	raw := data.Raw

	// When writing to a BMC LED (identify) device, we always expect there to be
	// raw data specified. If there isn't, we return an error.
	if len(raw) == 0 {
		return fmt.Errorf("no values specified for 'raw', but required")
	}

	if action == "state" {
		cmd := string(raw[0])

		var state protocol.IdentifyState
		// TODO (etd): figure out if we want to support intervals. if so, how? could be
		// its own action (LED interval).. could be a second value in the raw list (["on", "20"]),
		// could be a regular raw value for state here ({"state": "20"})
		switch strings.ToLower(cmd) {
		case "on":
			state = protocol.IdentifyOn
		case "off":
			state = protocol.IdentifyOff
		default:
			return fmt.Errorf("unsupported command for bmc chassis led (identify) 'state' action: %s", cmd)
		}

		err := protocol.SetChassisIdentify(device.Data, state)
		if err != nil {
			return err
		}
	} else {
		// If we reach here, then the specified action is not supported.
		return fmt.Errorf("action '%s' is not supported for bmc chassis led (identify) devices", action)
	}

	return nil
}
