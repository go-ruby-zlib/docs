# Performance

`go-ruby-zlib/zlib` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `zlib`. This
page records the **methodology** for the comparative benchmark of that module
against the reference Ruby runtimes, part of the ecosystem-wide per-module parity
suite.

## What is measured

The **same** Ruby script — a representative `Zlib` workload — is run under every
runtime. `rbgo`'s number reflects **this pure-Go library doing the work**; every
other column is that interpreter's own `zlib` stdlib. So the comparison is the
**Ruby-visible operation**, apples-to-apples across interpreters. The script
prints a deterministic checksum and its output is checked **byte-identical to
MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (OpenJDK);
  `truffleruby` (GraalVM CE Native).
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`zlib.rb` + `run.sh`). Reproduce:
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh 5`.

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-zlib) | 360 | 6.00× |
| MRI (ruby 4.0.5) | 60 | 1.00× |
| MRI + YJIT | 60 | 1.00× |
| JRuby 10.1.0.0 | 1220 | 20.33× |
| TruffleRuby 34.0.1 | 380 | 6.33× |

rbgo runs on **go-ruby-zlib** and is **~6x slower than MRI** here (6.0x): MRI's `zlib` is a thin wrapper over C zlib, while go-ruby-zlib goes through Go's `compress/flate` — competitive but not at C-zlib throughput on this deflate+inflate loop. Honest gap, flagged for the go-ruby-zlib perf backlog.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are **real measured numbers** from
    the 2026-06-30 run (Apple M-series; `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`,
    `truffleruby 34.0.1`) — nothing is fabricated or cherry-picked.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `zlib`?* Ruby's
`zlib` is a C extension wrapping the system C `zlib`, whereas `go-ruby-zlib` is
pure Go (`github.com/klauspost/compress` for DEFLATE, plus the standard library's
`hash/crc32` / `hash/adler32` for the checksums) — so this is a pure-Go stack
measured head-to-head against a C library, and we report the real numbers either
way.

- **Host:** Apple M4 Max (arm64), macOS — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Library pinned:** `github.com/go-ruby-zlib/zlib@v0.0.0-20260703114753-dbf176c108fc`
  (published pseudo-version, no `replace`).
- **Workload:** a deterministic **~64 KiB "mixed text / repetitive" buffer** (a
  repeating English phrase perturbed every 97th byte). The Go and Ruby drivers
  build the **byte-identical** buffer — both confirm `crc32=ec32c294`,
  `adler32=78a49dc4` — at a fixed **compression level 6**.
- **Correctness, verified before timing:** both drivers assert
  `inflate(deflate(buf)) == buf` (round-trip OK on both). The raw deflate byte
  streams **differ** (Go flate frames 2686 B, MRI C zlib frames 2499 B for this
  buffer): zlib never promises a canonical encoding, so this is expected — the
  **decompressed payload and the CRC-32 / Adler-32 checksums are byte-exact with
  MRI**.
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

The harness lives in this repo under
[`benchmarks/`](https://github.com/go-ruby-zlib/docs/tree/main/benchmarks)
(`go/` driver pinning the published library, `ruby/zlib.rb`, `run.sh`).
Reproduce: `bash benchmarks/run.sh`.

#### deflate-64KiB-L6

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 83199.6 | 0.14× |
| MRI | 597440.0 | 1.00× |
| MRI + YJIT | 598860.0 | 1.00× |
| JRuby | 784871.7 | 1.31× |
| TruffleRuby | 613396.2 | 1.03× |

#### inflate-64KiB-L6

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 43174.6 | 2.53× |
| MRI | 17090.0 | 1.00× |
| MRI + YJIT | 17450.0 | 1.02× |
| JRuby | 74047.1 | 4.33× |
| TruffleRuby | 43967.9 | 2.57× |

#### crc32-64KiB

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 5443.8 | 3.86× |
| MRI | 1412.0 | 1.00× |
| MRI + YJIT | 1552.0 | 1.10× |
| JRuby | 6236.3 | 4.42× |
| TruffleRuby | 2299.9 | 1.63× |

#### adler32-64KiB

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 17609.0 | 6.08× |
| MRI | 2894.0 | 1.00× |
| MRI + YJIT | 2844.0 | 0.98× |
| JRuby | 2385.1 | 0.82× |
| TruffleRuby | 3838.1 | 1.33× |

### Reading the numbers

- **Deflate — go-ruby is ~7× faster than MRI (0.14×).** This is the headline and
  it is real: `klauspost/compress` at level 6, with the engine pooling landed in
  the library (reusing the compressor across one-shot calls instead of allocating
  a fresh `z_stream` each time), beats MRI's C zlib on this buffer. Both YJIT and
  TruffleRuby track MRI (the work is inside the C extension, not the interpreter).
- **Inflate — go-ruby is ~2.5× slower than MRI (2.53×).** MRI's C zlib inflate is
  faster than pure-Go flate decompression here; go-ruby lands right alongside
  TruffleRuby (2.57×) and well ahead of JRuby (4.33×). This is the clearest
  optimization target for the library.
- **CRC-32 — go-ruby is ~3.9× slower than MRI (3.86×).** Go's `hash/crc32` IEEE
  path on arm64 is a generic slicing-by-8 table, not a hardware-CRC kernel, so it
  trails zlib's optimized checksum; go-ruby is still faster than JRuby (4.42×).
- **Adler-32 — go-ruby is ~6× slower than MRI (6.08×).** Go's scalar
  `hash/adler32` is the slowest column here; zlib's vectorized Adler-32 and even
  JRuby (0.82×) beat it. A SIMD Adler-32 is the largest single checksum gap.

The parity picture is therefore **split**: go-ruby-zlib is *faster than the C
reference on deflate* and *behind it on inflate and on the two checksums*. The
checksum gaps are hand-asm/SIMD walls (Go's standard-library kernels vs zlib's
tuned ones), which is the honest ceiling for a CGO=0 stack on this host.

!!! note "Honest framing"
    JRuby and TruffleRuby carry JIT warm-up; the 3-pass warm-up here lets them
    approach steady state but does not fully erase it, so treat their sub-`~5 µs`
    checksum rows as order-of-magnitude. `vs MRI` compares against MRI's C-zlib
    extension, not against a pure-Ruby baseline. These are **real measured
    numbers** from the 2026-07-03 run (host and versions above); best-of-25 was
    stable across repeated runs (deflate 0.14×, inflate ~2.4–2.5×, crc32
    ~3.5–3.9×, adler32 ~5.5–6.1×). Nothing is fabricated or cherry-picked.
