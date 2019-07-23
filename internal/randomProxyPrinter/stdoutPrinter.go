package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

type StdoutPrinter struct {
}

func (t *StdoutPrinter) Listen(ctx context.Context, printChannel <-chan Proxy) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case card, ok := <-printChannel:
			if !ok {
				return errors.New("print channel unexpectedly closed")
			}

			spew.Dump(card.Description)
		}
	}
}
