package randomProxyPrinter

import (
	"context"
	"github.com/davecgh/go-spew/spew"
)

type StdoutDisplay struct {

}

func (t *StdoutDisplay) Listen(ctx context.Context, displayChannel chan int) error {
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case value := <- displayChannel:
			spew.Dump(value)
		}
	}
}