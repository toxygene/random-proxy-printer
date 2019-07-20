package randomProxyPrinter

import (
	"context"
	"github.com/knq/escpos"
	"io"
)

type ESCPOSOutput struct {
	destination io.ReadWriter
}

func (t *ESCPOSOutput) Listen(ctx context.Context, outputChannel <-chan Proxy) error {
	p := escpos.New(t.destination)
	p.Init()

	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case proxy := <- outputChannel:
			//p.Image(map[string]string {}, proxy.Illustration) // todo string? I dunno
			p.Text(map[string]string {}, proxy.Description) // todo line wrapping
			p.Feed(map[string]string {})              // todo additional form feeds
		}
	}
}