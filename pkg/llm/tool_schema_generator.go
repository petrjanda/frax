package llm

import "github.com/invopop/jsonschema"

type SchemaGenerator interface {
	Generate(v interface{}) (*jsonschema.Schema, error)
	MustGenerate(v interface{}) *jsonschema.Schema
}

type GenericSchemaGenerator struct {
	reflector *jsonschema.Reflector
}

func NewGenericSchemaGenerator() *GenericSchemaGenerator {
	return &GenericSchemaGenerator{
		reflector: &jsonschema.Reflector{
			ExpandedStruct:             true,
			DoNotReference:             true,
			RequiredFromJSONSchemaTags: true,
			AllowAdditionalProperties:  true,
		},
	}
}

func (g *GenericSchemaGenerator) Generate(v interface{}) (*jsonschema.Schema, error) {
	return g.reflector.Reflect(v), nil
}

func (g *GenericSchemaGenerator) MustGenerate(v interface{}) *jsonschema.Schema {
	if schema, err := g.Generate(v); err != nil {
		panic(err)
	} else {
		return schema
	}
}
