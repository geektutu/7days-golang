package schema

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"geeorm/dialect"
)

type Schema struct {
	TableName    string
	PrimaryField *Field
	Fields       []*Field
}

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()

	schema := &Schema{
		TableName:    modelType.Name(),
		PrimaryField: &Field{Name: "ID", Value: 0},
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			schema.Fields = append(schema.Fields, &Field{
				Name: p.Name,
				Tag:  d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			})
		}
	}
	return schema
}

func (s *Schema) String() string {
	var fieldStr []string
	for _, field := range s.Fields {
		fieldStr = append(fieldStr, field.String())
	}

	return fmt.Sprintf("TABLE %s(%s)", s.TableName, strings.Join(fieldStr, ", "))
}
