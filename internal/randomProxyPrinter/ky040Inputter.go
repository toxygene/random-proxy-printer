package randomProxyPrinter

import (
	"context"

	"github.com/toxygene/periphio-ky-040-rotary-encoder/device"
	"golang.org/x/sync/errgroup"
)

type KY040Inputter struct {
	rotaryEncoder *device.RotaryEncoder
}

func NewKY040Inputter(rotaryEncoder *device.RotaryEncoder) *KY040Inputter {
	return &KY040Inputter{
		rotaryEncoder: rotaryEncoder,
	}
}

func (k *KY040Inputter) Run(ctx context.Context, actions chan<- Action) error {
	g := errgroup.Group{}

	deviceActions := make(chan device.Action)

	g.Go(func() error {
		err := k.rotaryEncoder.Run(ctx, deviceActions)
		close(deviceActions)
		return err
	})

	g.Go(func() error {
		for deviceAction := range deviceActions {
			if deviceAction == device.Clockwise {
				actions <- IncrementValue
			} else if deviceAction == device.CounterClockwise {
				actions <- DecrementValue
			} else if deviceAction == device.Click {
				actions <- PrintRandomProxy
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
