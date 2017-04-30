<p align="center">
  <img src="https://cdn.rawgit.com/libeclipse/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">A pure Go library that handles sensitive values in memory.</p>
  <p align="center">
    <a href="https://travis-ci.org/libeclipse/memguard"><img src="https://travis-ci.org/libeclipse/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/libeclipse/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/g6cg347cam7lli5m/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/libeclipse/memguard"><img src="https://godoc.org/github.com/libeclipse/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/libeclipse/memguard"><img src="https://goreportcard.com/badge/github.com/libeclipse/memguard"></a>
  </p>
</p>

---

This library is designed to allow you to easily handle sensitive values in memory. It supports all major operating systems and is written in pure Go.

## Features

* Memory is allocated using system calls, thereby bypassing the Go runtime and preventing the GC from messing with it.
* To prevent buffer overflows and underflows, the secure buffer is sandwiched between two protected guard pages. If these pages are accessed, a SIGSEGV violation is triggered.
* The secure buffer is prepended with a random canary. If this value changes, the process will panic. This is designed to prevent buffer underflows.
* All pages between the two guards are locked to stop them from being swapped to disk.
* The secure buffer can be made read or write-only so that any other action triggers a SIGSEGV violation.
* When freeing, secure memory is wiped.
* The API includes functions to disable system core dumps and catch signals.

Some of these features were inspired by [libsodium](https://github.com/jedisct1/libsodium), so credits to them.

Full documentation and a complete overview of the API can be found [here](https://godoc.org/github.com/libeclipse/memguard).

## Installation

`$ go get github.com/libeclipse/memguard`
