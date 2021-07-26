package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
	"time"
)

func main() {
	rotaryEncoderButtonPin := flag.String("button", "", "GPIO name of button for the rotary encoder")
	help := flag.Bool("help", false, "print help page")
	rotaryEncoderAPin := flag.String("pinA", "", "GPIO name of pin A for the rotary encoder")
	rotaryEncoderBPin := flag.String("pinB", "", "GPIO name of pin B for the rotary encoder")
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	rotaryEncoderTimeout := flag.Int("timeout", 2, "timeout (in seconds) for reading a pin")
	verbosePtr := flag.Bool("verbose", false, "verbose output")

	flag.Parse()

	if *help || *rotaryEncoderButtonPin == "" || *rotaryEncoderAPin == "" || *rotaryEncoderBPin == "" || *sqlitePathPtr == "" {
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

	aPin := gpioreg.ByName(*rotaryEncoderAPin)
	if aPin == nil {
		logger.WithField("pin_a", *rotaryEncoderAPin).Error("no gpio pin found for pin a")
		os.Exit(1)
	}

	bPin := gpioreg.ByName(*rotaryEncoderBPin)
	if bPin == nil {
		logger.WithField("pin_b", *rotaryEncoderBPin).Error("no gpio pin found for pin b")
		os.Exit(1)
	}

	buttonPin := gpioreg.ByName(*rotaryEncoderButtonPin)
	if buttonPin == nil {
		logger.WithField("button", *rotaryEncoderButtonPin).Error("no gpio bin found for button")
		os.Exit(1)
	}

	rotaryEncoder := device.NewRotaryEncoder(aPin, bPin, buttonPin, (time.Duration(*rotaryEncoderTimeout))*time.Second, logrus.NewEntry(logger))

	input := &randomProxyPrinter.RotaryEncoderInput{
		RotaryEncoder: rotaryEncoder,
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
