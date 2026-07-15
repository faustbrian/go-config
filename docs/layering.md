# Layering, defaults, merging, interpolation, and validation

## Default precedence

`NewDefaultPlan` applies these categories from lowest to highest:

1. typed/programmatic defaults
2. discovered base files
3. discovered profile files
4. explicit files in caller order
5. dotenv
6. process environment
7. explicit programmatic overrides

Order within a category is preserved. `NewPlan` supports a custom order by
sorting stable source priorities. Duplicate names, empty names, invalid
metadata, and nil sources fail before loading. Inspect `Plan.Sources()` before
load when startup diagnostics need to show the resolved order.

## Defaults and presence

`defaults.For[T]` reads `default` tags using the destination field type. Scalar
text hooks, durations, URLs, `ByteSize`, `Optional[T]`, JSON slices, and JSON
maps are supported. Invalid or trailing data fails without printing the default
text. Defaults are ordinary lowest-precedence sources.

Use `Optional[T]` when the application must distinguish:

- `Absent`: no winning source supplied the field;
- `Null`: a source explicitly supplied null;
- `Present`: a source supplied a value, including zero or empty;
- `Defaulted`: the winning value came from defaults.

Pointers still express nullable object ownership, while `Optional[T]` expresses
field presence. Do not infer presence from a Go zero value.

## Merge truth table

| Lower | Upper | Result |
|---|---|---|
| object | object | recursive key merge |
| scalar | same scalar kind | upper replacement |
| slice | slice | complete upper replacement |
| any | null | explicit null |
| any | `merge.Delete{}` | key removal |
| incompatible non-null kinds | any | `TypeConflictError` |

Slices never append or merge by index. Maps/objects merge recursively. A scalar
or slice replacing an object removes descendant provenance. A failure in any
layer discards the complete candidate snapshot.

## Strict decode and validation

`decode.Into` rejects unknown fields, missing required fields, ambiguous tags,
overflows, unsupported destinations, and conversion errors. Independent field
errors are sorted lexically and annotated with the nearest source/location
origin. Received descriptions identify a safe type category, never a value.

After complete decoding, `LoadWithValidators` runs `Validate() error` on the
candidate when implemented and then caller validators in registration order.
Use `validation.At("server.port", err)` for a safe path. Panics are recovered as
typed panic errors without retaining panic values. No failed candidate snapshot
is returned.

Interpolation belongs to the dotenv source and runs before environment mapping.
It does not inspect arbitrary process environment unless the caller explicitly
copies selected variables into `Interpolation.Variables`.
