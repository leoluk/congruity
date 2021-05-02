package libinput

import (
	"golang.org/x/sys/unix"
	"testing"
)

const (
	testDevice = "/dev/input/event9"
)

func TestNewPathContext(t *testing.T) {
	li := NewPathContext()
	defer li.Close()
}

func TestLibinputContext_PathAddDevice(t *testing.T) {
	li := NewPathContext()
	defer li.Close()

	if err := li.PathAddDevice(testDevice); err != nil {
		t.Error(err)
	}
}

func TestLibinputContext_Poll(t *testing.T) {
	li := NewPathContext()
	defer li.Close()

	if err := li.PathAddDevice(testDevice); err != nil {
		t.Fatal(err)
	}

	fd := li.GetPollFD()

	t.Log("please generate input")
	n, err := unix.Poll([]unix.PollFd{fd}, 10000)
	if err != nil {
		t.Fatal(err)
	}

	if err := li.Dispatch(); err != nil {
		t.Fatal(err)
	}

	if n <= 0 {
		t.Fatal("no events received")
	}

	ev, err := li.GetEvent()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("event: %+v", ev)
}
