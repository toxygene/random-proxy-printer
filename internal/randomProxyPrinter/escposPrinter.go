package randomProxyPrinter

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/cloudinn/escpos"
	"github.com/mitchellh/go-wordwrap"
)

type ESCPOSPrinter struct {
	escpos *escpos.Printer
}

func NewESCPOSPrinter(escpos *escpos.Printer) *ESCPOSPrinter {
	return &ESCPOSPrinter{escpos: escpos}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	if _, err := t.escpos.Write(proxy.PrintData); err != nil {
		return fmt.Errorf("write print data: %w", err)
	}

	for _, line := range strings.Split(proxy.Description, "\n") {
		for _, wrappedLine := range strings.Split(wordwrap.WrapString(line, 32), "\n") {
			if err := t.escpos.Text(map[string]string{}, wrappedLine); err != nil {
				return fmt.Errorf("write text: %w", err)
			}
		}
	}

	_ = t.escpos.Feed(map[string]string{})
	_ = t.escpos.Feed(map[string]string{})
	_ = t.escpos.Feed(map[string]string{})

	return nil
}
