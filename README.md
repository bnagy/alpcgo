Overview
======

Go tools for basic ALPC hacking. Built on my fork of
https://github.com/AllenDang/w32 which has additions to ntdll and avdapi32 to
support the required parts of the native API

package alpcgo
======

Higher level API for basic ALPC functions like Send, Connect...

alpcechosrv
======

PoC Echo Server ( part of hello world )

alpcechocli
======

PoC Echo Client ( part of hello world )

alpcbridge
======

JSON-RPC bridge with a simple API. Designed to make it easy to connect to raw
ALPC ports from any language to build fuzzers or other tools.

alpcechoclij
======

PoC Echo Client using the jsonrpc bridge

misc/alpclog
======

Quick Ruby script to log ALPC messages received by a process. Should work for most processes, but not on unbreakable system processes ( csrss etc ). Requires https://github.com/bnagy/rBuggery >= v1.1.0

TODO:
=======

- Add Attribute support

BUGS
=======

- No x86 suppport

Contributing
=======

Fork & pullreq

License
=======

BSD Style, See LICENSE file for details



