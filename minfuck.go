package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

const version = "0.1"
const mainHelp = `MinFuck 버전 %s
도움말: %s [operation] option1, ...
help: 지금 보고 있는 도움말을 출력합니다.
b2m [filename]: 주어진 Brainfuck 코드를 MinFuck 코드로 변환합니다.
run [filename]: 주어진 MinFuck 코드를 구동합니다.
`

func main() {
	if len(os.Args) < 2 {
		help()
	}
	switch os.Args[1] {
	case "help":
		help()
	case "b2m":
		b2m()
	case "run":
		run()
	default:
		fmt.Println("정의되지 않은 동작:", os.Args[1])
		help()
	}
}

func b2m() {
	if len(os.Args) < 3 {
		fmt.Println("변환할 Brainfuck 소스 파일이 필요합니다.")
		help()
	}
	b, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("파일 여는 중 오류:", err)
		os.Exit(3)
	}
	ioutil.WriteFile(
		os.Args[2][0:len(os.Args[2])-len(path.Ext(os.Args[2]))]+".mf",
		[]byte(bfToMf(string(b), 4096)),
		0644)
}
func run() {
	if len(os.Args) < 3 {
		fmt.Println("실행할 MinFuck 코드가 필요합니다.")
		help()
	}
	f, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println("파일 여는 중 오류:", err)
		os.Exit(3)
	}
	vm, err := vmFile(f)
	if err != nil {
		fmt.Println("VM 준비 중 오류:", err)
		os.Exit(4)
	}
	err = vm.run()
	if err != nil {
		fmt.Printf("\n==================\n코드가 비정상 종료되었습니다: %s\n", err.Error())
		os.Exit(2)
	} else {
		fmt.Printf("\n==================\n코드가 정상적으로 종료되었습니다\n")
		os.Exit(0)
	}
}

func help() {
	fmt.Printf(mainHelp, version, os.Args[0])
	os.Exit(0)
}

// ======= MinFuck VM =======

const mfMagic = "\xff\x6d\x66\xfd"

/*
MinFuck binary 포맷

.mf 파일의 첫 4바이트는 Magic Byte(\xff\x6d\x66\xfd)입니다.
다음 4바이트에 부호 없는 정수형으로 MinFuck VM에서 접근 가능한 최대 메모리 번지를 지정합니다.
(단, 실제 OS에서는 최소 해당 값 * 8 + 24바이트 이상을 할당합니다.)
다음 4바이트에는 부호 없는 정수형으로 코드의 크기를 명시합니다.
*/
type fileData struct {
	memsize uint32
	code    []byte
}

func readFile(f *os.File) (fileData, error) {
	f.Seek(0, 0)
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return fileData{}, err
	}
	if !bytes.Equal(magic, []byte(mfMagic)) {
		return fileData{}, fmt.Errorf("잘못된 MinFuck Magic: 0x" + hex.EncodeToString(magic))
	}

	membuf := make([]byte, 4)
	if _, err := f.Read(membuf); err != nil {
		return fileData{}, err
	}
	memsize, n := binary.Uvarint(membuf)
	if n <= 0 {
		return fileData{}, fmt.Errorf("메모리 주소 제한 값이 잘못되었습니다")
	}

	codebuf := make([]byte, 4)
	if _, err := f.Read(codebuf); err != nil {
		return fileData{}, err
	}
	codesize, n := binary.Uvarint(codebuf)
	if n <= 0 {
		return fileData{}, fmt.Errorf("코드 길이 값이 잘못되었습니다")
	}

	code := make([]byte, codesize)
	if _, err := f.Read(code); err != nil {
		return fileData{}, err
	}

	return fileData{memsize: uint32(memsize), code: code}, nil
}

func (f *fileData) toString() string {
	buf := bytes.NewBuffer([]byte(mfMagic))
	buf.Write(u32Bytes(f.memsize))
	buf.Write(u32Bytes(uint32(len(f.code))))
	buf.Write(f.code)
	return buf.String()
}

/*
MinFuck 코드 포맷

MinFuck의 코드는 기본적으로 Brainfuck과 1:1로 변환이 가능합니다.
Brainfuck의 []+-<>., 코드를 크기를 줄이기 위해 nibble(1/2 byte) 사이즈로 줄이고,
단순 반복으로 인한 코드 크기 증가를 막기 위해 반복 압축을 추가했습니다.

각 니블코드의 첫 비트를 제외하고, 뒤 7비트는 다음과 같은 의미를 가집니다:
0: Brainfuck의 + (주의: 메모리 셀은 Brainfuck과 다르게 32비트 부호 없는 정수형입니다. 오버플로가 일어날 수 있습니다.)
1: Brainfuck의 -
2: Brainfuck의 > (주의: >< 니블코드는 0 아래로 떨어지지 않습니다)
3: Brainfuck의 <
4: Brainfuck의 [
5: Brainfuck의 ]
6: Brainfuck의 . (주의: 256의 나머지만을 계산하여 출력합니다)
7: Brainfuck의 ,

니블코드의 첫 비트가 1인 경우, 다음의 8니블(4바이트)은 해당 코드를 반복하는 횟수를 표시합니다.
타입은 부호 없는 32비트 정수형입니다.
*/

// TODO: 테스트 케이스 추가(HelloWorld)
type minfuckVM struct {
	code []byte
	mem  []uint32
	pc   uint32 // Program counter, 'nibble' offset
	mp   uint32 // Memory offset
	bs   uint32 // Braces stack
	Out  io.Writer
	In   io.Reader
}

func vmFile(f *os.File) (*minfuckVM, error) {
	meta, err := readFile(f)
	if err != nil {
		return nil, err
	}

	vm := new(minfuckVM)
	vm.mem = make([]uint32, meta.memsize*2+3)
	for i := uint32(0); i < meta.memsize; i++ {
		vm.mem[3+i*2] = i + 1 // Memory init
	}

	vm.code = meta.code
	vm.Out, vm.In = os.Stdout, os.Stdin

	return vm, nil
}

// run 함수는 VM이 종료될 때까지 구동합니다
func (vm *minfuckVM) run() error {
	for {
		err := vm.process()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}

// process 함수는 단일 MinFuck operation을 처리합니다.
func (vm *minfuckVM) process() error {
	c, err := vm.nibble()
	if err != nil {
		return err
	}
	rep := uint32(1)
	if (c>>3)&1 == 1 {
		nn, err := vm.nibbleN(30)
		if err != nil {
			return err
		}
		rep = nibbleU32(nn)
	}
	for i := uint32(0); i < rep; i++ {
		vm.runcode(c & 7)
	}
	return nil
}

// runcode 함수는 한 개의 니블코드를 VM에서 실행합니다
func (vm *minfuckVM) runcode(nc byte) {
	switch nc {
	case 0: // +
		vm.mem[vm.mp]++
	case 1: // -
		vm.mem[vm.mp]--
	case 2: // >
		vm.mp = (vm.mp + 1) % uint32(len(vm.mem))
	case 3: // <
		vm.mp = (vm.mp - 1) % uint32(len(vm.mem))
	case 4: // [ TODO
	case 5: // ] TODO
	case 6: // .
		vm.Out.Write([]byte{byte(vm.mem[vm.mp])})
	case 7: // ,
		b := make([]byte, 1)
		vm.In.Read(b)
		vm.mem[vm.mp] = uint32(b[0])
	}
}

// TODO: 테스트 케이스 추가
func (vm *minfuckVM) nibble() (byte, error) {
	if vm.pc>>1 >= uint32(len(vm.code)) {
		return 0, io.EOF
	}
	n := vm.code[vm.pc>>1] >> (vm.pc & 1)
	vm.pc++
	return n, nil
}

// TODO: 테스트 케이스 추가
func (vm *minfuckVM) nibbleN(n uint32) ([]byte, error) {
	b := make([]byte, n)
	var err error
	for i := uint32(0); i < n; i++ {
		b[i], err = vm.nibble()
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

// ======= MinFuck utils =======

// nibbleU32는 8개의 니블 배열을 부호 없는 32비트 정수형으로 변환합니다
// 주의: 슬라이스의 길이가 8 미만일 경우 panic이 발생합니다.
// TODO: 테스트 케이스 추가
func nibbleU32(nb []byte) uint32 {
	return uint32(nb[0]&0xf)<<28 | uint32(nb[1]&0xf)<<24 |
		uint32(nb[2]&0xf)<<20 | uint32(nb[3]&0xf)<<16 |
		uint32(nb[4]&0xf)<<12 | uint32(nb[5]&0xf)<<8 |
		uint32(nb[6]&0xf)<<4 | uint32(nb[7]&0xf)
}

// u32Bytes는 부호 없는 32비트 정수를 바이트 슬라이스로 변환합니다.
// TODO: 테스트 케이스 추가
func u32Bytes(n uint32) []byte {
	return []byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

// bfToMf는 Brainfuck 코드를 MinFuck 코드로 변환합니다.
// Brainfuck에는 사실상 memory address limit이 없기 때문에, 이는 수동으로 지정해야 합니다.
// TODO: 코드 반복 압축
// TODO: 테스트 케이스 추가
func bfToMf(bf string, mem uint32) (mf string) {
	fd := fileData{memsize: mem, code: make([]byte, func() int {
		l := len(bf)
		if l&1 == 0 {
			return l >> 1
		}
		return (l >> 1) + 1
	}())}
	odd := false
	for i := 0; i < len(bf); i++ {
		m := bfOp(bf[i : i+1])
		if m > 0xf {
			continue
		}
		if odd {
			fd.code[i>>1] = (fd.code[i>>1] & 0xf0) | (m & 0xf)
		} else {
			fd.code[i>>1] = (m & 0xf) << 4
		}
	}
	mf = fd.toString()
	return
}

// TODO: 테스트 케이스 추가(BF 코드 아닌 경우 escape)
func bfOp(bf string) (mf byte) {
	switch bf {
	case "+":
		return 0
	case "-":
		return 1
	case ">":
		return 2
	case "<":
		return 3
	case "[":
		return 4
	case "]":
		return 5
	case ".":
		return 6
	case ",":
		return 7
	}
	return 255
}
