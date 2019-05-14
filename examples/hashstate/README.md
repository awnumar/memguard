# Blake2b

An example of initialising a struct over some memory allocated by memguard. We use it here to store the state of the Blake2b hash function being used as a cryptographically-secure pseudo-random number generator.

Blake2b was slightly modified to allow the caller to specify an already allocated struct instead of being handed one. Another way of doing this would be for the Blake2b package to import memguard and allocate the buffers itself, returning the Destroy method alongside the finalised hash state.
