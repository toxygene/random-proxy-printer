package tests

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	buttonDevice "github.com/toxygene/periphio-gpio-button/device"
	rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"testing"
	"time"
)

func TestGpioInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		entry := logrus.NewEntry(logrus.New())

		buttonPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
		if err := buttonPin.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
			assert.Fail(t, "setup of button failed")
		}

		aPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
		if err := aPin.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
			assert.Fail(t, "setup of pin a failed")
		}

		bPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
		if err := bPin.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
			assert.Fail(t, "setup of pin b failed")
		}

		gpio_input := randomProxyPrinter.GpioInput{
			Button:        buttonDevice.NewButton(buttonPin, time.Millisecond),
			RotaryEncoder: rotaryEncoderDevice.NewRotaryEncoder(aPin, bPin, time.Millisecond, entry),
		}

		ctx, cancel := context.WithCancel(context.Background())

		actions := make(chan randomProxyPrinter.Action)
		defer close(actions)

		go func() {
			defer cancel()

			buttonPin.EdgesChan <- gpio.High

			assert.Equal(t, randomProxyPrinter.PrintRandomProxy, <-actions)
		}()

		err := gpio_input.Run(ctx, actions)

		assert.Errorf(t, err, "context cancellation")
	})
}
