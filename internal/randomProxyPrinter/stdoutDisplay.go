package randomProxyPrinter

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

type StdoutDisplay struct {
	Logger *logrus.Logger
}

func (t *StdoutDisplay) Display(number int) error {
	spew.Dump(number)
	return nil
}
