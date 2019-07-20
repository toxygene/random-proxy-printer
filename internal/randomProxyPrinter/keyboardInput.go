package randomProxyPrinter

import (
	"context"
	"github.com/kenshaw/evdev"
)

type KeyboardInput struct {
	device *evdev.Evdev
}

func NewKeyboardInput(devicePath string) (*KeyboardInput, error) {

	d, err := evdev.OpenFile(devicePath)
	if err != nil {
		return nil, err
	}

	return &KeyboardInput{
		device: d,
	}, nil
}

func (t *KeyboardInput) Listen(ctx context.Context, clockwiseListener chan<- interface {}, counterClockwiseListener chan<- interface {}, pushListener chan<- interface {}) error {
	err := t.device.Lock()
	if err != nil {
		return err
	}

	defer func() {
		err := t.device.Unlock()
		if err != nil {
			panic(err)
		}
	}()

	eventEnvelopChannel := t.device.Poll(ctx)

	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case eventEnvelop := <- eventEnvelopChannel:
			if eventEnvelop.Type == evdev.KeyEnter && eventEnvelop.Event.Value == 1 {
				pushListener <- struct {} {}
			} else if eventEnvelop.Type == evdev.KeyUp && eventEnvelop.Event.Value == 1 {
				clockwiseListener <- struct {} {}
			} else if eventEnvelop.Type == evdev.KeyDown && eventEnvelop.Event.Value == 1 {
				counterClockwiseListener <- struct {} {}
			} else if eventEnvelop.Type == evdev.KeyEscape && eventEnvelop.Event.Value == 1 {
				return nil // todo remove
			}
		}
	}
}