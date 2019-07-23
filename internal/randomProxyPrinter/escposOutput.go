package randomProxyPrinter

import (
	"bytes"
	"context"
	"github.com/knq/escpos"
	"github.com/knq/escpos/raster"
	"github.com/mitchellh/go-wordwrap"
	"github.com/pkg/errors"
	"image"
	"io"
	"strings"
)

type ESCPOSOutput struct {
	destination io.ReadWriter
}

func (t *ESCPOSOutput) Listen(ctx context.Context, outputChannel <-chan Proxy) error {
	p := escpos.New(t.destination)
	p.Init()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case proxy := <-outputChannel:
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
		}
	}
}
