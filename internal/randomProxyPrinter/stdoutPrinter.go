package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type StdoutPrinter struct {
	Logger *log.Logger
}

func (t *StdoutPrinter) Listen(ctx context.Context, printChannel <-chan Proxy) error {
	for {
		select {
		case <-ctx.Done():
			t.Logger.
				Trace("context cancelled, shutting down stdout printer")

			return ctx.Err()
		case card, ok := <-printChannel:
			if !ok {
				return errors.New("print channel unexpectedly closed")
			}

			spew.Dump(card.Description)
		}
	}
}
