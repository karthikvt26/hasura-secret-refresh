# hasura-secret-refresh

The Hasura action must be configured to forward the request to this server. Any
substring like '##abc##' in any of the header values would be replaced with
the secret fetched from AWS Secrets Manager corresponding to the secret id 'abc'.

eg. if this server receives a header like Authorization: Bearer ##secretToken##
this server will try to fetch the secret corresponding to secretToken from AWS
Secrets Manager. If the fetched value is secretval, then this header will be
modified to Authorization: Bearer secretval

After modifying the headers, the request will be forwarded to the hostname given
in the header `X-Hasura-Forward-Host` and the response will be returned unchanged.
The `X-Hasura-Forward-Host` header is only for use by this server and isn't
forwarded to the destination server.

## Config format

```json
{
  "cache_ttl": 300 // cache ttl in seconds
}
```
