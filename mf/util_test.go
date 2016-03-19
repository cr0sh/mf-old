package mf

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var nbTestEntries = []struct {
	toWrite []byte
	expect  []byte
}{
	{
		toWrite: []byte{0x1},
		expect:  []byte{0x10},
	},
	{
		toWrite: []byte{0xf, 0xf, 0xf, 0xf},
		expect:  []byte{0xff, 0xff},
	},
	{
		toWrite: []byte{0xf, 0x1, 0x3, 0xf},
		expect:  []byte{0xf1, 0x3f},
	},
	{
		toWrite: []byte{0xf, 0x1, 0x2},
		expect:  []byte{0xf1, 0x20},
	},
}

func TestNibbles(t *testing.T) {
	for n, test := range nbTestEntries {
		nw := new(NibbleWriter)
		for _, b := range test.toWrite {
			nw.Put(b)
		}
		if !bytes.Equal(test.expect, nw.Nibbles) {
			t.Errorf("Test #%d failed:\ngot      %s\nexpected %s", n+1, hex.EncodeToString(nw.Nibbles), hex.EncodeToString(test.expect))
			return
		}
	}
}
