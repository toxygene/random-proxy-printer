package randomProxyPrinter

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	buttonDevice "github.com/toxygene/gpiod-button/device"
	rotaryEncoderDevice "github.com/toxygene/gpiod-ky-040-rotary-encoder/device"
	"golang.org/x/sync/errgroup"
)

func NewGpioInput(button *buttonDevice.Button, rotaryEncoder *rotaryEncoderDevice.RotaryEncoder, logger *logrus.Entry) *GpioInput {
	return &GpioInput{
		button:         button,
		buttonDebounce: 250,
		logger:         logger,
		rotaryEncoder:  rotaryEncoder,
	}
}

type GpioInput struct {
	button         *buttonDevice.Button
	buttonDebounce int
	logger         *logrus.Entry
	rotaryEncoder  *rotaryEncoderDevice.RotaryEncoder
}

func (t *GpioInput) Run(parentCtx context.Context, actions chan<- Action) error {
	t.logger.Info("starting gpio input")
	defer t.logger.Info("gpio input finished")

	g, ctx := errgroup.WithContext(parentCtx)

	rotaryEncoderActions := make(chan rotaryEncoderDevice.Action)

	g.Go(func() error {
		defer close(rotaryEncoderActions)

		t.logger.Info("rotary encoder goroutine starting")
		defer t.logger.Info("rotary encoder goroutine finished")

		if err := t.rotaryEncoder.Run(ctx, rotaryEncoderActions); err != nil {
			return fmt.Errorf("run rotary encoder: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		t.logger.Info("rotary encoder action handler goroutine started")
		defer t.logger.Info("rotary encoder action handler goroutine finished")

		for rotaryEncoderAction := range rotaryEncoderActions {
			t.logger.WithField("rotary_encoder_action", rotaryEncoderAction).Trace("rotary encoder action")

			if rotaryEncoderAction == rotaryEncoderDevice.Clockwise {
				actions <- IncrementValue
			} else if rotaryEncoderAction == rotaryEncoderDevice.CounterClockwise {
				actions <- DecrementValue
			}
		}

		return nil
	})

	buttonActions := make(chan buttonDevice.Action)

	g.Go(func() error {
		defer close(buttonActions)

		t.logger.Info("button goroutine started")
		defer t.logger.Info("button goroutine finished")

		if err := t.button.Run(ctx, buttonActions); err != nil {
			return fmt.Errorf("button run: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		t.logger.Info("button action handler goroutine started")
		defer t.logger.Info("button action handler goroutine finished")

		for buttonAction := range buttonActions {
			t.logger.WithField("button_action", buttonAction).Trace("button action")

			if buttonAction == buttonDevice.Press {
				actions <- PrintRandomProxy

				cont := true
				debounce := time.NewTimer(time.Duration(t.buttonDebounce) * time.Millisecond)

				for cont {
					select {
					case <-buttonActions:
					case <-debounce.C:
						cont = false
					}
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run gpio input: %w", err)
	}

	return nil
}
