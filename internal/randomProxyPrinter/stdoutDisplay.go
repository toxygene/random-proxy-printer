package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type StdoutDisplay struct {
	Logger *log.Logger
}

func (t *StdoutDisplay) Listen(ctx context.Context, displayChannel chan int) error {
	for {
		select {
		case <-ctx.Done():
			t.Logger.
				Trace("context cancelled, shutting down stdout display")

			return ctx.Err()
		case value, ok := <-displayChannel:
			if !ok {
				return errors.New("display channel unexpectedly closed")
			}

			spew.Dump(value)
		}
	}
}
