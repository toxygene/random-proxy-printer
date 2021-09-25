package randomProxyPrinter

import (
	"bytes"
	"image"
	_ "image/png"
	"strings"

	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/raster"
	"github.com/mitchellh/go-wordwrap"
	"github.com/pkg/errors"
)

type ESCPOSPrinter struct {
	escpos *escpos.Printer
}

func NewESCPOSPrinter(escpos *escpos.Printer) *ESCPOSPrinter {
	return &ESCPOSPrinter{escpos: escpos}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	r := bytes.NewReader(proxy.Illustration)
	img, _, err := image.Decode(r)

	if err != nil {
		return errors.Wrap(err, "could not decode the proxy illustration")
	}

	rasterConv := &raster.Converter{
		MaxWidth:  384,
		Threshold: 0.5,
	}

	rasterConv.Print(img, t.escpos)

	for _, line := range strings.Split(proxy.Description, "\n") {
		for _, wrappedLine := range strings.Split(wordwrap.WrapString(line, 32), "\n") {
			t.escpos.Text(map[string]string{}, wrappedLine)
		}
	}

	t.escpos.Feed(map[string]string{})
	t.escpos.Feed(map[string]string{})
	t.escpos.Feed(map[string]string{})

	return nil
}
