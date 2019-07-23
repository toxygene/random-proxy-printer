package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

type StdoutDisplay struct {
}

func (t *StdoutDisplay) Listen(ctx context.Context, displayChannel chan int) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value, ok := <-displayChannel:
			if !ok {
				return errors.New("display channel unexpectedly closed")
			}

			spew.Dump(value)
		}
	}
}
