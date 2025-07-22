//go:build tools
// +build tools

// ABOUTME: Tool dependencies management file for oapi-codegen
// ABOUTME: Ensures oapi-codegen is available as a tool dependency
package main

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)