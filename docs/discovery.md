# Explicit discovery

`discover.Search` only examines paths supplied through `Explicit`,
`Directories`, `StartDir`, or an explicitly enabled user-config directory. It
does not traverse parents or home directories by default and never reads file
contents.

`SearchPlaces` is an ordered list of safe filenames, not paths. `SearchFirst`
stops at the first regular file; `SearchAll` preserves explicit directory and
place order. Candidate, result, and upward-depth limits are always enforced.

Upward search requires `Upward: true`, an explicit `StartDir`, and an explicit
ancestor `StopDir`. A `Root` lexically contains candidates and also contains
resolved symlink targets. Symlinks are rejected by default; `AllowWithinRoot`
permits only targets still inside the resolved root. `OwnerOnly` rejects files
with group or other permission bits.

```go
results, err := discover.Search(ctx, discover.Options{
	Root:         repositoryRoot,
	StartDir:     workingDirectory,
	StopDir:      repositoryRoot,
	Upward:       true,
	SearchPlaces: []string{"app.yaml", "app.json", "app.toml"},
	Mode:         discover.SearchFirst,
})
```

Pass a result to `filesystem.FromDiscovered`. The file is opened through its
canonical resolved target while provenance records the lexical discovered path.
Opening and parsing remain separate from discovery, so malformed and unreadable
files are never mistaken for absence.

Paths in errors are policy diagnostics and may reveal deployment layout. Do not
put secret material in filenames, directories, application names, or search
places. Discovery does not protect against an attacker who can replace files
inside an otherwise trusted root; use deployment ownership and read-only mounts.
