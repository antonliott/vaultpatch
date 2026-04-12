// Package template provides Go text/template rendering for Vault secret values.
//
// Use NewRenderer to obtain a Renderer, then call Render or RenderAll to
// substitute secret data into configuration templates before writing them
// to disk or passing them to downstream services.
//
// Built-in template functions:
//
//	required key data  – returns the value for key or errors if absent/empty.
//	default  key fallback data – returns the value for key, or fallback.
package template
