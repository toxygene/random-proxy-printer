package randomProxyPrinter

import "context"

type Inputter interface {
	Run(context.Context, chan<- Action) error
}
