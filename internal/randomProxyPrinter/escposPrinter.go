package randomProxyPrinter

import (
	"fmt"
	_ "image/png"

	"github.com/kenshaw/escpos"
	"github.com/sirupsen/logrus"
)

type ESCPOSPrinter struct {
	escpos *escpos.Escpos
	logger *logrus.Entry
}

func NewESCPOSPrinter(escpos *escpos.Escpos, logger *logrus.Entry) *ESCPOSPrinter {
	return &ESCPOSPrinter{escpos: escpos, logger: logger}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	t.logger.Info("print starting")
	defer t.logger.Info("print finished")

	if _, err := t.escpos.WriteRaw(proxy.PrintData); err != nil {
		return fmt.Errorf("write print data: %w", err)
	}

	t.escpos.Text(map[string]string{}, "\n\n\n")
	t.escpos.WriteRaw([]byte{0x1B, 0x6D})

	return nil
}
