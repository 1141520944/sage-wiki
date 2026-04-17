package ontology

import "github.com/xoai/sage-wiki/internal/config"

// EntityTypeDef defines an entity type with optional description.
type EntityTypeDef struct {
	Name        string
	Description string
}

// BuiltinEntityTypes defines the 5 immutable entity types.
var BuiltinEntityTypes = []EntityTypeDef{
	{Name: TypeConcept, Description: "An abstract idea or principle"},
	{Name: TypeTechnique, Description: "A method or algorithm"},
	{Name: TypeSource, Description: "A reference document or data source"},
	{Name: TypeClaim, Description: "An assertion that may be true or false"},
	{Name: TypeArtifact, Description: "A concrete output or deliverable"},
}

// MergedEntityTypes 合并实体类型会将用户配置与内置默认值进行合并。
// 内置类型始终存在，即便未在配置中指定。
// 配置可以扩展内置描述或添加自定义类型。
func MergedEntityTypes(cfgTypes []config.EntityTypeConfig) []EntityTypeDef {
	// Start with copies of builtins
	result := make([]EntityTypeDef, len(BuiltinEntityTypes))
	copy(result, BuiltinEntityTypes)

	builtinIdx := make(map[string]int, len(result))
	for i, e := range result {
		builtinIdx[e.Name] = i
	}

	for _, ct := range cfgTypes {
		if idx, ok := builtinIdx[ct.Name]; ok {
			// Override description for built-in type
			if ct.Description != "" {
				result[idx].Description = ct.Description
			}
		} else {
			// New custom type
			result = append(result, EntityTypeDef{
				Name:        ct.Name,
				Description: ct.Description,
			})
		}
	}

	return result
}

// ValidEntityTypeNames returns the names from a merged entity type list.
func ValidEntityTypeNames(defs []EntityTypeDef) []string {
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name
	}
	return names
}
