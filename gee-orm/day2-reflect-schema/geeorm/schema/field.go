package schema

import "fmt"

type Field struct {
	Name  string
	Value interface{}
	Tag   string
}

func (f *Field) String() string {
	return fmt.Sprintf("%s %s", f.Name, f.Tag)
}
