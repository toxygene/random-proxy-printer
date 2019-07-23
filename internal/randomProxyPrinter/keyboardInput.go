package randomProxyPrinter

import (
	"context"
	"github.com/kenshaw/evdev"
	log "github.com/sirupsen/logrus"
)

type KeyboardInput struct {
	device   *evdev.Evdev
	logEntry *log.Entry
}

func NewKeyboardInput(l *log.Logger, devicePath string) (*KeyboardInput, error) {
	d, err := evdev.OpenFile(devicePath)
	if err != nil {
		return nil, err
	}

	return &KeyboardInput{
		device:   d,
		logEntry: l.WithField("keyboard device path", devicePath),
	}, nil
}

func (t *KeyboardInput) Listen(ctx context.Context, incrementValueChannel chan<- interface{}, decrementValueChannel chan<- interface{}, printCardChannel chan<- interface{}) error {
	t.logEntry.
		Trace("locking keyboard device")

	err := t.device.Lock()
	if err != nil {
		t.logEntry.
			WithError(err).
			Error("failed to lock keyboard device")

		return err
	}

	defer func() {
		t.logEntry.
			Trace("unlocking keyboard device")

		err := t.device.Unlock()
		if err != nil {
			t.logEntry.
				WithError(err).
				Error("failed to unlock keyboard device")

			panic(err)
		}
	}()

	t.logEntry.
		Trace("polling keyboard for events")

	eventEnvelopChannel := t.device.Poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case eventEnvelop := <-eventEnvelopChannel:
			if eventEnvelop.Type == evdev.KeyEnter && eventEnvelop.Event.Value == 1 {
				t.logEntry.
					Trace("enter key pressed")

				printCardChannel <- struct{}{}
			} else if eventEnvelop.Type == evdev.KeyUp && eventEnvelop.Event.Value == 1 {
				t.logEntry.
					Trace("up key pressed")

				incrementValueChannel <- struct{}{}
			} else if eventEnvelop.Type == evdev.KeyDown && eventEnvelop.Event.Value == 1 {
				t.logEntry.
					Trace("down key pressed")

				decrementValueChannel <- struct{}{}
			} else if eventEnvelop.Type == evdev.KeyEscape && eventEnvelop.Event.Value == 1 {
				t.logEntry.
					Trace("escape key pressed")

				return nil
			}
		}
	}
}
