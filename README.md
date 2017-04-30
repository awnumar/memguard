<p align="center">
  <img src="https://cdn.rawgit.com/libeclipse/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard (beta)</h3>
  <p align="center">A library that handles sensitive values in memory.</p>
  <p align="center">
    <a href="https://travis-ci.org/libeclipse/memguard"><img src="https://travis-ci.org/libeclipse/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/libeclipse/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/g6cg347cam7lli5m/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/libeclipse/memguard"><img src="https://godoc.org/github.com/libeclipse/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/libeclipse/memguard"><img src="https://goreportcard.com/badge/github.com/libeclipse/memguard"></a>
  </p>
</p>

---

This library is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary functions too.

**Currently the package is in beta stages and not yet ready for use in production. As such, you should NOT use it seriously. If you would like to contribute, feel free to open a pull request.**

## Installation

This library can be retrieved with `go get`.

`$ go get github.com/libeclipse/memguard`

## How it works

You request a protected buffer, and MemGuard allocates it as follows:

| Guard Page | Padding bytes | Canary | Requested Buffer | Guard Page |
|:----------:|:-------------:|:------:|:----------------:|:----------:|
| No Access | No SWAP | No SWAP | No SWAP | No Access|

If the guard pages are accessed, the process will crash with a `SIGSEGV`. If the canary is found to have been edited, the process will panic. If anything goes wrong with protected memory, the process will also panic.

This setup protects against both buffer overflows and underflows, and the swapping of sensitive memory to disk. 

For a complete API reference, refer to [godoc](https://godoc.org/github.com/libeclipse/memguard).
