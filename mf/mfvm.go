package mf

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const mfMagic = "\xff\x6d\x66\xfd"

/*
FileData 구조체는 MinFuck 소스 코드의 메타데이터를 정의합니다.

 MinFuck binary 포맷

 .mf 파일의 첫 4바이트는 Magic Byte(\xff\x6d\x66\xfd)입니다.
 다음 4바이트에 부호 없는 정수형으로 MinFuck VM에서 접근 가능한 최대 메모리 번지를 지정합니다.
 (단, 실제 OS에서는 최소 해당 값 * 8 + 24바이트 이상을 할당합니다.)
 다음 4바이트에는 부호 없는 정수형으로 코드의 크기를 명시합니다.
*/
type FileData struct {
	memsize uint32
	code    []byte
}

// ReadFile 함수는 주어진 파일로부터 정보를 읽어 MinFuck 파일 메타데이터로 변환합니다.
func ReadFile(f io.Reader) (FileData, error) {
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return FileData{}, err
	}
	if !bytes.Equal(magic, []byte(mfMagic)) {
		return FileData{}, fmt.Errorf("잘못된 MinFuck Magic: 0x" + hex.EncodeToString(magic))
	}

	membuf := make([]byte, 4)
	if _, err := f.Read(membuf); err != nil {
		return FileData{}, err
	}
	memsize, n := binary.Uvarint(membuf)
	if n <= 0 {
		return FileData{}, fmt.Errorf("메모리 주소 제한 값이 잘못되었습니다")
	}

	codebuf := make([]byte, 4)
	if _, err := f.Read(codebuf); err != nil {
		return FileData{}, err
	}
	codesize, n := binary.Uvarint(codebuf)
	if n <= 0 {
		return FileData{}, fmt.Errorf("코드 길이 값이 잘못되었습니다")
	}

	code := make([]byte, codesize)
	if _, err := f.Read(code); err != nil {
		return FileData{}, err
	}

	return FileData{memsize: uint32(memsize), code: code}, nil
}

// String 메서드는 FileData를 string으로 변환합니다.
func (f *FileData) String() string {
	buf := bytes.NewBuffer([]byte(mfMagic))
	buf.Write(U32Bytes(f.memsize))
	buf.Write(U32Bytes(uint32(len(f.code))))
	buf.Write(f.code)
	return buf.String()
}

/*
MinFuckVM 구조체는 MinFuck 코드를 구동하기 위한 가상 머신(VM) 환경을 정의합니다.

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

TODO: 테스트 케이스 추가(HelloWorld)
*/
type MinFuckVM struct {
	Code []byte
	Mem  []uint32
	pc   uint32 // Program counter, 'nibble' offset
	mp   uint32 // Memory offset
	bs   uint32 // Braces stack
	ts   uint32 // Target stack to match braces
	bt   byte   // Braces status; 0: nothing, 1: searching ']', 2: searching '['
	Out  io.Writer
	In   io.Reader
}

// VMFile 함수는 주어진 MinFuck 소스 파일로부터 VM을 생성해 반환합니다.
func VMFile(f io.Reader) (*MinFuckVM, error) {
	meta, err := ReadFile(f)
	if err != nil {
		return nil, err
	}

	vm := new(MinFuckVM)
	vm.Mem = make([]uint32, meta.memsize*2+3)
	for i := uint32(0); i < meta.memsize; i++ {
		vm.Mem[3+i*2] = i + 1 // Memory init
	}

	vm.Code = meta.code
	vm.Out, vm.In = os.Stdout, os.Stdin

	return vm, nil
}

// Run 메서드는 VM이 종료될 때까지 구동합니다.
// VM을 강제로 멈추려면 stop 채널에 신호를 보냅니다.
// VM이 실행을 마치면 에러 여부를 report 채널에 보고합니다.
func (vm *MinFuckVM) Run(stop <-chan struct{}, report chan<- error) {
	for {
		select {
		case <-stop:
			report <- nil
			return
		default:
			err := vm.Process()
			if err == io.EOF {
				report <- nil
				return
			} else if err != nil {
				report <- err
				return
			}
		}
	}
}

// Process 메서드는 단일 MinFuck operation을 처리합니다.
func (vm *MinFuckVM) Process() error {
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
		rep = NibbleU32(nn)
	}
	for i := uint32(0); i < rep; i++ {
		vm.RunCode(c & 7)
	}
	return nil
}

// RunCode 함수는 한 개의 니블코드를 VM에서 실행합니다
func (vm *MinFuckVM) RunCode(nc byte) {
	if vm.bracketCheck(nc) {
		// vm.dump()
		switch nc {
		case 0: // +
			vm.Mem[vm.mp]++
		case 1: // -
			vm.Mem[vm.mp]--
		case 2: // >
			vm.mp = (vm.mp + 1) % uint32(len(vm.Mem))
		case 3: // <
			vm.mp = (vm.mp - 1) % uint32(len(vm.Mem))
		case 6: // .
			vm.Out.Write([]byte{byte(vm.Mem[vm.mp])})
		case 7: // ,
			b := make([]byte, 1)
			vm.In.Read(b)
			vm.Mem[vm.mp] = uint32(b[0])
		}
	}
}

// bracketCheck 메서드는 VM이 맞는 대괄호 짝을 찾는 중일 때 코드를 실행하지 않도록 합니다.
func (vm *MinFuckVM) bracketCheck(nc byte) bool {
	if vm.bt == 0 {
		if nc == 4 && vm.Mem[vm.mp] == 0 {
			vm.bt, vm.ts = 1, vm.bs
			return false
		} else if nc == 5 && vm.Mem[vm.mp] != 0 {
			vm.bt, vm.ts = 2, vm.bs
			vm.pc -= 2
			return false
		}
		return true
	} else if vm.bt == 1 {
		vm.bracketStack(nc)
		return false
	} else {
		vm.pc -= 2
		vm.bracketStack(nc)
		return false
	}
}

// bracketStack 메서드는 대괄호 스택을 조정합니다
func (vm *MinFuckVM) bracketStack(nc byte) {
	if nc == 4 {
		vm.bs++
	} else if nc == 5 {
		if vm.bs == 0 {
			panic("Invalid loop brackets")
		}
		vm.bs--
	}
	if vm.ts == vm.bs {
		vm.bt = 0
	}
}

func (vm *MinFuckVM) nibble() (byte, error) {
	if vm.pc>>1 >= uint32(len(vm.Code)) {
		return 0, io.EOF
	}
	n := (vm.Code[vm.pc>>1] >> (((vm.pc & 1) ^ 1) << 2)) & 0xf
	vm.pc++
	return n, nil
}

func (vm *MinFuckVM) nibbleN(n uint32) ([]byte, error) {
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

func (vm *MinFuckVM) dump() {
	fmt.Printf(`VM Status Dump
    PC: %d
    MP: %d
    BS: %d
    TS: %d
    BT: %d

`, vm.pc, vm.mp, vm.bs, vm.ts, vm.bt)
}
