package randomProxyPrinter

import (
	device "github.com/toxygene/i2c-ht16k33"
)

var segments_map = map[int][7]bool{
	0: {true, true, true, true, true, true, false},
	1: {false, true, true, false, false, false, false},
	2: {true, true, false, true, true, false, true},
	3: {true, true, true, true, false, false, true},
	4: {false, true, true, false, false, true, true},
	5: {true, false, true, true, false, true, true},
	6: {true, false, true, true, true, true, true},
	7: {true, true, true, false, false, false, false},
	8: {true, true, true, true, true, true, true},
	9: {true, true, true, true, false, true, true},
}

var reverse_segments_map = map[int][7]bool{
	0: {true, true, true, true, true, true, false},
	1: {false, false, false, false, true, true, false},
	2: {true, true, false, true, true, false, true},
	3: {true, false, false, true, true, true, true},
	4: {false, false, true, false, true, true, true},
	5: {true, false, true, true, false, true, true},
	6: {true, true, true, true, false, true, true},
	7: {false, false, false, true, true, true, false},
	8: {true, true, true, true, true, true, true},
	9: {true, false, true, true, true, true, true},
}

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
	tens := number / 10
	ones := number % 10

	if tens == 0 {
		t.ht16k33.SetSegments(1, [7]bool{})
		t.ht16k33.SetSegments(2, [7]bool{})
	} else {
		t.ht16k33.SetSegments(1, segments_map[tens])
		t.ht16k33.SetSegments(2, reverse_segments_map[tens])
	}

	t.ht16k33.SetSegments(0, segments_map[ones])
	t.ht16k33.SetSegments(3, reverse_segments_map[ones])

	t.ht16k33.WriteData()

	return nil
}
