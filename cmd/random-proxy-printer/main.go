package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/davecheney/i2c"
	"github.com/kenshaw/escpos"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	buttonDevice "github.com/toxygene/gpiod-button/device"
	rotaryEncoderDevice "github.com/toxygene/gpiod-ky-040-rotary-encoder/device"
	ht16k33Device "github.com/toxygene/i2c-ht16k33"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"github.com/warthog618/gpiod"
	"golang.org/x/sync/errgroup"
)

func main() {
	buttonPinNumber := flag.Int("button", 0, "GPIO number of button for the rotary encoder")
	chipName := flag.String("chipName", "", "Chip name for the GPIO device of the rotary encoder and button")
	help := flag.Bool("help", false, "print help page")
	ht16k33Bus := flag.Int("ht16k33Bus", 0, "I2C bus number the HT16K33 is attached to")
	ht16k33HexAddress := flag.String("ht16k33HexAddress", "", "Address, in hex, of the HT16K33 on the I2C bus")
	logging := flag.String("logging", "", "logging level")
	printerBaud := flag.Int("printerBaud", 19200, "Printer baud rate")
	printerDevicePath := flag.String("printer", "", "Path to the printer device")
	rotaryEncoderClockPinNumber := flag.Int("rotaryEncoderClock", 0, "GPIO number of clock pin for the rotary encoder")
	rotaryEncoderDataPinNumber := flag.Int("rotaryEncoderData", 0, "GPIO number of data pin for the rotary encoder")
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")

	flag.Parse()

	if *help || *buttonPinNumber == 0 || *rotaryEncoderClockPinNumber == 0 || *rotaryEncoderDataPinNumber == 0 || *sqlitePathPtr == "" || *ht16k33HexAddress == "" || *printerDevicePath == "" {
		flag.Usage()

		if *help {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	logger := logrus.New()

	if *logging != "" {
		logLevel, err := logrus.ParseLevel(*logging)
		if err != nil {
			println(fmt.Errorf("parse log level: %w", err).Error())
			os.Exit(1)
		}

		logger.SetLevel(logLevel)
	}

	db, err := sql.Open("sqlite3", *sqlitePathPtr)
	if err != nil {
		logger.WithField("sqlite", *sqlitePathPtr).WithError(err).Error("could not open sqlite3 database")
		println(fmt.Errorf("open sqlite database: %w", err).Error())
		os.Exit(1)
	}

	defer db.Close()

	chip, err := gpiod.NewChip(*chipName)
	if err != nil {
		logger.WithField("chipName", chipName).WithError(err).Error("could not create gpiod chip")
		println(fmt.Errorf("create gpiod chip: %w", err).Error())
		os.Exit(1)
	}

	defer chip.Close()

	serialConfig := &serial.Config{
		Name:   *printerDevicePath,
		Baud:   *printerBaud,
		Parity: serial.ParityNone,
	}

	printerSerialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		logger.WithError(err).WithField("serialConfig", serialConfig).Error("could not open serial port")
		println(fmt.Errorf("open serial port: %w", err).Error())
		os.Exit(1)
	}

	defer printerSerialPort.Close()

	ht16k33Address, err := strconv.ParseUint(*ht16k33HexAddress, 16, 8)
	if err != nil {
		logger.WithError(err).WithField("ht16k33HexAddress", *ht16k33HexAddress).Error("could not decode HT16K33 hex address")
		println(fmt.Errorf("decoding ht16k33 hex address: %w", err).Error())
		os.Exit(1)
	}

	i2c, err := i2c.New(uint8(ht16k33Address), *ht16k33Bus)
	if err != nil {
		logger.WithError(err).WithField("ht16k33Address", uint8(ht16k33Address)).Error("could not create I2C device")
		println(fmt.Errorf("create i2c device: %w", err).Error())
		os.Exit(1)
	}

	escPos := escpos.New(printerSerialPort)
	escPos.Init()

	p := randomProxyPrinter.NewRandomProxyPrinter(
		db,
		randomProxyPrinter.NewHT16K33Display(ht16k33Device.NewI2cHt16k33(i2c)),
		randomProxyPrinter.NewGpioInput(
			buttonDevice.NewButton(chip, *buttonPinNumber, logrus.NewEntry(logger)),
			rotaryEncoderDevice.NewRotaryEncoder(chip, *rotaryEncoderClockPinNumber, *rotaryEncoderDataPinNumber, logrus.NewEntry(logger)),
			logrus.NewEntry(logger),
		),
		randomProxyPrinter.NewESCPOSPrinter(escPos, logrus.NewEntry(logger)),
		logrus.NewEntry(logger),
	)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		logger.Info("random proxy printer goroutine started")
		defer logger.Info("random proxy printer goroutine finished")

		if err := p.Run(ctx); err != nil {
			return fmt.Errorf("run random proxy printer: %w", err)
		}

		return nil
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	g.Go(func() error {
		defer close(c)

		logger.Info("interrupt handler started")
		defer logger.Info("interrupt handler finished")

		select {
		case <-c:
			logger.Trace("sigint received")

			return errors.New("application interrupted")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		println(fmt.Errorf("run proxy printer: %w", err).Error())
		os.Exit(1)
	}

	os.Exit(0)
}
