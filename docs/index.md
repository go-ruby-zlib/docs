# go-ruby-zlib documentation

**Ruby's `Zlib` in pure Go — deflate, gzip, byte-exact CRC-32/Adler-32, no cgo.**

`go-ruby-zlib/zlib` is a faithful, pure-Go (zero cgo) reimplementation of Ruby's [`Zlib`](https://docs.ruby-lang.org/en/master/Zlib.html),
matching reference Ruby (MRI). The module path is `github.com/go-ruby-zlib/zlib`.

It is the backend bound into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — just like
[go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb). The dependency runs the other way:
this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete"
    **Complete — checksums byte-exact, compressors MRI-interoperable.** Faithful port of Ruby's `Zlib`: **Deflate / Inflate**, **gzip** round-trip, the **byte-exact** **CRC-32 / Adler-32** checksums and their `combine` forms, and a **streaming** Deflater / Inflater. Validated by a **differential oracle** against the system `ruby` — checksums byte-for-byte, deflate / gzip / streaming round-tripped through MRI in both directions — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Quick taste

```go
comp, _ := zlib.Deflate([]byte("hello world"), zlib.BestCompression)
out, _ := zlib.Inflate(comp)
// out == "hello world"

// Byte-exact checksums (seed 0 for CRC, 1 for Adler — the MRI identities).
zlib.Crc32([]byte("hello world"), 0)   // 222957957
zlib.Adler32([]byte("hello world"), 1) // 436929629
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`zlib`](https://github.com/go-ruby-zlib/zlib) | the library — Ruby's Zlib in pure Go |
| [`docs`](https://github.com/go-ruby-zlib/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-zlib.github.io`](https://github.com/go-ruby-zlib/go-ruby-zlib.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-zlib/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** Output matches reference Ruby, validated by a differential
  oracle against the `ruby` binary.
- **Standalone & reusable.** A standalone module bound by `rbgo`; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why this slice of Ruby is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-zlib/zlib](https://github.com/go-ruby-zlib/zlib).
