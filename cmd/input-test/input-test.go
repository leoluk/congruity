package main

import (
	"flag"
	"github.com/leoluk/congruity/pkg/libinput"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"strings"
)

// sudo libinput list-devices
// sudo chown leopold:input /dev/input/event7

var (
	devices = flag.String("devices", "", "evdev devices (comma-separated)")
)

func init() {
	flag.Parse()

	if *devices == "" {
		log.Fatal("please specify devices")
	}
}

func main() {
	li := libinput.NewPathContext()
	defer li.Close()

	for _, dev := range strings.Split(*devices, ",") {
		if err := li.PathAddDevice(dev); err != nil {
			log.Fatalf("failed to add devices: %v", err)
		}
	}

	fd := li.GetPollFD()

	for {
		n, err := unix.Poll([]unix.PollFd{fd}, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			log.Fatalf("failed to poll libinput fd: %v", err)
		}

		if err := li.Dispatch(); err != nil {
			log.Fatal(err)
		}

		if n <= 0 {
			continue
		}

		ev, err := li.GetEvent()
		if err != nil {
			continue
		}

		log.Printf("event: %s", ev.Type)
	}
}
