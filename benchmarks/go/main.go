// SPDX-License-Identifier: BSD-3-Clause
//
// Zlib workload driver. Builds a deterministic ~64 KiB "mixed text / repetitive"
// buffer (byte-identical to the one ruby/zlib.rb builds), verifies deflate then
// inflate round-trips to the original BEFORE any timing, then times deflate,
// inflate, CRC-32 and Adler-32 through the pure-Go go-ruby-zlib API.
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-ruby-zlib/zlib"
)

// phrase is the repeating English text block; the buffer is mostly this (so it
// compresses like real text), perturbed every 97th byte with a deterministic
// printable byte so the stream is not trivially compressible. ruby/zlib.rb uses
// the identical phrase and the identical construction, so both sides feed the
// exact same bytes to every runtime.
const phrase = "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. "

const bufSize = 64 * 1024

func mixedBuf() []byte {
	p := []byte(phrase)
	b := make([]byte, bufSize)
	for i := range b {
		if i%97 == 0 {
			b[i] = byte(0x20 + ((i*31 + 7) & 0x3F)) // deterministic printable perturbation
		} else {
			b[i] = p[i%len(p)]
		}
	}
	return b
}

const level = 6 // fixed compression level, matching ruby/zlib.rb

func main() {
	buf := mixedBuf()

	// Correctness gate: deflate then inflate must reproduce the input exactly,
	// otherwise the timing below would be meaningless.
	comp, err := zlib.Deflate(buf, level)
	if err != nil {
		fmt.Fprintln(os.Stderr, "deflate error:", err)
		os.Exit(1)
	}
	rt, err := zlib.Inflate(comp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "inflate error:", err)
		os.Exit(1)
	}
	if !bytes.Equal(rt, buf) {
		fmt.Fprintln(os.Stderr, "round-trip mismatch: inflate(deflate(buf)) != buf")
		os.Exit(1)
	}
	// Emitted to stderr so run.sh's RESULT parser ignores it; lets a human confirm
	// the Go and Ruby sides build the identical buffer (same CRC) and see the
	// compressed size (Go flate vs C zlib framing differ, payload is identical).
	fmt.Fprintf(os.Stderr, "go: buf=%d B crc32=%08x adler32=%08x deflate(L%d)=%d B round-trip=OK\n",
		len(buf), zlib.Crc32(buf, 0), zlib.Adler32(buf, 1), level, len(comp))

	bench("deflate-64KiB-L6", 100, func() { c, _ := zlib.Deflate(buf, level); sink = c })
	bench("inflate-64KiB-L6", 100, func() { d, _ := zlib.Inflate(comp); sink = d })
	bench("crc32-64KiB", 500, func() { sink = zlib.Crc32(buf, 0) })
	bench("adler32-64KiB", 500, func() { sink = zlib.Adler32(buf, 1) })
}
