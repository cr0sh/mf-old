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
func U32Bytes(n uint32) []byte {
	return []byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

// BfToMf 함수는 Brainfuck 코드를 MinFuck 코드로 변환합니다.
// Brainfuck에는 사실상 memory address limit이 없기 때문에, 수동으로 지정해야 합니다.
// TODO: 코드 반복 압축
func BfToMf(bf string, mem uint32) (mf string) {
	fd := FileData{memsize: mem}
	nw := new(NibbleWriter)
	for i := 0; i < 4; i++ {
		nw.Put(2)
	}
	for _, b := range bf {
		op := FromBf(string(b))
		if op > 7 {
			continue
		}
		if op == 2 || op == 3 {
			nw.Put(op) // Increment/Decrement pointer twice because MF uses another mem structure
		}
		nw.Put(op)
	}
	fd.code = nw.Nibbles
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

// NibbleWriter 타입은 니블코드를 byte array로 변환해줍니다.
type NibbleWriter struct {
	Nibbles []byte
	odd     bool
}

// Put 메서드는 니블코드를 byte array에 작성합니다.
func (n *NibbleWriter) Put(nb byte) {
	if len(n.Nibbles) == 0 {
		n.Nibbles = []byte{(nb & 0xf) << 4}
		n.odd = true
	} else if n.odd {
		n.Nibbles[len(n.Nibbles)-1] |= nb & 0xf
		n.odd = false
	} else {
		b := (nb & 0xf) << 4
		n.Nibbles = append(n.Nibbles, b)
		n.odd = true
	}
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
