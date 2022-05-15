# go-shlex

go-shlex is a library to make a lexical analyzer like Unix shell for
Go.

## Install

    go get -u "github.com/chroblert/go-shlex"
## description
```
Split(<目标字符串>,<是否posix模式>,<是否保留所有字面量>,[指定分隔符])
    当未指定分隔符，且不保留所有字面量时,即Split(<目标字符串>,<是否posix模式>,false)，与原来的函数作用相同(desertbit/go-shlex)
OriginSplit(<目标字符串>,<是否posix模式>)
```
```
是否保留所有字面量：指是否保留',",\这些特殊字符。
假设输入：set arg="argV1\"",argV2
- 若为true:  set arg "argV1\"",argV2
- 若为false: set arg argV1",argV2
```
## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/chroblert/go-shlex"
)

func main() {
    cmd := `cp -Rdp "file name" 'file name2' dir\ name`
    words, err := shlex.Split(cmd, true,false)
    if err != nil {
        log.Fatal(err)
    }

    for _, w := range words {
        fmt.Println(w)
    }
}
```
output

    cp
    -Rdp
    file name
    file name2
    dir name

## Documentation

http://godoc.org/github.com/anmitsu/go-shlex

