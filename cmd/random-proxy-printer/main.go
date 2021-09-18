package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	buttonDevice "github.com/toxygene/periphio-gpio-button/device"
	rotaryEncoderDevice "github.com/toxygene/periphio-gpio-rotary-encoder/v2/device"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"golang.org/x/sync/errgroup"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

func main() {
	buttonPinName := flag.String("button", "", "GPIO name of button for the rotary encoder")
	help := flag.Bool("help", false, "print help page")
	ht16k33Bus := flag.String("ht16k33Bus", "", "Name of the I2C bus the HT16K33 is attached to")
	ht16k33Address := flag.Int("ht16k33Address", 0, "Address of the HT16K33 on the I2C bus")
	printerDevicePath := flag.String("printer", "", "Path to the printer device")
	rotaryEncoderAPinName := flag.String("pinA", "", "GPIO name of pin A for the rotary encoder")
	rotaryEncoderBPinName := flag.String("pinB", "", "GPIO name of pin B for the rotary encoder")
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	waitTimeout := flag.Int("timeout", 2, "timeout (in seconds) for reading a pin")
	logging := flag.String("logging", "", "logging level")

	flag.Parse()

	if *help || *buttonPinName == "" || *rotaryEncoderAPinName == "" || *rotaryEncoderBPinName == "" || *sqlitePathPtr == "" || *ht16k33Bus == "" || *ht16k33Address == 0 || *printerDevicePath == "" {
		flag.Usage()
		os.Exit(0)
	}

	logger := logrus.New()

	if *logging != "" {
		logLevel, err := logrus.ParseLevel(*logging)
		if err != nil {
			panic(err)
		}

		logger.SetLevel(logLevel)
	}

	db, err := sql.Open("sqlite3", *sqlitePathPtr)
	if err != nil {
		logger.WithField("sqlite", *sqlitePathPtr).WithError(err).Error("could not open sqlite3 database")
		os.Exit(1)
	}

	if _, err := host.Init(); err != nil {
		logger.WithError(err).Error("could not initialize the host")
		os.Exit(1)
	}

	bus, err := i2creg.Open(*ht16k33Bus)
	if err != nil {
		logger.WithField("bus", *ht16k33Bus).WithError(err).Error("could not find i2c bus")
		os.Exit(1)
	}

	ht16k33Dev := i2c.Dev{
		Bus:  bus,
		Addr: uint16(*ht16k33Address),
	}

	ht16k33, err := randomProxyPrinter.NewHT16K33Display(ht16k33Dev)
	if err != nil {
		logger.WithField("device", ht16k33Dev).WithError(err).Error("could not create ht16k33 display")
		os.Exit(1)
	}

	timeout := (time.Duration(*waitTimeout)) * time.Second

	buttonPin := gpioreg.ByName(*buttonPinName)
	if buttonPin == nil {
		logger.WithField("button", *buttonPinName).Error("no gpio bin found for button")
		os.Exit(1)
	}

	if err := buttonPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("button", buttonPin).WithError(err).Error("could not setup button for input")
		os.Exit(1)
	}

	aPin := gpioreg.ByName(*rotaryEncoderAPinName)
	if aPin == nil {
		logger.WithField("pin_a", *rotaryEncoderAPinName).Error("no gpio pin found for pin a")
		os.Exit(1)
	}

	if err := aPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("pin_a", aPin).WithError(err).Error("could not setup pin a for input")
		os.Exit(1)
	}

	bPin := gpioreg.ByName(*rotaryEncoderBPinName)
	if bPin == nil {
		logger.WithField("pin_b", *rotaryEncoderBPinName).Error("no gpio pin found for pin b")
		os.Exit(1)
	}

	if err := bPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("pin_b", bPin).WithError(err).Error("could not setup pin b for input")
		os.Exit(1)
	}

	escpos, err := os.OpenFile(*printerDevicePath, os.O_RDWR, 0)
	if err != nil {
		logger.WithField("escpos", *printerDevicePath).Error("could not open the printer device")
		os.Exit(1)
	}
	defer escpos.Close()

	p := randomProxyPrinter.NewRandomProxyPrinter(
		db,
		ht16k33,
		randomProxyPrinter.NewGpioInput(
			buttonDevice.NewButton(buttonPin, timeout),
			rotaryEncoderDevice.NewRotaryEncoder(aPin, bPin, timeout, logrus.NewEntry(logger)),
			logrus.NewEntry(logger),
		),
		randomProxyPrinter.NewESCPOSPrinter(escpos),
		logrus.NewEntry(logger),
	)

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

		os.Exit(1)
	}

	os.Exit(0)
}
