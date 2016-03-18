package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/cr0sh/minfuck/mf"
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
	err = vm.Run()
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
