# alpcgo

Go tools for basic ALPC hacking. Built on my fork of
https://github.com/AllenDang/w32 which has additions to ntdll and avdapi32 to
support the required parts of the native API

This code is not heavily tested, its purpose is primarily didactic.

## Documentation

Use `import "github.com/bnagy/alpcgo"` in your own Go code.

Get godoc at: http://godoc.org/github.com/bnagy/alpcgo

## Installation

You should follow the [instructions](https://golang.org/doc/install) to
install Go, if you haven't already done so. Then:

```bash
$ go get github.com/bnagy/alpcgo
```

## package alpcgo

Higher level API for basic ALPC functions like Send, Connect...

## Utility Commands ( cmd/ directory )

### alpcechosrv

PoC Echo Server ( part of hello world )

### alpcechocli

PoC Echo Client ( part of hello world )

### alpcbridge

JSON-RPC bridge with a simple API. Designed to make it easy to connect to raw
ALPC ports from any language to build fuzzers or other tools.

### alpcechoclij

PoC Echo Client using the jsonrpc bridge

### alpcrest

A jsonrpc bridge that listens to http POST on a /rpc endpoint, for HLL clients
that find that easier.

## TODO

- Add Attribute support

## Bugs

- No x86 suppport

## Contributing

Fork & pullreq

## License

BSD Style, See LICENSE file for details



