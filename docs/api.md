# Usage & API

The public API lives at the module root (`github.com/go-ruby-zlib/zlib`). It is **Ruby-shaped but
Go-idiomatic**: the surface mirrors MRI's `Zlib`, while following Go conventions —
an explicit `error` where Ruby raises, value types, no global state.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-zlib/zlib`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-zlib/zlib
```

## Worked example

```go
package main

import (
	"fmt"

	github.com/go-ruby-zlib/zlib
)

func main() {
	comp, _ := zlib.Deflate([]byte("hello world"), zlib.BestCompression)
	out, _ := zlib.Inflate(comp)
	// out == "hello world"

	// Byte-exact checksums (seed 0 for CRC, 1 for Adler — the MRI identities).
	zlib.Crc32([]byte("hello world"), 0)   // 222957957
	zlib.Adler32([]byte("hello world"), 1) // 436929629
	_ = fmt.Sprint
}
```

## Shape

```go
// One-shot
func Deflate(data []byte, level int) ([]byte, error) // Zlib::Deflate.deflate
func Inflate(data []byte) ([]byte, error)            // Zlib::Inflate.inflate
func GzipCompress(data []byte, level int) ([]byte, error)
func GzipDecompress(data []byte) ([]byte, error)

// Checksums (byte-exact with MRI)
func Crc32(data []byte, seed uint32) uint32
func Adler32(data []byte, seed uint32) uint32
func Crc32Combine(crc1, crc2 uint32, len2 int64) uint32
func Adler32Combine(adler1, adler2 uint32, len2 int64) uint32
func Crc32Table() *crc32.Table

// Streaming
type Deflater struct{ /* … */ }
func NewDeflater(level int) *Deflater
func NewDeflaterLevel(level int) (*Deflater, error)
func (d *Deflater) Deflate(data []byte, flush int) ([]byte, error)
func (d *Deflater) Finish() ([]byte, error)
func (d *Deflater) TotalIn() int64
func (d *Deflater) TotalOut() int64
func (d *Deflater) Adler() uint32
func (d *Deflater) Finished() bool

type Inflater struct{ /* … */ }
func NewInflater() *Inflater
func (i *Inflater) Inflate(data []byte) ([]byte, error)
func (i *Inflater) Finish() ([]byte, error)
func (i *Inflater) TotalIn() int64
func (i *Inflater) TotalOut() int64
func (i *Inflater) Adler() uint32
func (i *Inflater) Finished() bool

// Errors — *Error carries the MRI exception Class name; all wrap to Error.
type Error struct{ Class, Msg string /* … */ }
var ErrStream, ErrBuf, ErrData, ErrGzipFile *Error
```

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a wide
corpus through both the system `ruby` and this library and compares the results
**byte-for-byte** — not approximated from memory. The oracle tests skip
themselves where `ruby` is not on `PATH` (e.g. the qemu arch lanes), so the
cross-arch builds still validate the library.

## Relationship to Ruby

`go-ruby-zlib/zlib` is **standalone and reusable**, and is the backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a
native module — the same way [go-ruby-regexp](https://github.com/go-ruby-regexp)
and [go-ruby-erb](https://github.com/go-ruby-erb) are bound. The dependency runs
the other way: this library has no dependency on the Ruby runtime.
