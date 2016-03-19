package mf

import "io"

// NibbleU32 함수는 8개의 니블 배열을 부호 없는 32비트 정수형으로 변환합니다
// 주의: 슬라이스의 길이가 8 미만일 경우 panic이 발생합니다.
// TODO: 테스트 케이스 추가
func NibbleU32(nb []byte) uint32 {
	return uint32(nb[0]&0xf)<<28 | uint32(nb[1]&0xf)<<24 |
		uint32(nb[2]&0xf)<<20 | uint32(nb[3]&0xf)<<16 |
		uint32(nb[4]&0xf)<<12 | uint32(nb[5]&0xf)<<8 |
		uint32(nb[6]&0xf)<<4 | uint32(nb[7]&0xf)
}

// U32Bytes 함수는 부호 없는 32비트 정수를 바이트 슬라이스로 변환합니다.
// TODO: 테스트 케이스 추가
func U32Bytes(n uint32) []byte {
	return []byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

// BfToMf 함수는 Brainfuck 코드를 MinFuck 코드로 변환합니다.
// Brainfuck에는 사실상 memory address limit이 없기 때문에, 이는 수동으로 지정해야 합니다.
// TODO: 코드 반복 압축
// TODO: 테스트 케이스 추가
func BfToMf(bf string, mem uint32) (mf string) {
	fd := FileData{memsize: mem, code: make([]byte, func() int {
		l := len(bf)
		if l&1 == 0 {
			return l >> 1
		}
		return (l >> 1) + 1
	}())}
	odd := false
	for i := 0; i < len(bf); i++ {
		m := FromBf(bf[i : i+1])
		if m > 0xf {
			continue
		}
		if odd {
			fd.code[i>>1] = (fd.code[i>>1] & 0xf0) | (m & 0xf)
		} else {
			fd.code[i>>1] = (m & 0xf) << 4
		}
	}
	mf = fd.String()
	return
}

// FromBf 함수는 Brainfuck 코드를 MinFuck 코드로 변환합니다.
// TODO: 테스트 케이스 추가(BF 코드 아닌 경우 escape)
func FromBf(bf string) (mf byte) {
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

// ToBf 함수는 MinFuck 코드를 BrainFuck 코드로 변환합니다.
func ToBf(mf byte) (p string) {
	switch mf {
	case 0:
		p = "+"
	case 1:
		p = "-"
	case 2:
		p = ">"
	case 3:
		p = "<"
	case 4:
		p = "["
	case 5:
		p = "]"
	case 6:
		p = "."
	case 7:
		p = ","
	}
	return p
}

// IOStream 구조체는 stdin/stdout을 에뮬레이션합니다.
// 주로 디버깅에 사용됩니다.
type IOStream struct {
	Stdin  string
	Stdout string
	offset uint64
}

// Read 메서드는 io.Reader 인터페이스를 구현합니다.
// MinFuck VM의 Stdin을 에뮬레이션합니다.
func (i *IOStream) Read(b []byte) (int, error) {
	if i.offset >= uint64(len(i.Stdin)) {
		return 0, io.EOF
	}
	n := copy(b, []byte(i.Stdin[i.offset:]))
	return n, nil
}

// Write 메서드는 io.Writer 인터페이스를 구현합니다.
// MinFuck VM의 Stdout을 에뮬레이션합니다.
func (i *IOStream) Write(b []byte) (int, error) {
	i.Stdout += string(b)
	return len(b), nil
}
