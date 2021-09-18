package tests

import (
    "context"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    buttonDevice "github.com/toxygene/periphio-gpio-button/device"
    rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
    "github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
    "io/ioutil"
    "periph.io/x/periph/conn/gpio"
    "periph.io/x/periph/conn/gpio/gpiotest"
    "testing"
    "time"
)

func TestGpioInput(t *testing.T) {
    t.Run("clockwise 1", func(t *testing.T) {
        logger := logrus.New()
        logger.SetOutput(ioutil.Discard)
        entry := logrus.NewEntry(logger)

        buttonPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err := buttonPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        aPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err = aPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        bPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err = bPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        gpio_input := randomProxyPrinter.NewGpioInput(
            buttonDevice.NewButton(buttonPin, time.Millisecond),
            rotaryEncoderDevice.NewRotaryEncoder(aPin, bPin, time.Millisecond, entry),
            entry,
        )

        ctx, cancel := context.WithCancel(context.Background())

        actions := make(chan randomProxyPrinter.Action)
        defer close(actions)

        go func() {
            defer cancel()

            aPin.EdgesChan <- gpio.High
            time.Sleep(time.Millisecond)
            bPin.EdgesChan <- gpio.High
            time.Sleep(time.Millisecond)
            aPin.EdgesChan <- gpio.Low
            time.Sleep(time.Millisecond)
            bPin.EdgesChan <- gpio.Low

            assert.Equal(t, randomProxyPrinter.IncrementValue, <-actions)

            cancel()
        }()

        err = gpio_input.Run(ctx, actions)

        assert.Errorf(t, err, "context cancellation")
    })

    t.Run("button success", func(t *testing.T) {
        logger := logrus.New()
        logger.SetOutput(ioutil.Discard)
        entry := logrus.NewEntry(logger)

        buttonPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err := buttonPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        aPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err = aPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        bPin := &gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
        err = bPin.In(gpio.PullNoChange, gpio.BothEdges)
        assert.NoError(t, err)

        gpio_input := randomProxyPrinter.NewGpioInput(
            buttonDevice.NewButton(buttonPin, time.Millisecond),
            rotaryEncoderDevice.NewRotaryEncoder(aPin, bPin, time.Millisecond, entry),
            entry,
        )

        ctx, cancel := context.WithCancel(context.Background())

        actions := make(chan randomProxyPrinter.Action)
        defer close(actions)

        go func() {
            defer cancel()

            buttonPin.EdgesChan <- gpio.High

            assert.Equal(t, randomProxyPrinter.PrintRandomProxy, <-actions)
        }()

        err = gpio_input.Run(ctx, actions)

        assert.Errorf(t, err, "context cancellation")
    })
}
