# kirana

Low-latency soft real-time platform for Go. Kirana provides the foundational components for system-critical applications that have galactic performance and self-healing.

### Closing the Gap to C/C++/Rust Performance

C/C++/Rust and Go all output CPU specific Assembly. Go's Assembly output is generally less optimized than GCC/LLVM C/Rust Assembly.
Go has 2 options to tap into that:
- Assembly
- CGO

CGO is too expensive for any hot path, however an Assembly based blocking trampoline cuts that cost from 30-50ns to 2-3ns.
  - github.com/moontrade/unsafe

In a general sense, we can tap into C/C++/Rust at an extremely low cost for hot paths.

### What about Go's Garbage Collector

A central theme in Kirana is eliminating and/or managing non-determinism. Although Garbage Collectors are generally deterministic based on GC managed allocation patterns, GC managed allocation patterns are generally the side-effect produced by development.

### Pools

Lock-free, GC-Free ring based queues are used rather than sync.Pool which is unbounded and node based (GC allocation for every Put). Generally, Pools are filled "with" GC allocations

### sync.Pool produces Garbage

sync.Pool is a great general purpose Pool for Go, but allocating a Node for every Put produces unnecessary garbage.

### Reactors with Nanosecond to Microsecond Latency - (NS Capable)

- Lock-free, GC free ring based queues (single-digit to low double-digit nanoseconds)
- GC free high frequency timing wheel
- GC free file-backed queues and streams

### Append Only Files (MMAP sorcery)

The package "aof" contains the primitives for nanosecond capable systems like queues and streams. AOF allows for single writers and many Readers. Readers are NEVER blocked.

### Transactional Key-Value via MDBX

## High-Frequency Trading

Kirana is designed to be fully capable of microsecond HFT.

### Why the name Kirana?

We are based in Indonesia and Kirana loosely translates to "ray of light or beam" in English.
