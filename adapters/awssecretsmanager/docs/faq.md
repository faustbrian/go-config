# FAQ

## Does this adapter discover AWS credentials?

No. Construct the SDK client from `awsconfig.LoadDefaultConfig`; the caller
therefore controls workload identity and the default credential chain.

## Does it reload rotated secrets?

No. A `Load` that is not already cancelled invokes the supplied client at most
once, and the caller owns whether and when another load is safe. The client,
not the adapter, owns any network operations or retries behind that invocation.

## Can environment variables override the secret?

Yes. Put the AWS source at discovered-profile priority and the environment
source at environment priority in the same plan.

## Can the secret contain dotenv text?

No. The payload must be one JSON object so structure, duplicate keys, bounds,
and typed decoding remain deterministic.
