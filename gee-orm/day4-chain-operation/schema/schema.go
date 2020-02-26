package schema

import (
	"fmt"
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// Field represents a column of database
type Field struct {
	Name string
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	TableName    string
	PrimaryField *Field
	Fields       []*Field
	FieldNames   []string
}

// Values return the values of dest's member variables
func (schema *Schema) Values(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

// Parse a struct to a Schema instance
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		TableName:    modelType.Name(),
		PrimaryField: &Field{Name: "ID", Tag: ""},
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Tag:  d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok && v == "primary_key" {
				schema.PrimaryField = field
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
		}
	}
	return schema
}

// String returns readable string
func (field *Field) String() string {
	return fmt.Sprintf("(%s %s)", field.Name, field.Tag)
}

// String returns readable string
func (schema *Schema) String() string {
	return fmt.Sprintf("TABLE %s %v", schema.TableName, schema.Fields)
}
