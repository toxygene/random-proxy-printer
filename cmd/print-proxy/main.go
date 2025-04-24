package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/kenshaw/escpos"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
)

func main() {
	cardName := flag.String("name", "", "Card name")
	printerDevicePath := flag.String("printer", "", "Path to the printer device")
	sqlitePath := flag.String("proxies", "", "path to the SQLite proxies database")
	help := flag.Bool("help", false, "print help page")

	flag.Parse()

	if *help || *cardName == "" || *sqlitePath == "" || *printerDevicePath == "" {
		flag.Usage()
		os.Exit(0)
	}

	db, err := sql.Open("sqlite3", *sqlitePath)
	if err != nil {
		println(fmt.Errorf("open sqlite3 database: %w", err).Error())
		os.Exit(1)
	}
	defer db.Close()

	f, err := os.OpenFile(*printerDevicePath, os.O_RDWR, 0)
	if err != nil {
		println(fmt.Errorf("open printer device: %w", err))
		os.Exit(1)
	}
	defer f.Close()

	logger := logrus.New()

	proxyPrinter := randomProxyPrinter.NewESCPOSPrinter(escpos.New(f), logrus.NewEntry(logger))

	proxy := randomProxyPrinter.Proxy{}

	row := db.QueryRow("SELECT name, print_data FROM proxies WHERE name=?", *cardName)

	if err := row.Scan(&proxy.Name, &proxy.PrintData); err != nil {
		println(fmt.Errorf("scan select proxy row: %w", err).Error())
		os.Exit(1)
	}

	if err := proxyPrinter.Print(proxy); err != nil {
		println(fmt.Errorf("print proxy: %w", err).Error())
		os.Exit(1)
	}
}
