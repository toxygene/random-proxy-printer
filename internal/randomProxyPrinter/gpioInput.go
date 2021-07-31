package randomProxyPrinter

import (
	"context"
	"fmt"
	buttonDevice "github.com/toxygene/periphio-gpio-button/device"
	rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"golang.org/x/sync/errgroup"
)

type GpioInput struct {
	Button        *buttonDevice.Button
	RotaryEncoder *rotaryEncoderDevice.RotaryEncoder
}

func (t *GpioInput) Run(ctx context.Context, actions chan<- Action) error {
	g := new(errgroup.Group)

	rotaryEncoderActions := make(chan rotaryEncoderDevice.Action)

	g.Go(func() error {
		defer close(rotaryEncoderActions)

		if err := t.RotaryEncoder.Run(ctx, rotaryEncoderActions); err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		for rotaryEncoderAction := range rotaryEncoderActions {
			if rotaryEncoderAction == rotaryEncoderDevice.CW {
				actions <- IncrementValue
			} else if rotaryEncoderAction == rotaryEncoderDevice.CCW {
				actions <- DecrementValue
			}
		}

		return nil
	})

	buttonActions := make(chan buttonDevice.Action)

	g.Go(func() error {
		defer close(buttonActions)

		if err := t.Button.Run(ctx, buttonActions); err != nil {
			return fmt.Errorf("button run failed: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		for buttonAction := range buttonActions {
			if buttonAction == buttonDevice.Push {
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
