package randomProxyPrinter

import (
	device "github.com/toxygene/i2c-ht16k33"
)

type HT16K33Display struct {
	ht16k33 *device.I2cHt16k33
}

func NewHT16K33Display(ht16k33 *device.I2cHt16k33) *HT16K33Display {
	ht16k33.Clear()
	ht16k33.OscillatorOn()
	ht16k33.DisplayOn()

	return &HT16K33Display{ht16k33: ht16k33}
}

func (t *HT16K33Display) Display(number int) error {
	t.ht16k33.SetNumber(number)
	t.ht16k33.WriteData()

	return nil
}
