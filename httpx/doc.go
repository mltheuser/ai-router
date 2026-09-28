// Package httpx holds the HTTP plumbing shared across the router: on the
// serving side, the error type every handler responds with and JSON
// request/response helpers; on the calling side, the Client providers use to
// reach their upstream APIs, which maps upstream failures onto that same
// error type.
package httpx
