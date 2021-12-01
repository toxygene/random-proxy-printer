package randomProxyPrinter

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/kenshaw/escpos"
	"github.com/mitchellh/go-wordwrap"
)

type ESCPOSPrinter struct {
	escpos *escpos.Escpos
}

func NewESCPOSPrinter(escpos *escpos.Escpos) *ESCPOSPrinter {
	return &ESCPOSPrinter{escpos: escpos}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	if _, err := t.escpos.WriteRaw(proxy.PrintData); err != nil {
		return fmt.Errorf("write print data: %w", err)
	}

	t.escpos.Feed(map[string]string{})

	for _, line := range strings.Split(proxy.Description, "\n") {
		for _, wrappedLine := range strings.Split(wordwrap.WrapString(line, 32), "\n") {
			t.escpos.Text(map[string]string{}, wrappedLine)
			t.escpos.Text(map[string]string{}, "\n")
		}
		t.escpos.Text(map[string]string{}, "\n")
	}

	t.escpos.Text(map[string]string{}, "\n\n")

	return nil
}
