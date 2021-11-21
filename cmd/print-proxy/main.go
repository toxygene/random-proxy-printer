package main

import (
	"database/sql"
	"flag"
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

	if *help ||  *cardName == "" || *sqlitePath == "" || *printerDevicePath == "" || *printerBaud == 0 {
		flag.Usage()
		os.Exit(0)
	}

	db, err := sql.Open("sqlite3", *sqlitePath)
	if err != nil {
		println("could not open sqlite3 database")
		os.Exit(1)
	}

	serialConfig := &serial.Config{
		Name:   *printerDevicePath,
		Baud:   *printerBaud,
		Parity: serial.ParityNone,
	}

	printerSerialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		println("could not open serial port")
		os.Exit(1)
	}
	defer printerSerialPort.Close()

	escposPrinter := escpos.New(printerSerialPort)

	proxyPrinter := randomProxyPrinter.NewESCPOSPrinter(escposPrinter)

	proxy := randomProxyPrinter.Proxy{}

	row := db.QueryRow("SELECT name, description, print_data FROM proxies WHERE name=?", *cardName)

	if err := row.Scan(&proxy.Name, &proxy.Description, &proxy.PrintData); err != nil {
		println("failed to fetch random proxy from database")
		os.Exit(1)
	}

	if err := proxyPrinter.Print(proxy); err != nil {
		println("failed to print random proxy")
		os.Exit(1)
	}
}
