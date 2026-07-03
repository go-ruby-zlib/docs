<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-zlib` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-zlib` library**
against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby, TruffleRuby). It
measures the **library primitive** through its Go API, isolated from the rbgo
interpreter, so the numbers answer: *is the pure-Go implementation as fast as the
reference runtime's own `zlib`?*

Ruby's `zlib` is a C extension wrapping the system C `zlib`; `go-ruby-zlib` is
pure Go (`github.com/klauspost/compress` DEFLATE + the standard library's
SIMD `hash/crc32` / `hash/adler32`), so on raw deflate throughput MRI's C zlib
may win — the harness reports the real numbers either way.

## Layout

- `go/`           — self-contained Go driver; `go.mod` pins the **published**
  library by pseudo-version (no `replace`). The built `bench` binary is
  `.gitignore`d.
- `ruby/zlib.rb`  — the equivalent workload; `ruby/_harness.rb` is the shared timer.
- `run.sh`        — runs every available runtime and prints one Markdown table per
  sub-benchmark (ns/op + ratio vs MRI).

## Workload

A deterministic **~64 KiB "mixed text / repetitive" buffer** (a repeating English
phrase perturbed every 97th byte with a deterministic printable byte). The Go and
Ruby sides build the **byte-identical** buffer — both print its CRC-32 / Adler-32
to stderr so you can confirm. Operations, all at a fixed compression level 6:

- `deflate-64KiB-L6` — compress the buffer to a zlib stream.
- `inflate-64KiB-L6` — decompress that stream back.
- `crc32-64KiB`      — CRC-32 (IEEE) of the buffer.
- `adler32-64KiB`    — Adler-32 of the buffer.

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. **Round-trip correctness is verified before timing**: both drivers
assert `inflate(deflate(buf)) == buf` and abort otherwise.

The exact deflate byte stream is implementation-defined — Go's flate writer and
C zlib frame the stream differently, so `go-ruby-zlib`'s compressed bytes need
not equal MRI's (both sides print their compressed size). The **decompressed
payload, and the CRC-32 / Adler-32 checksums, are byte-exact with MRI.** Results
are published, dated, in [`../docs/performance.md`](../docs/performance.md).
