package awssecretsmanager

import (
	"context"
	"testing"

	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// FuzzSourceLoad feeds arbitrary binary payloads through a stub Client and 4096-byte
// source cap; no AWS service is contacted.
func FuzzSourceLoad(f *testing.F) {
	f.Add([]byte(`{"database":{"url":"postgres://synthetic"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte("not-json"))

	f.Fuzz(func(t *testing.T, payload []byte) {
		client := &stubClient{output: &secretsmanager.GetSecretValueOutput{
			SecretBinary: append([]byte(nil), payload...),
		}}
		source, err := New(client, Options{
			Name: "fuzz", SecretID: "track/fuzz", MaximumBytes: 4_096,
		})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		_, _ = source.Load(context.Background())
	})
}
