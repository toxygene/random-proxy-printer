package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	buttonDevice "github.com/toxygene/periphio-gpio-button/device"
	rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
	"time"
)

func main() {
	buttonPinName := flag.String("button", "", "GPIO name of button for the rotary encoder")
	help := flag.Bool("help", false, "print help page")
	rotaryEncoderAPinName := flag.String("pinA", "", "GPIO name of pin A for the rotary encoder")
	rotaryEncoderBPinName := flag.String("pinB", "", "GPIO name of pin B for the rotary encoder")
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	rotaryEncoderTimeout := flag.Int("timeout", 2, "timeout (in seconds) for reading a pin")
	verbosePtr := flag.Bool("verbose", false, "verbose output")

	flag.Parse()

	if *help || *buttonPinName == "" || *rotaryEncoderAPinName == "" || *rotaryEncoderBPinName == "" || *sqlitePathPtr == "" {
		flag.Usage()
		os.Exit(0)
	}

	logger := logrus.New()

	if *verbosePtr {
		logger.SetLevel(logrus.TraceLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	db, err := sql.Open("sqlite3", *sqlitePathPtr)
	if err != nil {
		panic(err)
	}

	display := &randomProxyPrinter.StdoutDisplay{Logger: logger}

	if _, err := host.Init(); err != nil {
		panic(err)
	}

	aPin := gpioreg.ByName(*rotaryEncoderAPinName)
	if aPin == nil {
		logger.WithField("pin_a", *rotaryEncoderAPinName).Error("no gpio pin found for pin a")
		os.Exit(1)
	}

	if err := aPin.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		logger.WithField("pin_a", aPin).Error("could not setup pin a for input")
		os.Exit(1)
	}

	bPin := gpioreg.ByName(*rotaryEncoderBPinName)
	if bPin == nil {
		logger.WithField("pin_b", *rotaryEncoderBPinName).Error("no gpio pin found for pin b")
		os.Exit(1)
	}

	if err := bPin.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		logger.WithField("pin_b", bPin).Error("could not setup pin b for input")
		os.Exit(1)
	}

	rotaryEncoder := rotaryEncoderDevice.NewRotaryEncoder(aPin, bPin, (time.Duration(*rotaryEncoderTimeout))*time.Second, logrus.NewEntry(logger))

	buttonPin := gpioreg.ByName(*buttonPinName)
	if buttonPin == nil {
		logger.WithField("button", *buttonPinName).Error("no gpio bin found for button")
		os.Exit(1)
	}

	button := buttonDevice.NewButton(buttonPin, (time.Duration(*rotaryEncoderTimeout))*time.Second)

	input := &randomProxyPrinter.GpioInput{
		RotaryEncoder: rotaryEncoder,
		Button:        button,
	}

	printer := &randomProxyPrinter.StdoutPrinter{Logger: logger}

	p := randomProxyPrinter.NewRandomProxyPrinter(logger, db, display, input, printer)

	ctx, cancel := context.WithCancel(context.Background())

	g := new(errgroup.Group)

	g.Go(func() error {
		return p.Run(ctx)
	})

	g.Go(func() error {
		osSignalChannel := make(chan os.Signal, 1)
		signal.Notify(osSignalChannel, os.Interrupt)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-osSignalChannel:
				logger.Trace("SIGINT received")

				cancel()

				return nil
			}
		}
	})

	logger.Trace("starting run group")

	if err := g.Wait(); err != nil {
		logger.WithError(err).
			Error("run log group failed")

		panic(err)
	}

	os.Exit(0)
}
