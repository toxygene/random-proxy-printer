package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
)

type StdoutPrinter struct {

}

func (t *StdoutPrinter) Listen(ctx context.Context, printChannel <-chan Proxy) error {
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case card := <- printChannel:
			spew.Dump(card.Description)
		}
	}
}
