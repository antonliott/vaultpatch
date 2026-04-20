// Package reorder implements key reordering for Vault KV secrets.
//
// Keys can be sorted alphabetically, in reverse alphabetical order,
// or according to a caller-supplied explicit ordering list. Keys not
// present in an explicit list are appended after the ordered subset,
// sorted alphabetically.
//
// The feature is useful when teams have agreed conventions for key
// layout (e.g. credentials before metadata) and want to enforce them
// consistently across all secrets in a namespace.
package reorder
