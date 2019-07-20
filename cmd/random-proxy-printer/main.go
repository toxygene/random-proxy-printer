package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/run"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
	"os"
)

func main() {
	sqlitePathPtr := flag.String("proxies", "", "path to the SQLite proxies database")
	keyboardPath := flag.String("keyboard", "/dev/input/event6", "path to the keyboard input")

	flag.Parse()

	// todo validate flags

	db, err := sql.Open("sqlite3", *sqlitePathPtr)
	if err != nil {
		panic(err)
	}

	clockwiseChannel := make(chan interface{})
	counterClockwiseChannel := make(chan interface{})
	pushChannel := make(chan interface{})
	printChannel := make(chan randomProxyPrinter.Proxy)
	displayChannel := make(chan int)

	re, err := randomProxyPrinter.NewKeyboardInput(*keyboardPath)

	if err != nil {
		panic(err)
	}

	printer := randomProxyPrinter.StdoutPrinter{}
	display := randomProxyPrinter.StdoutDisplay{}

	p := randomProxyPrinter.NewRandomProxyPrinter(db,
		clockwiseChannel,
		counterClockwiseChannel,
		pushChannel,
		displayChannel,
		printChannel)

	ctx, cancel := context.WithCancel(context.Background())

	g := run.Group{}

	g.Add(func() error {
		return re.Listen(ctx, clockwiseChannel, counterClockwiseChannel, pushChannel)
	}, func(err error) {
		cancel()
	})

	g.Add(func() error {
		return printer.Listen(ctx, printChannel)
	}, func(err error) {
		cancel()
	})

	g.Add(func() error {
		return display.Listen(ctx, displayChannel)
	}, func(err error) {
		cancel()
	})

	g.Add(func() error {
		return p.Run(ctx)
	}, func(err error) {
		cancel()
	})

	err = g.Run()
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}
