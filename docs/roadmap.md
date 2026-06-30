# Roadmap

`go-ruby-zlib/zlib` is grown **test-first**, each capability differential-tested against MRI
rather than built in isolation. Ruby's Zlib — the deterministic,
interpreter-independent slice — is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| Deflate / Inflate | `Deflate(data, level)` / `Inflate(data)` — zlib-stream compression at any level, mirroring `Zlib::Deflate.deflate` / `Zlib::Inflate.inflate`; round-trips and interoperates with MRI in both directions. | **Done** |
| Gzip round-trip | `GzipCompress` / `GzipDecompress` mirror `Zlib::GzipWriter` / `Zlib::GzipReader`, with a deterministic (zero) header mtime; compare by decompressed payload + CRC, not raw gzip bytes. | **Done** |
| Byte-exact checksums | `Crc32` / `Adler32` and `Crc32Combine` / `Adler32Combine` — **byte-exact** with MRI, with seeded (running) checksums; a value computed here equals the integer MRI prints. | **Done** |
| Streaming Deflater / Inflater | The `Zlib::Deflate.new(level).deflate(s, flush).finish` / `Zlib::Inflate.new.inflate(s)` idiom, with `NoFlush`/`SyncFlush`/`FullFlush`/`Finish` modes and the `TotalIn` / `TotalOut` / `Adler` / `Finished` accessors. | **Done** |
| Constants & errors | Compression levels, strategies, flush modes, `Version` / `ZlibVersion`, and an `*Error` family carrying MRI's class names (`Zlib::StreamError`, `Zlib::BufError`, `Zlib::DataError`, `Zlib::GzipFile::Error`). | **Done** |
| Differential oracle & coverage | Checksums and `combine` values compared byte-exact with the system `ruby`; deflate / gzip / streaming round-tripped through MRI in both directions. 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Conformance targets MRI's behaviour;
  differences across MRI releases are matched to the reference used by the
  differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
