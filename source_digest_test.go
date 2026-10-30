package config_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"testing"
	"testing/fstest"

	config "github.com/faustbrian/go-config/v3"
	"github.com/faustbrian/go-config/v3/dotenv"
	"github.com/faustbrian/go-config/v3/filesystem"
	jsonsource "github.com/faustbrian/go-config/v3/json"
	"github.com/faustbrian/go-config/v3/programmatic"
	tomlsource "github.com/faustbrian/go-config/v3/toml"
	yamlsource "github.com/faustbrian/go-config/v3/yaml"
)

type digestSettings struct {
	Name string `config:"name" env:"NAME"`
}

func TestByteSourcesReportTheDigestOfTheExactAcceptedBytes(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		data  string
		build func([]byte) (config.Source, error)
	}{
		{"json", "{\"name\": \"json\"}\n", func(data []byte) (config.Source, error) {
			return jsonsource.Bytes(data, jsonsource.Options{Name: "json"})
		}},
		{"yaml", "name: yaml\n", func(data []byte) (config.Source, error) {
			return yamlsource.Bytes(data, yamlsource.Options{Name: "yaml"})
		}},
		{"empty-yaml", "  \n", func(data []byte) (config.Source, error) {
			return yamlsource.Bytes(data, yamlsource.Options{Name: "empty-yaml"})
		}},
		{"toml", "name = \"toml\"\n", func(data []byte) (config.Source, error) {
			return tomlsource.Bytes(data, tomlsource.Options{Name: "toml"})
		}},
		{"dotenv", "NAME=dotenv\n", func(data []byte) (config.Source, error) {
			return dotenv.BytesFor[digestSettings](data, dotenv.Options{Name: "dotenv"})
		}},
		{"reader", "{\"name\":\"reader\"}", func(data []byte) (config.Source, error) {
			return filesystem.Reader(func(context.Context) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			}, filesystem.Options{Name: "reader", Format: filesystem.FormatJSON})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			source, err := test.build([]byte(test.data))
			if err != nil {
				t.Fatalf("construct source: %v", err)
			}
			document, err := source.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			want := config.SourceDigest{SHA256: sha256.Sum256([]byte(test.data))}
			if document.Digest == nil || *document.Digest != want {
				t.Fatalf("Document.Digest = %v, want SHA-256 of the accepted bytes", document.Digest)
			}
		})
	}
}

func TestRejectedBytesReportNoDocumentOrDigest(t *testing.T) {
	t.Parallel()

	source, err := jsonsource.Bytes([]byte(`{"name":`), jsonsource.Options{Name: "json"})
	if err != nil {
		t.Fatal(err)
	}
	document, err := source.Load(context.Background())
	if err == nil || document.Digest != nil {
		t.Fatalf("Load() = (%v, %v), want a parse error without a digest", document.Digest, err)
	}
}

func TestSnapshotRecordsEachLoadedFileSourceDigestByName(t *testing.T) {
	t.Parallel()

	base := []byte(`{"name":"base"}`)
	override := []byte("name: override\n")
	files := fstest.MapFS{
		"base.json":     {Data: base},
		"override.yaml": {Data: override},
	}
	baseSource, err := filesystem.FromFS(files, "base.json", filesystem.Options{Name: "base", Priority: 10})
	if err != nil {
		t.Fatal(err)
	}
	overrideSource, err := filesystem.FromFS(files, "override.yaml", filesystem.Options{Name: "override", Priority: 20})
	if err != nil {
		t.Fatal(err)
	}
	missing, err := filesystem.FromFS(files, "missing.json", filesystem.Options{Name: "missing", Priority: 30, Optional: true})
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := programmatic.Defaults("defaults", map[string]any{"name": "default"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := config.NewPlan(defaults, baseSource, overrideSource, missing)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := config.Load[digestSettings](context.Background(), plan)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if snapshot.Value().Name != "override" {
		t.Fatalf("Load() name = %q", snapshot.Value().Name)
	}
	for name, data := range map[string][]byte{"base": base, "override": override} {
		digest, ok := snapshot.SourceDigest(name)
		if !ok || digest.SHA256 != sha256.Sum256(data) {
			t.Fatalf("SourceDigest(%q) = (%x, %t), want the loaded bytes even when overridden", name, digest.SHA256, ok)
		}
	}
	for _, name := range []string{"defaults", "missing", "unknown"} {
		if digest, ok := snapshot.SourceDigest(name); ok {
			t.Fatalf("SourceDigest(%q) = %x, want none for a byte-free, skipped or unplanned source", name, digest.SHA256)
		}
	}

	tree, err := config.LoadTree(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if digest, ok := tree.SourceDigest("base"); !ok || digest.SHA256 != sha256.Sum256(base) {
		t.Fatalf("LoadTree SourceDigest(base) = (%x, %t)", digest.SHA256, ok)
	}
}

func TestEachLoadDigestsTheBytesThatLoadAccepted(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{"settings.json": {Data: []byte(`{"name":"first"}`)}}
	source, err := filesystem.FromFS(files, "settings.json", filesystem.Options{Name: "settings"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := config.NewPlan(source)
	if err != nil {
		t.Fatal(err)
	}
	first, err := config.Load[digestSettings](context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	files["settings.json"] = &fstest.MapFile{Data: []byte(`{"name":"second"}`)}
	second, err := config.Load[digestSettings](context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	firstDigest, _ := first.SourceDigest("settings")
	secondDigest, _ := second.SourceDigest("settings")
	if firstDigest.SHA256 != sha256.Sum256([]byte(`{"name":"first"}`)) ||
		secondDigest.SHA256 != sha256.Sum256([]byte(`{"name":"second"}`)) {
		t.Fatal("each snapshot must keep the digest of the bytes its own load accepted")
	}

	files["settings.json"] = &fstest.MapFile{Data: []byte(`{"name":`)}
	if _, err := config.Load[digestSettings](context.Background(), plan); err == nil {
		t.Fatal("a rejected re-read must not publish a snapshot")
	}
	if digest, _ := first.SourceDigest("settings"); digest != firstDigest {
		t.Fatal("a later load changed an earlier snapshot's digest")
	}
}

func TestZeroSnapshotHasNoSourceDigest(t *testing.T) {
	t.Parallel()

	var snapshot config.Snapshot[digestSettings]
	if _, ok := snapshot.SourceDigest("settings"); ok {
		t.Fatal("zero Snapshot reported a digest")
	}
}

func TestSourceDigestFromACustomSourceIsCopied(t *testing.T) {
	t.Parallel()

	digest := &config.SourceDigest{SHA256: sha256.Sum256([]byte("custom"))}
	plan, err := config.NewPlan(customDigestSource{digest: digest})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := config.LoadTree(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	want := *digest
	digest.SHA256[0] ^= 0xff
	if got, ok := snapshot.SourceDigest("custom"); !ok || got != want {
		t.Fatal("snapshot digest aliases the source's Document.Digest")
	}
}

type customDigestSource struct{ digest *config.SourceDigest }

func (customDigestSource) Info() config.SourceInfo { return config.SourceInfo{Name: "custom"} }

func (s customDigestSource) Load(context.Context) (config.Document, error) {
	return config.Document{Tree: map[string]any{}, Digest: s.digest}, nil
}
