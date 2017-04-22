libsodium-go
============
A binding library made in Go for the popular portable cryptography library [Sodium](https://download.libsodium.org/doc/).


Purpose
-------
The goal of this binding library is to make use of Sodium in a more Go friendly matter.  And of course making it easier to make secure software.

Team (as of now...)
----------------
<ul>
<li>Stephen Chavez (@redragonx)</li>
<li>Graham Smith (@neuegram)</l>
</ul>

How to build
------------
For linux, this should be easy since there's pkg-config support. Please make sure libsodium is installed on your system first.

1. `go get -d github.com/GoKillers/libsodium-go`
2. `cd $GOPATH/src/github.com/GoKillers/libsodium-go`
3. `./build.sh`

For Windows, we need help here. Do a pull request if you know how to compile this on Windows.

License
---------
Copyright 2015 - GoKillers
