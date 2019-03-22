# Examples

In this directory are some useful and interesting code samples both for writing effective code using the library and for learning or exploratory purposes.

Exploits can also go in this directory. All programs exist in the same `examples` package and so share that inner namespace. This allows us to build up modules of functionality that can interact with each other.

## Adding a program

1. Write the spirit of your program within a single file, say `skit.go`. The package should be set to `examples`.
2. Add a test function to `skit_test.go`.

    The programs can be run with

    ```bash
    go test -v -race ./examples/skit_test.go ./examples/skit.go
    ```

3. Optionally add benchmarks under your test function.
4. Add a comment above your test function, substituting your chosen licence code.

    ```go
    /* [Apache-2.0] Skit :: Jon Snow <jon@beyondthe.wall> */
    ```

## Licencing

You own your intellectual property. As such you are free to choose any licence for your program. To do this, add a licence header to the top of your file and include a licence comment above your test function.

## Programs

0. [Apache-2.0] [`socketkey.go`](socketkey.go) :: Streaming multi-threaded client->server transfer of secure data over a socket.