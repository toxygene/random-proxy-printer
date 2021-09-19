package randomProxyPrinter

import (
	"bytes"
	"image"
	_ "image/png"
	"io"
	"strings"

	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/raster"
	"github.com/mitchellh/go-wordwrap"
	"github.com/pkg/errors"
)

type ESCPOSPrinter struct {
	destination io.ReadWriter
}

func NewESCPOSPrinter(destination io.ReadWriter) *ESCPOSPrinter {
	return &ESCPOSPrinter{destination: destination}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	p := escpos.New(t.destination)
	p.Init()

	r := bytes.NewReader(proxy.Illustration)
	img, _, err := image.Decode(r)

	if err != nil {
		return errors.Wrap(err, "could not decode the proxy illustration")
	}

	rasterConv := &raster.Converter{
		MaxWidth:  384,
		Threshold: 0.5,
	}

	rasterConv.Print(img, p)

	for _, line := range strings.Split(proxy.Description, "\n") {
		for _, wrappedLine := range strings.Split(wordwrap.WrapString(line, 32), "\n") {
			p.Text(map[string]string{}, wrappedLine)
		}
	}

	p.Feed(map[string]string{})
	p.Feed(map[string]string{})
	p.Feed(map[string]string{})

	return nil
}
