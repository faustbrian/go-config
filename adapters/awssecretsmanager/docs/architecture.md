# Architecture

The adapter is a leaf module between the AWS SDK and
`github.com/faustbrian/go-config`. It owns no credentials, client, retry loop,
cache, refresh goroutine, or global state.

Each `Load` checks cancellation first. It invokes the supplied
`Client.GetSecretValue` at most once, or zero times when cancellation is already
observable. The client owns whether its invocation performs network operations
or retries. After a successful invocation, the adapter copies and bounds
exactly one payload, parses it with the strict
`github.com/faustbrian/go-config/json` source, and returns a sensitive document.
AWS SDK configuration and client lifetime remain composition-root
responsibilities. Typed decoding and source precedence remain responsibilities
of the enclosing `config.Plan`.
