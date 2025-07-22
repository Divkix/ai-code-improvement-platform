// ABOUTME: Code generation directives for oapi-codegen
// ABOUTME: Generates Go types, server interfaces, and client from OpenAPI specification
package generated

//go:generate go tool oapi-codegen --config=types.yaml ../../api/openapi.yaml
//go:generate go tool oapi-codegen --config=server.yaml ../../api/openapi.yaml
//go:generate go tool oapi-codegen --config=client.yaml ../../api/openapi.yaml
