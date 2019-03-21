<p align="center">
  <img src="https://cdn.rawgit.com/awnumar/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">Easy and secure handling of sensitive data, in pure Go.</p>
  <p align="center">
    <a href="https://travis-ci.org/awnumar/memguard"><img src="https://travis-ci.org/awnumar/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/awnumar/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/nrtqmdolndm0pcac/branch/master?svg=true"></a>
    <a href="https://godoc.org/github.com/awnumar/memguard"><img src="https://godoc.org/github.com/awnumar/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/awnumar/memguard"><img src="https://goreportcard.com/badge/github.com/awnumar/memguard"></a>
  </p>
</p>

---

This package attempts to reduce the likelihood of sensitive data being exposed. It supports all major operating systems and is written in pure Go.

## Features

* Sensitive data is encrypted and authenticated in memory using xSalsa20 and Poly1305 respectively. This is implemented using Go's native NaCl library.
* Memory allocation bypasses the language runtime entirely by using system calls to query the kernel for resources directly. This avoids interference from the garbage-collector.
* Buffers that store plaintext data are fortified with guard pages and canary values to detect spurious accesses and overflows.
* Effort is taken to prevent sensitive data from ever touching the disk. The data is locked to prevent swapping, system core dumps can be disabled, and the kernel is advised (where possible) to never dump secure memory.
* Kernel-level immutability is implemented. That means that if anything attempts to modify an immutable container, the kernel will throw an access violation and the process will terminate.
* It is extremely easy to wipe and destroy stored data. Multiple API endpoints expose session purging capabilities and special Panic and Exit functions allow for process termination without worrying about data being left hanging around.
* Side-channel attacks are mitigated against by making sure that the copying and comparison of data is done in constant-time.
* Accidental memory leaks are mitigated against by harnessing the garbage-collector to automatically destroy containers that have run out of scope.

Some features were inspired by [libsodium](https://github.com/jedisct1/libsodium), so credits to them.

Full documentation and a complete overview of the API can be found [here](https://godoc.org/github.com/awnumar/memguard).

## Installation

```
$ go get github.com/awnumar/memguard
```

We **strongly** encourage you to vendor your dependencies for a clean and reliable build. Go [modules](https://github.com/golang/go/wiki/Modules) make this task relatively frictionless.

## Contributing

* Reading the source code and looking for improvements.
* Developing Proof-of-Concept attacks and mitigations.
* Help with formalizing an appropriate threat model.
* Improving compatibility with more kernels and architectures.
* Implementing kernel-specific and cpu-specific protections.
* Systems to further harden [`core::Coffer`](core/coffer.go) against attack.
* Implement the catching of segmentation faults to wipe memory sectors before continuing.
* Write forks of popular crypto and security APIs using the memguard interface.
* Submit performance improvements or benchmarking code.

Issues are for reporting bugs and for discussion on proposals. Pull requests should be made against master.