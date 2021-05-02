package main

import (
	"flag"
	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/agent"
	log "github.com/sirupsen/logrus"
)

var (
	adapter = flag.String("adapter", "hci0", "Bluetooth adapter")
	btName  = flag.String("btName", "", "Set Bluetooth name (optional)")
)

func init() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)
}

func hidService(adapterID string) error {
	options := service.AppOptions{
		AdapterID: adapterID,
		AgentCaps: agent.CapNoInputNoOutput,

		// Bluetooth HID: 00001124-0000-1000-8000-00805f9b34fb
		// https://github.com/wolfeidau/bluez/blob/e7ea01cf88e010f4c64e70d59fe97c18b5203555/lib/uuid.h#L90
		UUID:       "0000",
		UUIDSuffix: "-0000-1000-8000-00805f9b34fb",
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

	hidService, err := a.NewService("1124") // BT HID (see above)
	if err != nil {
		return err
	}

	if err := a.AddService(hidService); err != nil {
		return err
	}

	if err := a.Run(); err != nil {
		return err
	}

	log.Infof("Exposed service %s", hidService.Properties.UUID)

	select {}
}

func main() {
	if err := hidService(*adapter); err != nil {
		log.Fatal(err)
	}
}
