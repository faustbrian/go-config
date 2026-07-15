# Sources and formats

Every source has a stable name, priority, sensitivity flag, and optional flag.
Inputs are copied or reopened for each load. A source never writes the process
environment or destination struct.

## Structured formats

JSON uses `encoding/json` token semantics, detects duplicate keys, preserves
integer precision through signed/unsigned conversion, rejects multiple roots,
and requires an object root. YAML uses `go.yaml.in/yaml/v4`, requires a mapping
root, rejects duplicate keys, aliases, merge keys, custom tags, non-string keys,
multiple documents, NaN, and infinity. TOML uses BurntSushi TOML, rejects
duplicate definitions, and normalizes date, local time, local date-time, and
offset date-time values to strings.

All three bound bytes, nesting depth, and key count. JSON and YAML check context
during recursive parsing/conversion; TOML checks it during normalization.
Equivalent documents intentionally normalize to `map[string]any`, `[]any`,
`int64`/`uint64`, `float64`, `bool`, `string`, and `nil`. TOML has no null.
Numeric conversion failures report only safe path/location and category data;
oversized numeric tokens are never included in diagnostics.

Use byte sources for immutable in-memory input, `FromFS` for embedded/test
filesystems, and `filesystem.FromPath` for files that must be reopened after
atomic replacement. `filesystem.Reader` requires an explicit format because a
reader has no extension.

## Dotenv

Dotenv accepts blank lines, comments, `export NAME=value`, single and double
quotes, multiline quotes, inline comments after whitespace, and documented
escapes. Duplicate variables, NUL bytes, invalid names, unsupported escapes,
trailing content, and unterminated quotes fail. Bounds cover bytes, lines, line
length, and key count.

Interpolation is disabled unless `Interpolation` is supplied. `${NAME}` and
`${NAME:-fallback}` are supported. The caller chooses external variables and
whether values from the same file participate. Expansion is depth- and
byte-bounded, cycle-detected, and cancellation-aware. Single-quoted values and
escaped dollars are literal.

## Environment

Environment mapping is schema-driven from `T`. Use `env` tags, an optional
prefix, and a nested separator (default `__`). `CaseNative` follows the host;
`CaseSensitive` and `CaseInsensitive` are explicit portable choices.
Post-normalization field collisions fail during construction.

Values decode as booleans, signed/unsigned integers, floats, durations,
timestamps, URLs, byte sizes, enums, text-unmarshaling types, JSON arrays, and
JSON objects. Collection JSON must contain exactly one value of the target
shape. Empty strings remain present and are not treated as absent.

`EnvironFor` copies a caller-provided `[]string`; use it in tests and controlled
composition roots. `ProcessFor` reads the current process environment on each
load but never changes it.

## Programmatic sources

`programmatic.Defaults` and `Overrides` select the documented low/high default
priorities. `Map` accepts an explicit priority. Maps and nested collections are
normalized to the same tree model and deeply cloned, so caller mutation cannot
change later loads. Pointer, function, channel, and arbitrary struct values are
rejected rather than retained by reference. `merge.Delete{}` is the explicit
deletion marker.

Every `Source` implementation is rechecked at the plan boundary. Cyclic maps
or slices, trees deeper than 64 levels, trees exceeding 100,000 aggregate keys,
and trees exceeding 100,000 aggregate array elements fail before merge. This
protects callers implementing custom sources in addition to the tighter
configurable limits on built-in sources.
