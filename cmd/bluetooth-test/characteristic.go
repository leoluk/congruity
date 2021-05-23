package main

import (
	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	log "github.com/sirupsen/logrus"
	"time"
)

func registerStaticCharacteristic(svc *service.Service, uuid string, flags []string, payload []byte) error {
	char, err := svc.NewChar(uuid)
	if err != nil {
		return err
	}

	char.Properties.Flags = flags

	char.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Warnf("Replying to read request for %s: %02x", uuid, payload)
		return payload, nil
	}))

	char.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		log.Warnf("Ignoring write request for %s: %02x", uuid, value)
		return value, nil
	}))

	return nil
}

func registerHIDReport(svc *service.Service) error {
	char, err := svc.NewChar(charHIDReport)
	if err != nil {
		return err
	}

	char.Properties.Flags = []string{
		gatt.FlagCharacteristicSecureRead,
		gatt.FlagCharacteristicNotify,
	}

	char.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Warnf("Replying to read request for %s")
		return []byte{0x00, 0x00}, nil
	}))

	char.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		log.Warnf("Ignoring write request for %s: %02x", value)
		return value, nil
	}))

	descr, err := char.NewDescr("2908") // TODO missing constant
	if err != nil {
		return err
	}

	descr.Properties.Flags = []string{
		gatt.FlagCharacteristicRead, // TODO secure read?
	}

	descr.OnRead(service.DescrReadCallback(func(c *service.Descr, options map[string]interface{}) ([]byte, error) {
		log.Warnf("GOT READ REQUEST")
		return []byte{0x01, 0x01}, nil
	}))
	descr.OnWrite(service.DescrWriteCallback(func(d *service.Descr, value []byte) ([]byte, error) {
		log.Warnf("GOT WRITE REQUEST")
		return value, nil
	}))

	err = char.AddDescr(descr)
	if err != nil {
		return err
	}

	if err := char.StartNotify(); err != nil {
		return err
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}
