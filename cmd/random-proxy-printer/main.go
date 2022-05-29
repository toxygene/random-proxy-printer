package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/kenshaw/escpos"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"github.com/toxygene/periphio-ky-040-rotary-encoder/device"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"golang.org/x/sync/errgroup"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {
	switchPinName := flag.String("switch", "", "GPIO name of switch for the rotary encoder")
	help := flag.Bool("help", false, "print help page")
	ht16k33Bus := flag.String("ht16k33Bus", "", "Name of the I2C bus the HT16K33 is attached to")
	ht16k33Address := flag.Int("ht16k33Address", 0, "Address of the HT16K33 on the I2C bus")
	printerBaud := flag.Int("printerBaud", 19200, "Printer baud rate")
	printerDevicePath := flag.String("printer", "", "Path to the printer device")
	rotaryEncoderClockPinName := flag.String("clock", "", "GPIO name of the clock pin for the rotary encoder")
	rotaryEncoderDataPinName := flag.String("data", "", "GPIO name of the data pin for the rotary encoder")
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	waitTimeout := flag.Int("timeout", 1, "timeout (in seconds) for reading a pin")
	logging := flag.String("logging", "", "logging level")

	flag.Parse()

	if *help || *switchPinName == "" || *rotaryEncoderClockPinName == "" || *rotaryEncoderDataPinName == "" || *sqlitePathPtr == "" || *ht16k33Bus == "" || *ht16k33Address == 0 || *printerDevicePath == "" {
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

	switchPin := gpioreg.ByName(*switchPinName)
	if switchPin == nil {
		logger.WithField("switch", *switchPinName).Error("no gpio pin found for switch")
		os.Exit(1)
	}

	if err := switchPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("switch", switchPin).WithError(err).Error("could not setup switch for input")
		os.Exit(1)
	}

	clockPin := gpioreg.ByName(*rotaryEncoderClockPinName)
	if clockPin == nil {
		logger.WithField("clock_pin", *rotaryEncoderClockPinName).Error("no gpio pin found for clock")
		os.Exit(1)
	}

	if err := clockPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("clock_pin", clockPin).WithError(err).Error("could not setup the clock pin for input")
		os.Exit(1)
	}

	dataPin := gpioreg.ByName(*rotaryEncoderDataPinName)
	if dataPin == nil {
		logger.WithField("data_pin", *rotaryEncoderDataPinName).Error("no gpio pin found for data")
		os.Exit(1)
	}

	if err := dataPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		logger.WithField("data_pin", dataPin).WithError(err).Error("could not setup data pin for input")
		os.Exit(1)
	}

	serialConfig := &serial.Config{
		Name:   *printerDevicePath,
		Baud:   *printerBaud,
		Parity: serial.ParityNone,
	}

	printerSerialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		logger.WithError(err).WithField("serial_config", serialConfig).Error("could not open serial port")
		os.Exit(1)
	}
	defer printerSerialPort.Close()

	escposPrinter := escpos.New(printerSerialPort)
	escposPrinter.Init()

	p := randomProxyPrinter.NewRandomProxyPrinter(
		db,
		ht16k33,
		randomProxyPrinter.NewKY040Inputter(
			device.NewRotaryEncoder(
				clockPin,
				dataPin,
				switchPin,
				timeout,
			),
		),
		randomProxyPrinter.NewESCPOSPrinter(escposPrinter),
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
