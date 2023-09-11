/*
Package rpc is a provider that defines an HTTP request/response contract for handling BMC interactions.
It allows users a simple way to interoperate with an existing/bespoke out-of-band management solution.

The rpc provider request/response payloads are modeled after JSON-RPC 2.0, but are not JSON-RPC 2.0
compliant so as to allow for more flexibility and interoperability with existing systems.

The rpc provider has options that can be set to include an HMAC signature in the request header.
It follows the features found at https://webhooks.fyi/security/hmac, this includes hash algorithms sha256
and sha512, replay prevention, versioning, and key rotation.
*/
package rpc
