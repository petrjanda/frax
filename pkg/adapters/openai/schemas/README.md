# OpenAI Schemas Package

The `schemas` package provides OpenAI-compatible JSON schema generation from Go structs using the [invopop/jsonschema](https://github.com/invopop/jsonschema) library. This package is specifically designed to work with the OpenAI adapter and ensures schemas are compatible with OpenAI's tool system.

## Features

- **OpenAI Compatibility**: Generates schemas that work seamlessly with OpenAI's tool system
- **Automatic Validation**: Ensures schemas only use features supported by OpenAI
- **Type Safety**: Leverages Go's type system for robust schema generation
- **Customizable**: Configurable options for different use cases

## Usage

### Basic Usage

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/petrjanda/frax/pkg/adapters/openai/schemas"
)

// Define your struct with jsonschema tags
type Person struct {
    Name     string `json:"name" jsonschema:"required"`
    Age      int    `json:"age" jsonschema:"required"`
    Email    string `json:"email" jsonschema:"required"`
    Location string `json:"location" jsonschema:"required"`
}

func main() {
    // Create a schema generator
    generator := schemas.NewOpenAISchemaGenerator()
    
    // Generate schema from a struct instance
    schema, err := generator.GenerateSchema(Person{})
    if err != nil {
        panic(err)
    }
    
    // Use the schema with OpenAI tools
    fmt.Printf("Generated schema: %s\n", string(schema))
}
```

### Using Generic Function

```go
// Generate schema directly from struct type
schema, err := schemas.GenerateSchemaFromStruct[Person]()
if err != nil {
    panic(err)
}
```

### Schema Validation

```go
// Validate that a schema is OpenAI compatible
err := generator.ValidateOpenAICompatibility(schema)
if err != nil {
    fmt.Printf("Schema compatibility warning: %v\n", err)
}
```

## OpenAI Compatibility Features

The package automatically ensures schemas are compatible with OpenAI's tool system by:

- **Expanding Structs**: Uses `ExpandedStruct: true` to avoid `$defs` references
- **No References**: Sets `DoNotReference: true` to avoid `$ref` usage
- **Required Fields**: Automatically marks fields with `jsonschema:"required"` as required
- **Validation**: Provides methods to check for unsupported features

## Supported JSON Schema Features

The generated schemas support these OpenAI-compatible features:

- ✅ Basic types: `string`, `integer`, `number`, `boolean`
- ✅ Arrays with item schemas
- ✅ Objects with property schemas
- ✅ Required field validation
- ✅ Nested structs (expanded inline)
- ✅ Enums and const values
- ✅ String formats and patterns
- ✅ Numeric constraints (min/max, etc.)

## Unsupported Features

These features are intentionally avoided for OpenAI compatibility:

- ❌ `$defs` (definitions)
- ❌ `$ref` (references)
- ❌ `additionalProperties` (not well supported)
- ❌ Complex conditional schemas

## Configuration Options

The `OpenAISchemaGenerator` can be customized:

```go
generator := &schemas.OpenAISchemaGenerator{
    // Custom configuration options can be added here
}
```

## Examples

See the `examples/structured_output/` directory for a complete working example that demonstrates:

- Schema generation from Go structs
- Integration with the LLM package
- OpenAI tool usage
- Schema validation

## Dependencies

- `github.com/invopop/jsonschema` - Core schema generation library
- Go 1.18+ (for generics support in `GenerateSchemaFromStruct`)

## Best Practices

1. **Use jsonschema tags**: Always tag required fields with `jsonschema:"required"`
2. **Keep schemas simple**: Avoid complex nested structures that might cause issues
3. **Validate schemas**: Use `ValidateOpenAICompatibility` to check for issues
4. **Test with OpenAI**: Always test generated schemas with actual OpenAI API calls
5. **Document constraints**: Use Go comments to document field constraints and validation rules

## Package Organization

This package is located within the OpenAI adapter (`pkg/adapters/openai/schemas/`) to keep OpenAI-specific code centralized and avoid spreading OpenAI dependencies across the codebase. 