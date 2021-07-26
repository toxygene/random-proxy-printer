package randomProxyPrinter

import (
	"context"
	"github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"golang.org/x/sync/errgroup"
)

type RotaryEncoderInput struct {
	RotaryEncoder *device.RotaryEncoder
}

func (t *RotaryEncoderInput) Run(parentCtx context.Context, actions chan<- Action) error {
	ctx, cancel := context.WithCancel(parentCtx)

	g := new(errgroup.Group)

	rotaryEncoderActions := make(chan device.Action)

	g.Go(func() error {
		defer close(rotaryEncoderActions)

		if err := t.RotaryEncoder.Run(ctx, rotaryEncoderActions); err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		for {
			rotaryEncoderAction, ok := <-rotaryEncoderActions

			if !ok {
				cancel()
				return nil
			}

			if rotaryEncoderAction == device.CW {
				actions <- IncrementValue
			} else if rotaryEncoderAction == device.CCW {
				actions <- DecrementValue
			} else if rotaryEncoderAction == device.Press {
				actions <- PrintRandomProxy
			}
		}
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
