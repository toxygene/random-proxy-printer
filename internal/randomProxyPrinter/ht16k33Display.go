package randomProxyPrinter

import (
	"fmt"
	"periph.io/x/periph/conn/i2c"
)

var numbers = map[int][]byte{
	0:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3f},
	1:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06},
	2:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5b},
	3:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4f},
	4:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x66},
	5:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6d},
	6:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7d},
	7:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07},
	8:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7f},
	9:  {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6f},
	10: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x3f},
	11: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x06},
	12: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x5b},
	13: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x4f},
	14: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x66},
	15: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x6d},
	16: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x7d},
}

type HT16K33Display struct {
	dev i2c.Dev
}

func NewHT16K33Display(dev i2c.Dev) (*HT16K33Display, error) {
	if _, err := dev.Write([]byte{0x21}); err != nil {
		return nil, fmt.Errorf("enable oscillator: %w", err)
	}

	if _, err := dev.Write([]byte{0xe2}); err != nil {
		return nil, fmt.Errorf("set full brightness: %w", err)
	}

	if _, err := dev.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
		return nil, fmt.Errorf("clear display: %w", err)
	}

	return &HT16K33Display{dev: dev}, nil
}

func (t *HT16K33Display) Display(number int) error {
	if _, err := t.dev.Write(numbers[number]); err != nil {
		return fmt.Errorf("writing to ht16k33: %w", err)
	}

	return nil
}
