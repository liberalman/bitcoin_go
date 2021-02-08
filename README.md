# bitcoin_go
bitcoin by go

set env
```sh
$ go env -w GO111MODULE=on

$ go env -w GOPROXY=https://goproxy.io,direct
# 或者
$ export GOPROXY=https://goproxy.io
```

build
```sh
$ go build
```

run
```sh
$ ./bitcoin_go 
```

unit test
```sh
$ go test -v ./tests/                  # test all cases
$ go test -v ./tests/ -bench="Base58$" # test Base58
```



