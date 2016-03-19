# MinFuck [![GoDoc](https://godoc.org/github.com/cr0sh/minfuck/mf?status.svg)](https://godoc.org/github.com/cr0sh/minfuck/mf) [![Go Report Card](https://goreportcard.com/badge/github.com/cr0sh/minfuck/mf)](https://goreportcard.com/report/github.com/cr0sh/minfuck/mf)

MinFuck은 Brainfuck에서 영감을 얻은 난해한 프로그래밍 언어입니다.

기본적으로 모든 MF 코드는 BF로 1:1 변환 가능하며(역은 메모리 가용성이 충분한 상황에서 허용됩니다), 각 BF 코드는 4비트 크기의 니블코드로 변환됩니다. 또한, 각 니블코드의 첫 비트는 같은 코드를 최대 42억 번 반복할지를 결정할 수 있으므로 코드의 길이를 획기적으로 줄일 수 있습니다.

MinFuck 구현의 주 목적은 `Polygon`을 비롯한 고수준 프로그래밍 언어에 대한 저수준 VM 기계어를 제공하는 것입니다.

MinFuck 및 minfuck/mf 모듈은 MIT 허가서 하에서 배포됩니다.

## Usage
```
사용법: minfuck [command] [option1 option2 ...]
help:
    지금 보고 있는 도움말을 출력합니다.

b2m [filename] [mem]:
    주어진 Brainfuck 코드를 MinFuck 코드로 변환합니다.
    mem은 할당할 메모리 주소의 최댓값이며, 기본값은 4096입니다.

run [filename]:
    주어진 MinFuck 코드를 구동합니다.
bfr [filename]:
    주어진 Brainfuck 코드를 구동합니다.
```

## Credits&Thanks
먼저 `Brainfuck`이라는 멋진 언어를 만든 Urban Müller에게 감사드립니다.

`helloworld.bf` 테스트 코드는 [esolangs.org](https://esolangs.org/wiki/Brainfuck)에서 제공되는 예제입니다.

## License

Copyright (c) 2016 cr0sh(Nam J.H.)(ska827@naver.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
