# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
#
# Reference Ruby zlib workload — the same operations the Go driver runs, over the
# byte-identical ~64 KiB "mixed text / repetitive" buffer. Round-trip correctness
# is asserted before any timing.
require "zlib"
require_relative "_harness"

# Identical phrase + construction to benchmarks/go/main.go; both sides therefore
# feed every runtime the exact same bytes.
PHRASE  = "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. "
BUFSIZE = 64 * 1024
LEVEL   = 6

def mixed_buf
  p = PHRASE.b
  plen = p.bytesize
  bytes = Array.new(BUFSIZE)
  BUFSIZE.times do |i|
    bytes[i] = i % 97 == 0 ? (0x20 + ((i * 31 + 7) & 0x3F)) : p.getbyte(i % plen)
  end
  bytes.pack("C*")
end

buf  = mixed_buf
comp = Zlib::Deflate.deflate(buf, LEVEL)
rt   = Zlib::Inflate.inflate(comp)
raise "round-trip mismatch: inflate(deflate(buf)) != buf" unless rt == buf
warn format("ruby: buf=%d B crc32=%08x adler32=%08x deflate(L%d)=%d B round-trip=OK",
            buf.bytesize, Zlib.crc32(buf), Zlib.adler32(buf), LEVEL, comp.bytesize)

bench("deflate-64KiB-L6", 100) { Zlib::Deflate.deflate(buf, LEVEL) }
bench("inflate-64KiB-L6", 100) { Zlib::Inflate.inflate(comp) }
bench("crc32-64KiB",      500) { Zlib.crc32(buf) }
bench("adler32-64KiB",    500) { Zlib.adler32(buf) }
