package main

import (
	"encoding/hex"
	"flag"
	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/agent"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	adapter = flag.String("adapter", "hci0", "Bluetooth adapter")
	btName  = flag.String("btName", "", "Set Bluetooth name (optional)")
)

// TODO: HID over GATT (HOGP) over BR/EDR vs. BR/EDR HID Profile?
//
// Most BT HID emulation examples use plain BR/EDR profiles. This uses GATT.
// Are there any particular security concerns with GATT or GATT over EDR?
//
const (
	// HID-over-GATT (HID-over-EDR is 1124)
	uuidHID = "1812"

	// HID characteristics
	charHIDInfo         = "2A4A"
	charHIDReportMap    = "2A4B"
	charHIDControlPoint = "2A4C"
	charHIDReport       = "2A4D"
	charHIDProtocolMode = "2A4E"

	// Descriptors on charHIDReport
	descrCCC             = "2902"
	descrReportReference = "2908"
)

const (
	// Device Information
	uuidDeviceInfo = "180A"

	// Device Info characteristics
	charDeviceInfoVendor  = "2A29"
	charDeviceInfoProduct = "2A24"
	charDeviceInfoVersion = "2A28"
	charDeviceInfoPNPID   = "2A50"
)

func init() {
	flag.Parse()
	log.SetLevel(log.TraceLevel)
}

func hidService(adapterID string) error {
	options := service.AppOptions{
		AdapterID: adapterID,
		AgentCaps: agent.CapKeyboardDisplay, // if we want interactive pairing via our client, change this
		UUID:      "0000",
	}

	a, err := service.NewApp(options)
	if err != nil {
		return err
	}
	defer a.Close()

	if *btName != "" {
		log.Infof("Setting device name to %s", *btName)
		a.SetName(*btName)
	}

	log.Infof("HW address %s", a.Adapter().Properties.Address)

	if !a.Adapter().Properties.Powered {
		if err := a.Adapter().SetPowered(true); err != nil {
			return err
		}
	}

	if err := registerDeviceInformation(a); err != nil {
		return err
	}

	if err := registerHIDService(a); err != nil {
		return err
	}

	if err := a.Run(); err != nil {
		return err
	}

	timeout := 10 * time.Minute
	log.Infof("Advertising for %v", timeout)
	cancel, err := a.Advertise(uint32(timeout.Seconds()))
	if err != nil {
		return err
	}

	defer cancel()

	select {}
}

func registerHIDService(a *service.App) error {
	svc, err := a.NewService(uuidHID)
	if err != nil {
		return err
	}

	if err := a.AddService(svc); err != nil {
		return err
	}

	// Inspired from (thanks!):
	// https://github.com/HeadHodge/Bluez-HID-over-GATT-Keyboard-Emulator-Example/blob/main/gattServer.py

	// Register Protocol Mode characteristic
	if err := registerStaticCharacteristic(svc, charHIDProtocolMode, []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWriteWithoutResponse,
	}, []byte{
		// Protocol Mode Value
		//
		// 0x00 = Boot Protocol Mode
		// 0x01 = Report Protocol Mode
		0x01,
	}); err != nil {
		return err
	}

	// Register HID Info characteristic
	if err := registerStaticCharacteristic(svc, charHIDInfo, []string{
		gatt.FlagCharacteristicRead,
	}, []byte{
		// bcdHID: protocol version number
		0x01, 0x11,
		// bCountryCode: localization
		0x00, // None
		// Flags
		0x02, // RemoteWake & NormallyConnectable

	}); err != nil {
		return err
	}

	// Register Control Point characteristic
	if err := registerStaticCharacteristic(svc, charHIDControlPoint, []string{
		gatt.FlagCharacteristicWriteWithoutResponse,
	}, []byte{0x00}); err != nil {
		return err
	}

	// Register HID Report Map characteristic
	// https://github.com/HeadHodge/Bluez-HID-over-GATT-Keyboard-Emulator-Example/blob/c19a079eb3f1d690d3319f52fff1c13dfa87a977/gattServer.py#L601
	hidReportMap, err := hex.DecodeString("05010906a1018501050719e029e71500250175019508810295017508150025650507190029658100c0050C0901A101850275109501150126ff0719012Aff078100C0")
	if err != nil {
		return err
	}
	if err := registerStaticCharacteristic(svc, charHIDReportMap, []string{
		gatt.FlagCharacteristicRead,
	}, hidReportMap); err != nil {
		return err
	}

	// Register HID Report
	if err := registerHIDReport(svc); err != nil {
		return err
	}

	log.Infof("Exposed HID service %s", svc.Properties.UUID)

	return nil
}

func registerDeviceInformation(a *service.App) error {
	svc, err := a.NewService(uuidDeviceInfo)
	if err != nil {
		return err
	}

	if err := a.AddService(svc); err != nil {
		return err
	}

	if err := registerStaticCharacteristic(svc, charDeviceInfoProduct, []string{
		gatt.FlagCharacteristicRead,
	}, []byte("Congruity")); err != nil {
		return err
	}

	if err := registerStaticCharacteristic(svc, charDeviceInfoVendor, []string{
		gatt.FlagCharacteristicRead,
	}, []byte("Congruity")); err != nil {
		return err
	}

	if err := registerStaticCharacteristic(svc, charDeviceInfoVersion, []string{
		gatt.FlagCharacteristicRead,
	}, []byte("Mrow")); err != nil {
		return err
	}

	if err := registerStaticCharacteristic(svc, charDeviceInfoPNPID, []string{
		gatt.FlagCharacteristicRead,
	}, []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
		return err
	}

	log.Infof("Exposed Device Info service %s", svc.Properties.UUID)

	return nil
}

func main() {
	if err := hidService(*adapter); err != nil {
		log.Fatal(err)
	}
}
