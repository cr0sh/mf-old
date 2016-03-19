package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/cr0sh/minfuck/mf"
)

const version = "0.1"
const mainHelp = `MinFuck 버전 %s
도움말: %s [operation] option1, ...
help: 지금 보고 있는 도움말을 출력합니다.
b2m [filename]: 주어진 Brainfuck 코드를 MinFuck 코드로 변환합니다.
run [filename]: 주어진 MinFuck 코드를 구동합니다.
bfr [filename]: 주어진 Brainfuck 코드를 구동합니다.
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
	case "bfr":
		bfr()
	default:
		fmt.Println("정의되지 않은 동작:", os.Args[1])
		help()
	}
}

func b2m() {
	fmt.Println("아직 사용할 수 없습니다") // FIXME
	os.Exit(0)
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
		[]byte(mf.BfToMf(string(b), 4096)),
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
	vm, err := mf.VMFile(f)
	if err != nil {
		fmt.Println("VM 준비 중 오류:", err)
		os.Exit(4)
	}
	result := make(chan error, 1)
	vm.Run(nil, result)
	if err := <-result; err != nil {
		fmt.Printf("\n코드가 비정상 종료되었습니다: %s\n", err.Error())
		os.Exit(2)
	} else {
		fmt.Printf("\n코드가 정상적으로 종료되었습니다\n")
		os.Exit(0)
	}
}

func bfr() {
	if len(os.Args) < 3 {
		fmt.Println("실행할 Brainfuck 코드가 필요합니다.")
		help()
	}
	s, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("파일 여는 중 오류:", err)
		os.Exit(3)
	}

	b := new(bytes.Buffer)
	var bitbuf byte
	odd := false
	for _, c := range []byte(s) {
		c = mf.FromBf(string(c))
		if c == 255 {
			continue
		}
		if odd {
			bitbuf = (bitbuf & 0xf0) | c
			b.WriteByte(bitbuf)
		} else {
			bitbuf = c << 4
		}
		odd = !odd
	}
	if odd {
		b.WriteByte(bitbuf)
	}

	vm := mf.MinFuckVM{Code: b.Bytes(), Mem: make([]uint32, 1<<20), Out: os.Stdout, In: os.Stdin}
	stop, result := make(chan struct{}, 1), make(chan error, 1)
	duration, _ := time.ParseDuration("10s")
	time.AfterFunc(duration, func() {
		fmt.Println("프로그램이 너무 길게 동작합니다. 강제로 종료합니다.")
		stop <- struct{}{}
	})
	vm.Run(stop, result)
	if err := <-result; err != nil {
		fmt.Printf("\n코드가 비정상 종료되었습니다: %s\n", err.Error())
		os.Exit(2)
	} else {
		fmt.Printf("\n코드가 정상적으로 종료되었습니다\n")
		os.Exit(0)
	}
}

func help() {
	fmt.Printf(mainHelp, version, os.Args[0])
	os.Exit(0)
}
