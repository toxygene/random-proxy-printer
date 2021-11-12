package randomProxyPrinter

import (
	"context"
	"github.com/sirupsen/logrus"
	buttonDevice "github.com/toxygene/periphio-gpio-button/device"
	rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"golang.org/x/sync/errgroup"
)

func NewGpioInput(button *buttonDevice.Button, rotaryEncoder *rotaryEncoderDevice.RotaryEncoder, logger *logrus.Entry) *GpioInput {
	return &GpioInput{
		button:        button,
		logger:        logger,
		rotaryEncoder: rotaryEncoder,
	}
}

type GpioInput struct {
	button        *buttonDevice.Button
	logger        *logrus.Entry
	rotaryEncoder *rotaryEncoderDevice.RotaryEncoder
}

func (t *GpioInput) Run(ctx context.Context, actions chan<- Action) error {
	g := new(errgroup.Group)

	rotaryEncoderActions := make(chan rotaryEncoderDevice.Action)

	g.Go(func() error {
		defer close(rotaryEncoderActions)

		t.logger.Trace("starting rotary encoder")

		if err := t.rotaryEncoder.Run(ctx, rotaryEncoderActions); err != nil {
			return err
		}

		t.logger.Trace("rotary encoder finished")

		return nil
	})

	g.Go(func() error {
		for rotaryEncoderAction := range rotaryEncoderActions {
			t.logger.WithField("rotary_encoder_action", rotaryEncoderAction).Trace("rotary encoder action")

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
		for buttonAction := range buttonActions {
			t.logger.WithField("button_action", buttonAction).Trace("button action")

			if buttonAction == buttonDevice.Push {
				actions <- PrintRandomProxy
			}
		}

		return nil
	})

	g.Go(func() error {
		defer close(buttonActions)

		t.logger.Trace("starting button")

		if err := t.button.Run(ctx, buttonActions); err != nil {
			return err
		}

		t.logger.Trace("button finished")

		return nil
	})

	t.logger.Trace("starting gpio input")

	if err := g.Wait(); err != nil {
		return err
	}

	t.logger.Trace("gpio input finished")

	return nil
}
