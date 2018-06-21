<p align="center">
  <img src="https://cdn.rawgit.com/awnumar/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">Easy and secure handling of sensitive memory, in pure Go.</p>
  <p align="center">
    <a href="https://travis-ci.org/awnumar/memguard"><img src="https://travis-ci.org/awnumar/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/awnumar/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/nrtqmdolndm0pcac/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/awnumar/memguard"><img src="https://godoc.org/github.com/awnumar/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/awnumar/memguard"><img src="https://goreportcard.com/badge/github.com/awnumar/memguard"></a>
  </p>
</p>

---

This is a thread-safe package designed to allow you to easily and securely handle sensitive data in memory. It supports all major operating systems and is written in pure Go.

## Features

* All sensitive data is encrypted and authenticated in RAM using xSalsa20 and Poly1305 respectively. This is implemented with Go's native NaCl library.
* Sensitive internal values (like encryption keys) are split and stored in multiple locations in memory and are regularly re-keyed and regenerated.
* Instead of asking the Go runtime to allocate memory for us, we bypass it entirely and use system-calls to ask the kernel directly. This blocks interference from the garbage-collector.
* It is difficult for another process to find or access sensitive memory as the data is sandwiched between guard-pages. This feature also acts as an immediate access alarm in case of buffer overflows.
* Buffer overflows are further protected against by using a random canary value. If this value changes, the process will panic.
* We try our best to prevent the system from writing anything sensitive to the disk. The data is locked to prevent swapping, system core dumps can be disabled, and the kernel is advised (where possible) to never dump secure memory.
* True (kernel-level) immutability is implemented. That means that if _anything_ attempts to modify an immutable container, the kernel will throw an access violation and the process will terminate.
* All sensitive data is wiped before the associated memory is released back to the operating system.
* Side-channel attacks are mitigated against by making sure that the copying and comparison of data is done in constant-time.
* Accidental memory leaks are mitigated against by harnessing Go's own garbage-collector to automatically destroy containers that have run out of scope.

Some features were inspired by [libsodium](https://github.com/jedisct1/libsodium), so credits to them.

Full documentation and a complete overview of the API can be found [here](https://godoc.org/github.com/awnumar/memguard).

## Installation

Although we do recommend using a release, the simplest way to install the library is to `go get` it:

```
$ go get github.com/awnumar/memguard
```

If you would prefer a signed release that you can verify and manually compile yourself, download and extract the [latest release](https://github.com/awnumar/memguard/releases/latest). Then go ahead and run:

```
$ go install -v ./
```

The [latest release](https://github.com/awnumar/memguard/releases/latest) is guaranteed to be cryptographically signed with [my most recent PGP key](https://cryptolosophy.org/assets/pgp/public_key.txt). To import it directly into GPG, run:

```
$ curl https://cryptolosophy.org/assets/pgp/public_key.txt | gpg --import
```

We **strongly** encourage you to vendor your dependencies for a clean and reliable build. Go's [dep](https://github.com/golang/dep) makes this task relatively frictionless.
