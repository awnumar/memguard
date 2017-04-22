<p align="center">
  <img src="https://cdn.rawgit.com/libeclipse/memguard/master/logo.svg" height="140" />
  <h3 align="center">MemGuard</h3>
  <p align="center">A library that handles the secure handling of sensitive values in memory.</p>
  <p align="center">
    <a href="https://travis-ci.org/libeclipse/memguard"><img src="https://travis-ci.org/libeclipse/memguard.svg?branch=master"></a>
    <a href="https://ci.appveyor.com/project/libeclipse/memguard/branch/master"><img src="https://ci.appveyor.com/api/projects/status/g6cg347cam7lli5m/branch/master?svg=true"></a>
    <a href="https://dependencyci.com/github/libeclipse/memguard"><img src="https://dependencyci.com/github/libeclipse/memguard/badge"></a>
    <a href="https://godoc.org/github.com/libeclipse/memguard"><img src="https://godoc.org/github.com/libeclipse/memguard?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/libeclipse/memguard"><img src="https://goreportcard.com/badge/github.com/libeclipse/memguard"></a>
  </p>
</p>

---

This library is designed to allow you to easily handle sensitive values in memory. The main functionality is to lock and watch portions of memory and wipe them on exit, but there are some supplementary function too.

Multiple operating systems are supported, using `mlock` on Unix and `VirtualLock` on Windows.
