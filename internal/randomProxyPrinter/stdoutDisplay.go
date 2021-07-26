package randomProxyPrinter

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

type StdoutDisplay struct {
	Logger *logrus.Logger
}

func (t *StdoutDisplay) Display(proxy Proxy) error {
	spew.Dump(proxy)

	return nil
}
