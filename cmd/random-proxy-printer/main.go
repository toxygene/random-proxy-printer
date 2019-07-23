package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/run"
	log "github.com/sirupsen/logrus"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"os"
)

func main() {
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	keyboardPathPtr := flag.String("keyboard", "/dev/input/event6", "path to the keyboard input")
	verbosePtr := flag.Bool("verbose", false, "verbose output")

	flag.Parse()

	// todo validate flags

	logger := log.New()

	if *verbosePtr {
		logger.SetLevel(log.TraceLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	db, err := sql.Open("sqlite3", *sqlitePathPtr)
	if err != nil {
		panic(err)
	}

	incrementValueChannel := make(chan interface{})
	decrementValueChannel := make(chan interface{})
	printCardChannel := make(chan interface{})
	outputProxyChannel := make(chan randomProxyPrinter.Proxy)
	outputValueChannel := make(chan int)

	re, err := randomProxyPrinter.NewKeyboardInput(logger, *keyboardPathPtr)

	if err != nil {
		panic(err)
	}

	printer := randomProxyPrinter.StdoutPrinter{}
	display := randomProxyPrinter.StdoutDisplay{}

	p := randomProxyPrinter.NewRandomProxyPrinter(logger,
		db,
		incrementValueChannel,
		decrementValueChannel,
		printCardChannel,
		outputValueChannel,
		outputProxyChannel)

	ctx, cancel := context.WithCancel(context.Background())

	g := run.Group{}

	g.Add(func() error {
		return re.Listen(ctx, incrementValueChannel, decrementValueChannel, printCardChannel)
	}, func(err error) {
		if err != nil {
			cancel()
		}
	})

	g.Add(func() error {
		return printer.Listen(ctx, outputProxyChannel)
	}, func(err error) {
		if err != nil {
			cancel()
		}
	})

	g.Add(func() error {
		return display.Listen(ctx, outputValueChannel)
	}, func(err error) {
		if err != nil {
			cancel()
		}
	})

	g.Add(func() error {
		return p.Run(ctx)
	}, func(err error) {
		if err != nil {
			cancel()
		}
	})

	logger.Trace("starting run group")

	err = g.Run()
	if err != nil && err != randomProxyPrinter.StopRunningError {
		logger.WithError(err).
			Error("run log group failed")

		panic(err)
	}

	os.Exit(0)
}
