# Runtime dependencies
Runtime dependencies are identified when an `arg` or `param` definition references either a `with`, `input` or `param` block

They are populated on the resource as part of the argument decoding (this is handled by `QueryProviderBase`)

When constructing `LeafRun` objects for resources, the `LeafRun` runtime dependencies are populated from the resource
in `resolveRuntimeDependencies`



CHANGES
- only top level nodes can have param or with
- for query providers, base does not inherit with, params or args. Instead store a reference to the base,
- only execute with runs trhat are needed by runtime dep
- in leaf run, if resource has a base and its with are required, resolve runtime depos to populate args/params on base object
