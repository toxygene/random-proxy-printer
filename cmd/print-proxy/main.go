package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/kenshaw/escpos"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tarm/serial"
	"github.com/toxygene/random-proxy-printer/internal/randomProxyPrinter"
)

func main() {
	cardName := flag.String("name", "", "Card name")
	printerBaud := flag.Int("printerBaud", 19200, "Printer baud rate")
	printerDevicePath := flag.String("printer", "", "Path to the printer device")
	sqlitePath := flag.String("proxies", "", "path to the SQLite proxies database")
	help := flag.Bool("help", false, "print help page")

	flag.Parse()

	if *help || *cardName == "" || *sqlitePath == "" || *printerDevicePath == "" || *printerBaud == 0 {
		flag.Usage()
		os.Exit(0)
	}

	db, err := sql.Open("sqlite3", *sqlitePath)
	if err != nil {
		println(fmt.Errorf("open sqlite3 database: %w", err).Error())
		os.Exit(1)
	}

	serialConfig := &serial.Config{
		Name:   *printerDevicePath,
		Baud:   *printerBaud,
		Parity: serial.ParityNone,
	}

	printerSerialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		println(fmt.Errorf("open serial port: %w", err).Error())
		os.Exit(1)
	}
	defer printerSerialPort.Close()

	proxyPrinter := randomProxyPrinter.NewESCPOSPrinter(escpos.New(printerSerialPort))

	proxy := randomProxyPrinter.Proxy{}

	row := db.QueryRow("SELECT name, description, print_data FROM proxies WHERE name=?", *cardName)

	if err := row.Scan(&proxy.Name, &proxy.Description, &proxy.PrintData); err != nil {
		println(fmt.Errorf("scan select proxy row: %w", err).Error())
		os.Exit(1)
	}

	if err := proxyPrinter.Print(proxy); err != nil {
		println(fmt.Errorf("print proxy: %w", err).Error())
		os.Exit(1)
	}
}
