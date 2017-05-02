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

This is a thread-safe package, designed to allow you to easily handle sensitive values in memory. It supports all major operating systems and is written in pure Go.

## Features

* Memory is allocated using system calls, thereby bypassing the Go runtime and preventing the GC from messing with it.
* To prevent buffer overflows and underflows, the secure buffer is sandwiched between two protected guard pages. If these pages are accessed, a SIGSEGV violation is triggered.
* The secure buffer is prepended with a random canary. If this value changes, the process will panic. This is designed to prevent buffer underflows.
* All pages between the two guards are locked to stop them from being swapped to disk.
* The secure buffer can be made read-only so that any other action triggers a SIGSEGV violation.
* When freeing, secure memory is wiped.
* The API includes functions to disable system core dumps and catch signals.

Some of these features were inspired by [libsodium](https://github.com/jedisct1/libsodium), so credits to them.

Full documentation and a complete overview of the API can be found [here](https://godoc.org/github.com/libeclipse/memguard).

## Installation

Although we do recommend using a release, the simplest way to install the library is to `go get` it:

```
$ go get github.com/libeclipse/memguard
```

If you would prefer a signed release that you can verify and manually compile yourself, download and extract the [latest release](https://github.com/libeclipse/memguard/releases/latest). Then go ahead and run:

```
$ go install -v ./
```

The releases are cryptographically signed. My PGP public-key can be found on my [keybase](https://keybase.io/awn). To import it directly into GPG, run:

```
$ curl https://keybase.io/awn/pgp_keys.asc | gpg --import
```
