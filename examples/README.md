# Examples

In this directory are some useful and interesting code samples both for writing effective code using the library and for learning or exploratory purposes. Exploits and proof-of-concepts are also welcome.

Packages are able to import one another. This allows us to build up modules of functionality that work together to create more complex systems.

## Adding a package

1. Create a directory for your program and populate it with code. Check an existing module for guidance if needed. 

2. Add test code and benchmarks.

    The programs can then be run individually with

    ```bash
    go test -v -race ./examples/module_name
    ```

    or one after the other

    ```bash
    go test -v -race ./examples/...
    ```

3. Add your program to the end of the [packages](#packages) section of this document.

## Licencing

You own your intellectual property and so you are free to choose any licence for your program. To do this, add a licence header to the top of your source files.

## Packages

0. [Apache-2.0] [`socketkey`](socketkey) :: Streaming multi-threaded client->server transfer of secure data over a socket.
1. [Apache-2.0] [`casting`](casting) :: Some examples of representing the data in allocated buffers as different types.
