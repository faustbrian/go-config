package awssecretsmanager_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go/tracing"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	source "github.com/faustbrian/go-config/adapters/awssecretsmanager/v3"
)

type sdkResponseBody struct {
	io.Reader
	closes int
}

func (b *sdkResponseBody) Close() error { b.closes++; return nil }

type sdkTransport struct {
	body  *sdkResponseBody
	calls int
}

func (t *sdkTransport) Do(r *http.Request) (*http.Response, error) {
	t.calls++
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/x-amz-json-1.1"}}, Body: t.body, Request: r}, nil
}

type sdkRejection struct{}

func (*sdkRejection) Error() string { return "private interceptor diagnostic" }

type sdkInterceptor struct {
	rejection *sdkRejection
	calls     int
}

func (i *sdkInterceptor) AfterTransmit(context.Context, *smithyhttp.InterceptorContext) error {
	i.calls++
	return i.rejection
}
func (i *sdkInterceptor) BeforeDeserialization(context.Context, *smithyhttp.InterceptorContext) error {
	i.calls++
	return i.rejection
}

type sdkTraceProvider struct{ names *[]string }

func (p sdkTraceProvider) Tracer(scope string, options ...tracing.TracerOption) tracing.Tracer {
	return sdkTracer{base: tracing.NopTracerProvider{}.Tracer(scope, options...), names: p.names}
}

type sdkTracer struct {
	base  tracing.Tracer
	names *[]string
}

func (t sdkTracer) StartSpan(ctx context.Context, name string, options ...tracing.SpanOption) (context.Context, tracing.Span) {
	*t.names = append(*t.names, name)
	return t.base.StartSpan(ctx, name, options...)
}

// Exercise received-response ownership through the real generated SDK client,
// including interceptors that stop deserialization before the adapter sees it.
func TestSourceClosesSDKResponsesAndKeepsOneRetryLoop(t *testing.T) {
	for _, stage := range []string{"success", "AfterTransmit", "BeforeDeserialization"} {
		t.Run(stage, func(t *testing.T) {
			body := &sdkResponseBody{Reader: strings.NewReader(`{"SecretString":"{\"workers\":4}"}`)}
			transport := &sdkTransport{body: body}
			rejection := &sdkRejection{}
			interceptor := &sdkInterceptor{rejection: rejection}
			var spans []string
			client := secretsmanager.New(secretsmanager.Options{
				Region: "eu-north-1", Credentials: aws.AnonymousCredentials{}, RetryMaxAttempts: 1,
				HTTPClient: transport, TracerProvider: sdkTraceProvider{names: &spans},
			}, func(o *secretsmanager.Options) {
				o.BaseEndpoint = aws.String("https://secretsmanager.invalid")
				switch stage {
				case "AfterTransmit":
					o.Interceptors.AddAfterTransmit(interceptor)
				case "BeforeDeserialization":
					o.Interceptors.AddBeforeDeserialization(interceptor)
				}
			})
			s, err := source.New(client, source.Options{Name: "fixture", SecretID: "fixture/secret"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			document, err := s.Load(ctx)
			if stage == "success" {
				if err != nil || !reflect.DeepEqual(document.Tree, map[string]any{"workers": int64(4)}) || !s.Info().Sensitive {
					t.Fatalf("sensitive decoded document/error = %#v/%v", document, err)
				}
			} else {
				var cause *sdkRejection
				if !errors.Is(err, source.ErrOperation) || !errors.Is(err, rejection) || !errors.As(err, &cause) || cause != rejection {
					t.Fatalf("rejection category/identity = %v", err)
				}
				if strings.Contains(err.Error(), rejection.Error()) {
					t.Fatal("private rejection exposed")
				}
				if interceptor.calls != 1 {
					t.Fatalf("interceptor calls = %d", interceptor.calls)
				}
			}
			if transport.calls != 1 {
				t.Fatalf("requests = %d", transport.calls)
			}
			if body.closes == 0 {
				t.Error("received response was not closed before Load returned")
			}
			loops := 0
			for _, name := range spans {
				if name == "RetryLoop" {
					loops++
				}
			}
			if loops != 1 {
				t.Errorf("RetryLoop spans = %d, all spans = %v", loops, spans)
			}
		})
	}
}
