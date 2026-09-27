# Choosing a Go OpenAPI Approach

`go-api-docs` is a small code-first router. It is useful when a service wants
stdlib handler semantics, runtime OpenAPI 3.1 generation, and optionally a
source-only static snapshot.

It is not intended to replace every OpenAPI workflow.

| Project | Direction | Good fit |
|---------|-----------|----------|
| `go-api-docs` | Go types and routes to OpenAPI | Small stdlib services and source-only snapshots |
| [Huma](https://huma.rocks/) | Go types and routes to OpenAPI | Mature validation, negotiation, transforms, and framework adapters |
| [Fuego](https://github.com/go-fuego/fuego) | Go handlers to OpenAPI | Applications that want a broader web framework |
| [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) | OpenAPI to Go | Contract-first APIs and generated clients or servers |
| [ogen](https://github.com/ogen-go/ogen) | OpenAPI to Go | Strict generated clients and servers |
| [swaggo/swag](https://github.com/swaggo/swag) | Go annotations to OpenAPI | Existing annotation-based projects |

## Trade-offs

Choose `go-api-docs` when all of these are true:

- The API can use the package's typed handler signatures and fixed JSON response model.
- Runtime reflection is acceptable, or the source fits `goapi-gen`'s documented subset.
- Schema constraint tags are documentation unless the application implements `Validator`.
- A compact API matters more than broad OpenAPI feature coverage.

Choose Huma when runtime validation, response headers, content negotiation,
schema composition, and established framework adapters matter more than a
stdlib-only runtime dependency graph.

Choose `oapi-codegen` or ogen when the OpenAPI document is the reviewed contract
and Go code should be generated from it.

Choose swaggo when an existing application already uses its comment annotations
and ecosystem tooling.
