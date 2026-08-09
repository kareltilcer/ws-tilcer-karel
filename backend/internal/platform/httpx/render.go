// Package httpx holds the HTTP layer shared across modules: JSON rendering, the
// shared Error shape, middleware (request id, logging, recovery), the health
// probes, and the SPA fallback.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// maxBodyBytes caps JSON request bodies to a sane default. Endpoints with their
// own explicit cap (e.g. contact) enforce it before this.
const maxBodyBytes = 1 << 20 // 1 MiB

// JSON writes v as a JSON response with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// DecodeJSON strictly decodes a JSON request body into dst (unknown fields are
// rejected). Returns a client-facing error suitable for a 422.
func DecodeJSON(r *http.Request, dst any) error {
	return DecodeJSONLimit(r, dst, maxBodyBytes)
}

// DecodeJSONLimit is DecodeJSON with an explicit byte cap (contact/score use a
// configurable cap enforced before decode).
func DecodeJSONLimit(r *http.Request, dst any, limit int64) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, limit))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("unexpected trailing content in request body")
	}
	return nil
}
