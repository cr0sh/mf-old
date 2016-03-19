package mf

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var fpTestEntries = []struct {
	file    []byte
	err     bool
	memsize uint32
	code    []byte
}{
	{ // Test #1: empty file
		file: []byte{},
		err:  true,
	},
	{ // Test #2: truncated magic
		file: []byte{0xff},
		err:  true,
	},
	{ // Test #3: magic only
		file: []byte(mfMagic),
		err:  true,
	},
	{ // Test #4: missing code length
		file: append([]byte(mfMagic), []byte{0xff, 0x6d, 0x66, 0xfd, 0x00, 0x00, 0xff, 0xff}...),
		err:  true,
	},
	{ // Test #5: insufficient code length
		file: append([]byte(mfMagic), []byte{0xff, 0x6d, 0x66, 0xfd, 0x00, 0x00, 0xff, 0xff, 0x00}...),
		err:  true,
	},
	{ // Test #6: File with code length 0 (works)
		file:    append([]byte(mfMagic), []byte{0xff, 0x6d, 0x66, 0xfd, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}...),
		memsize: 65535,
		code:    []byte{},
	},
	{ // Test #7: Unexpected EOF while reading code
		file: append([]byte(mfMagic), []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x01, 0x02, 0x03, 0x04}...),
		err:  true,
	},
	{ // Test #8: Valid MinFuck file (works)
		file:    append([]byte(mfMagic), []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}...),
		memsize: 256,
		code:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	},
}

func TestFileParse(t *testing.T) {
	for i, test := range fpTestEntries {
		e := &IOStream{Stdin: string(test.file)}
		fd, err := ReadFile(e)
		if test.err {
			if err == nil {
				t.Errorf("Test #%d should return error but nothing happened: %v", i+1, test)
			}
			return
		}
		if err != nil {
			t.Errorf("Test #%d should not return error: %v", i+1, test)
			return
		}
		if fd.memsize != test.memsize {
			t.Errorf("Test #%d allocates %d memory alloc size: expected %d", i+1, fd.memsize, test.memsize)
			return
		}
		if !bytes.Equal(fd.code, test.code) {
			t.Errorf("Test #%d mismatches code: \ngot      %s\nexpected %s", i+1, hex.EncodeToString(fd.code), hex.EncodeToString(test.code))
		}
	}
}

var nTestEntries = []struct {
	code []byte
	read int
	gets []byte
	err  bool
}{
	{ // Test #1: Single nibble read
		code: []byte{0xff},
		read: 1,
		gets: []byte{0xf},
	},
	{ // Test #2: Multiple nibble read
		code: []byte{0xef, 0xa0},
		read: 3,
		gets: []byte{0xe, 0xf, 0xa},
	},
	{ // Test #3: Unexpected EOF
		code: []byte{0x0f},
		read: 3,
		err:  true,
	},
}

func TestNibble(t *testing.T) {
	for n, test := range nTestEntries {
		vm := &MinFuckVM{code: test.code}
		gets := make([]byte, test.read)
		for i := 0; i < test.read; i++ {
			var err error
			gets[i], err = vm.nibble()
			if err != nil {
				if !test.err {
					t.Errorf("Test #%d failed: VM should not return error while nibble(): %v", n+1, err)
				}
				return
			}
		}
		if test.err {
			t.Errorf("Test #%d failed: VM should return error but nothing happened", n+1)
			return
		}
		if !bytes.Equal(gets, test.gets) {
			t.Errorf("Test #%d failed: \ngot      %s\nexpected %s", n+1, hex.EncodeToString(gets), hex.EncodeToString(test.gets))
			return
		}
	}
}

var nnTestEntries = []struct {
	code []byte
	read []int
	gets [][]byte
	err  bool
}{
	{
		code: []byte{},
		read: []int{1},
		err:  true,
	},
	{
		code: []byte{0xff},
		read: []int{2},
		gets: [][]byte{{0xf, 0xf}},
	},
	{
		code: []byte{0xff, 0xa0, 0xf3, 0xaf},
		read: []int{3, 3, 2},
		gets: {[]byte{0xf, 0xf, 0xa}, []byte{0x0, 0xf, 0x3}, []byte{0xa, 0xf}},
	},
	{
		code: []byte{0xff, 0xa0, 0xf3, 0xaf},
		read: []int{8},
		gets: [][]byte{{0xf, 0xf, 0xa, 0x0, 0xf, 0x3, 0xa, 0xf}},
	},
	{
		code: []byte{0xf0},
		read: []int{3},
		err:  true,
	},
	{
		code: []byte{0xfd},
		read: []int{1, 1, 1},
		err:  true,
	},
}

func TestNibbleN(t *testing.T) {
	for n, test := range nnTestEntries {
		vm := &MinFuckVM{code: test.code}
		for i, r := range test.read {
			g, err := vm.nibbleN(uint32(r))
			if err != nil {
				if !test.err {
					t.Errorf("Test #%d failed: VM should not return error while nibbleN(): %v", n+1, err)
				}
				return
			}
			if !bytes.Equal(test.gets[i], g) {
				t.Errorf("Test #%d failed: result mismatch\ngot      %s\nexpected %s", n+1, hex.EncodeToString(g), hex.EncodeToString(test.gets[i]))
				return
			}
		}
	}
}
