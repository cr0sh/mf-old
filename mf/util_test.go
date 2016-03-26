package mf

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var nbTestEntries = []struct {
	toWrite, expect []byte
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

var nboTestEntries = []struct {
	toWrite, expect []byte
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
	{
		toWrite: []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1},
		expect:  []byte{0x90, 0x00, 0x00, 0x00, 0x90},
	},
	{
		toWrite: []byte{0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1},
		expect:  []byte{0x29, 0x00, 0x00, 0x00, 0x09},
	},
	{
		toWrite: []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x2},
		expect:  []byte{0x90, 0x00, 0x00, 0x00, 0x92},
	},
	{
		toWrite: []byte{0x2, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x2},
		expect:  []byte{0x29, 0x00, 0x00, 0x00, 0x09, 0x20},
	},
	{
		toWrite: []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
			0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2},
		expect: []byte{0x90, 0x00, 0x00, 0x00, 0x92, 0x22, 0x22, 0x22, 0x20},
	},
	{
		toWrite: []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
			0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2},
		expect: []byte{0x90, 0x00, 0x00, 0x00, 0x9a, 0x00, 0x00, 0x00, 0x0a},
	},
}

func TestNibblesOptimized(t *testing.T) {
	for n, test := range nboTestEntries {
		nw := new(NibbleWriterOptimized)
		nw.NibbleWriter = new(NibbleWriter)
		for _, b := range test.toWrite {
			nw.Put(b)
		}
		nw.Flush()
		if !bytes.Equal(test.expect, nw.Nibbles) {
			t.Errorf("Test #%d failed:\ngot      %s\nexpected %s", n+1, hex.EncodeToString(nw.Nibbles), hex.EncodeToString(test.expect))
			return
		}
	}
}

var u32TestEntries = []struct {
	bs  []byte
	u32 uint32
}{
	{
		bs:  []byte{0xff, 0xff, 0xff, 0xff},
		u32: 1<<32 - 1,
	},
	{
		bs:  []byte{0xde, 0xad, 0xbe, 0xef},
		u32: 3735928559,
	},
	{
		bs:  []byte{0, 0, 0, 0},
		u32: 0,
	},
}

func TestU32(t *testing.T) {
	for n, test := range u32TestEntries {
		if !bytes.Equal(test.bs, U32Bytes(test.u32)) {
			t.Errorf("Test #%d failed: expected %v, got %v", n+1, test.bs, U32Bytes(test.u32))
			return
		}
		if test.u32 != BytesU32(test.bs) {
			t.Errorf("Test #%d failed: expected %d, got %d", n+1, test.u32, BytesU32(test.bs))
			return
		}
	}
}

var nuTestEntries = []struct {
	n uint32
	b []byte
}{
	{
		3,
		[]byte{0, 0, 0, 0, 0, 0, 0, 3},
	},
	{
		52,
		[]byte{0, 0, 0, 0, 0, 0, 3, 4},
	},
	{
		256,
		[]byte{0, 0, 0, 0, 0, 1, 0, 0},
	},
	{
		^uint32(0),
		[]byte{15, 15, 15, 15, 15, 15, 15, 15},
	},
}

func TestNibblesU32(t *testing.T) {
	for n, test := range nuTestEntries {
		if test.n != NibblesU32(test.b) {
			t.Errorf("Test #%d failed: expected %d, got %d", n+1, test.n, NibblesU32(test.b))
			return
		}
	}
}

func TestU32Nibbles(t *testing.T) {
	for n, test := range nuTestEntries {
		if !bytes.Equal(test.b, U32Nibbles(test.n)) {
			t.Errorf("Test #%d failed: expected %d, got %d", n+1, test.b, U32Nibbles(test.n))
			return
		}
	}
}
