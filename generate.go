// Package sesc_test only provides tools to generate some code, like the swagger schema.
// It is named _test to not include it in the binary.
//
//go:generate swag init -g ./cmd/api/main.go -o ./api/docs
package sesc_test
