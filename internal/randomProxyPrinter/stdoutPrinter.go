package randomProxyPrinter

import (
    "github.com/davecgh/go-spew/spew"
    "github.com/sirupsen/logrus"
)

type StdoutPrinter struct {
    Logger *logrus.Logger
}

func (t *StdoutPrinter) Print(proxy Proxy) error {
    spew.Dump(proxy)
    return nil
}
