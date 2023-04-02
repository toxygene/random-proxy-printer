module github.com/toxygene/random-proxy-printer

go 1.20

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/kenshaw/escpos v0.0.0-20201012084129-81d0344e35fa
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/rafalop/sevensegment v0.0.0-20220501111324-57abbea36ab7
	github.com/sirupsen/logrus v1.9.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/toxygene/gpiod-button v1.0.5
	github.com/toxygene/gpiod-ky-040-rotary-encoder v1.0.4
	github.com/toxygene/i2c-ht16k33 v1.0.0
	golang.org/x/sync v0.1.0
	periph.io/x/periph v3.6.8+incompatible
)

require github.com/davecheney/i2c v0.0.0-20140823063045-caf08501bef2 // indirect

require (
	github.com/warthog618/gpiod v0.8.1
	golang.org/x/sys v0.3.0 // indirect
)
