package randomProxyPrinter

import (
	"fmt"
	_ "image/png"
	"strings"
	"sync"

	"github.com/kenshaw/escpos"
	"github.com/mitchellh/go-wordwrap"
)

type ESCPOSPrinter struct {
	escpos *escpos.Escpos
	mu     sync.Mutex
}

func NewESCPOSPrinter(escpos *escpos.Escpos) *ESCPOSPrinter {
	return &ESCPOSPrinter{escpos: escpos}
}

func (t *ESCPOSPrinter) Print(proxy Proxy) error {
	if _, err := t.escpos.WriteRaw(proxy.PrintData); err != nil {
		return fmt.Errorf("write print data: %w", err)
	}

	for _, line := range strings.Split(proxy.Description, "\n") {
		for _, wrappedLine := range strings.Split(wordwrap.WrapString(line, 32), "\n") {
			t.escpos.Text(map[string]string{}, wrappedLine)
			t.escpos.Feed(map[string]string{})
		}
		t.escpos.Feed(map[string]string{})
	}

	t.escpos.Feed(map[string]string{"line": "2"})

	return nil
}
