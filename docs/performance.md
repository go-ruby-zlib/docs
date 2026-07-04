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
pure Go (`github.com/klauspost/compress` for DEFLATE, the hand-written arm64 PMULL
carryless-multiply kernel of `go-simd/crc32` for CRC-32, and `go-simd/adler32`
for Adler-32) — so
this is a pure-Go stack measured head-to-head against a C library, and we report
the real numbers either way.

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
| **go-ruby (pure Go)** | 1275.2 | 0.90× |
| MRI | 1412.0 | 1.00× |
| MRI + YJIT | 1552.0 | 1.10× |
| JRuby | 6236.3 | 4.42× |
| TruffleRuby | 2299.9 | 1.63× |

*(2026-07-03: was 5443.8 / 3.86× before the SIMD kernel — a 4.3× speed-up that
now beats MRI's C zlib and YJIT. See below.)*

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
- **CRC-32 — go-ruby is now ~10% *faster* than MRI (0.90×) and beats YJIT
  (0.82× of YJIT).** The library folds CRC-32 with a hand-written arm64 **PMULL
  fold-by-eight** carryless-multiply kernel ([`go-simd/crc32`](https://github.com/go-simd/crc32)), bit-identical
  to `hash/crc32` but ~4.3× faster than it here: Go's standard-library IEEE path
  on arm64 is a **latency-bound serial `CRC32X`** instruction, whereas eight
  independent 128-bit accumulators saturate the M-series PMULL units and clear
  both the C-zlib reference (MRI 1412 ns) and YJIT (1552 ns). This flips CRC-32
  from the second-worst checksum column to a win.
- **Adler-32 — still ~6× slower than MRI (6.08×), unchanged on this host.** The
  library now routes Adler-32 through [`go-simd/adler32`](https://github.com/go-simd/adler32),
  a bit-exact SIMD drop-in — but its NEON kernel needs the integer widening
  multiply `VUMULL`, which Go's arm64 assembler only exposes in **Go 1.27+**; on
  the stable Go 1.26.4 used here that path compiles to the scalar fallback, so the
  measured number does not move. `go-simd/adler32` *does* accelerate Adler-32 on
  **amd64 (SSE/AVX2)**, riscv64, ppc64le and s390x, so those arches gain now; the
  arm64 win lands automatically on Go 1.27 (or a 1.26-compatible multiply-free
  NEON kernel — tracked as follow-up). This is the **honest floor**: without an
  integer NEON multiply on stable-Go arm64, a 6× SIMD gap over zlib's vectorized
  Adler-32 cannot be closed here.

The parity picture is therefore **improved**: go-ruby-zlib is now *faster than the
C reference on deflate **and** CRC-32*, and behind it on inflate and Adler-32. The
CRC-32 gap is closed with a carryless-multiply kernel; the remaining Adler-32 gap
is a stable-Go arm64 toolchain wall (no integer NEON multiply until Go 1.27), not
an algorithmic one.

!!! note "Per-architecture"
    The CRC-32 PMULL kernel is **arm64-only**: on amd64 the Go standard library's
    IEEE path is *already* a fold-by-four (and AVX512 fold-by-sixteen) PCLMULQDQ
    kernel, so the library defers to it there (parity, already at C-zlib class);
    ppc64le/s390x likewise use their hardware-assisted standard-library path. The
    hand kernel targets arm64 precisely because its standard-library path is the
    serial `CRC32X`. The reusable fold now lives in the standalone
    [`go-simd/crc32`](https://github.com/go-simd/crc32) (a bit-exact `hash/crc32`
    drop-in), which the library consumes directly.

!!! note "Honest framing"
    JRuby and TruffleRuby carry JIT warm-up; the 3-pass warm-up here lets them
    approach steady state but does not fully erase it, so treat their sub-`~5 µs`
    checksum rows as order-of-magnitude. `vs MRI` compares against MRI's C-zlib
    extension, not against a pure-Ruby baseline. These are **real measured
    numbers** from the 2026-07-03 run (host and versions above); best-of-25 was
    stable across repeated runs (deflate 0.14×, inflate ~2.4–2.5×, crc32
    ~0.88–0.92×, adler32 ~5.5–6.3×). Nothing is fabricated or cherry-picked.
