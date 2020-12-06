---
title: 动手写ORM框架 - GeeORM第二天 对象表结构映射
date: 2020-03-08 00:20:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。使用反射(reflect)获取任意 struct 对象的名称和字段，映射为数据中的表；使用 dialect 隔离不同数据库之间的差异，便于扩展；数据库表的创建(create)、删除(drop)。
tags:
- Go
nav: 从零实现
categories:
- ORM框架 - GeeORM
keywords:
- Go语言
- 从零实现ORM框架
- database/sql
- sqlite
- reflect
- table mapping
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day2 对象表结构映射
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第二篇。

- 使用 dialect 隔离不同数据库之间的差异，便于扩展。
- 使用反射(reflect)获取任意 struct 对象的名称和字段，映射为数据中的表。
- 数据库表的创建(create)、删除(drop)。**代码约150行**

## 1 Dialect

SQL 语句中的类型和 Go 语言中的类型是不同的，例如Go 语言中的 `int`、`int8`、`int16` 等类型均对应 SQLite 中的 `integer` 类型。因此实现 ORM 映射的第一步，需要思考如何将 Go 语言的类型映射为数据库中的类型。

同时，不同数据库支持的数据类型也是有差异的，即使功能相同，在 SQL 语句的表达上也可能有差异。ORM 框架往往需要兼容多种数据库，因此我们需要将差异的这一部分提取出来，每一种数据库分别实现，实现最大程度的复用和解耦。这部分代码称之为 `dialect`。

在根目录下新建文件夹 dialect，并在 dialect 文件夹下新建文件 `dialect.go`，抽象出各个数据库差异的部分。

[day2-reflect-schema/dialect/dialect.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day2-reflect-schema/dialect)

```go
package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string
	TableExistSQL(tableName string) (string, []interface{})
}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
```

`Dialect` 接口包含 2 个方法：

- `DataTypeOf` 用于将 Go 语言的类型转换为该数据库的数据类型。
- `TableExistSQL` 返回某个表是否存在的 SQL 语句，参数是表名(table)。

当然，不同数据库之间的差异远远不止这两个地方，随着 ORM 框架功能的增多，dialect 的实现也会逐渐丰富起来，同时框架的其他部分不会受到影响。

同时，声明了 `RegisterDialect` 和 `GetDialect` 两个方法用于注册和获取 dialect 实例。如果新增加对某个数据库的支持，那么调用 `RegisterDialect` 即可注册到全局。

接下来，在`dialect` 目录下新建文件 `sqlite3.go` 增加对 SQLite 的支持。

[day2-reflect-schema/dialect/sqlite3.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day2-reflect-schema/dialect)

```go
package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type sqlite3 struct{}

var _ Dialect = (*sqlite3)(nil)

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

func (s *sqlite3) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (s *sqlite3) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}
```

- `sqlite3.go` 的实现虽然比较繁琐，但是整体逻辑还是非常清晰的。`DataTypeOf` 将 Go 语言的类型映射为 SQLite 的数据类型。`TableExistSQL` 返回了在 SQLite 中判断表 `tableName` 是否存在的 SQL 语句。
- 实现了 `init()` 函数，包在第一次加载时，会将 sqlite3 的 dialect 自动注册到全局。

## 2 Schema

Dialect 实现了一些特定的 SQL 语句的转换，接下来我们将要实现 ORM 框架中最为核心的转换——对象(object)和表(table)的转换。给定一个任意的对象，转换为关系型数据库中的表结构。

在数据库中创建一张表需要哪些要素呢？

- 表名(table name) —— 结构体名(struct name)
- 字段名和字段类型 —— 成员变量和类型。
- 额外的约束条件(例如非空、主键等) —— 成员变量的Tag（Go 语言通过 Tag 实现，Java、Python 等语言通过注解实现）

举一个实际的例子：

```go
type User struct {
    Name string `geeorm:"PRIMARY KEY"`
    Age  int
}
```

期望对应的 schema 语句：

```sql
CREATE TABLE `User` (`Name` text PRIMARY KEY, `Age` integer);
```

我们将这部分代码的实现放置在一个子包 `schema/schema.go` 中。

[day2-reflect-schema/schema/schema.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day2-reflect-schema/schema)

```go
package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// Field represents a column of database
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}
```

- Field 包含 3 个成员变量，字段名 Name、类型 Type、和约束条件 Tag
- Schema 主要包含被映射的对象 Model、表名 Name 和字段 Fields。
- FieldNames 包含所有的字段名(列名)，fieldMap 记录字段名和 Field 的映射关系，方便之后直接使用，无需遍历 Fields。

接下来实现 Parse 函数，将任意的对象解析为 Schema 实例。

```go
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
```

- `TypeOf()` 和 `ValueOf()` 是 reflect 包最为基本也是最重要的 2 个方法，分别用来返回入参的类型和值。因为设计的入参是一个对象的指针，因此需要 `reflect.Indirect()` 获取指针指向的实例。
- `modelType.Name()` 获取到结构体的名称作为表名。
- `NumField()` 获取实例的字段的个数，然后通过下标获取到特定字段 `p := modelType.Field(i)`。
- `p.Name` 即字段名，`p.Type` 即字段类型，通过 `(Dialect).DataTypeOf()` 转换为数据库的字段类型，`p.Tag` 即额外的约束条件。

写一个测试用例来验证 Parse 函数。

```go
// schema_test.go
type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	if schema.Name != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
	if schema.GetField("Name").Tag != "PRIMARY KEY" {
		t.Fatal("failed to parse primary key")
	}
}
```

## 3 Session

Session 的核心功能是与数据库进行交互。因此，我们将数据库表的增/删操作实现在子包 session 中。在此之前，Session 的结构需要做一些调整。

```go
type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect
	refTable *schema.Schema
	sql      strings.Builder
	sqlVars  []interface{}
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}
```

- `Session` 成员变量新增 dialect 和 refTable
- 构造函数 `New` 的参数改为 2 个，db 和 dialect。

在文件夹 `session` 下新建 `table.go` 用于放置操作数据库表相关的代码。

[day2-reflect-schema/session/table.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day2-reflect-schema/session)

```go
func (s *Session) Model(value interface{}) *Session {
	// nil or different model, update refTable
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}
```

- `Model()` 方法用于给 refTable 赋值。解析操作是比较耗时的，因此将解析的结果保存在成员变量 refTable 中，即使 `Model()` 被调用多次，如果传入的结构体名称不发生变化，则不会更新 refTable 的值。
- `RefTable()` 方法返回 refTable 的值，如果 refTable 未被赋值，则打印错误日志。

接下来实现数据库表的创建、删除和判断是否存在的功能。三个方法的实现逻辑是相似的，利用 `RefTable()` 返回的数据库表和字段的信息，拼接出 SQL 语句，调用原生 SQL 接口执行。

```go
func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}
```

在 `table_test.go` 中实现对应的测试用例：

```go
type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}
```

## 4 Engine

因为 Session 构造函数增加了对 dialect 的依赖，Engine 需要作一些细微的调整。

[day2-reflect-schema/geeorm.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day2-reflect-schema)

```go
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}
```

- `NewEngine` 创建 Engine 实例时，获取 driver 对应的 dialect。
- `NewSession` 创建 Session 实例时，传递 dialect 给构造函数 New。

至此，第二天的内容已经完成了，总结一下今天的成果：

- 1）为适配不同的数据库，映射数据类型和特定的 SQL 语句，创建 Dialect 层屏蔽数据库差异。
- 2）设计 Schema，利用反射(reflect)完成结构体和数据库表结构的映射，包括表名、字段名、字段类型、字段 tag 等。
- 3）构造创建(create)、删除(drop)、存在性(table exists) 的 SQL 语句完成数据库表的基本操作。

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go Test 单元测试简明教程](https://geektutu.com/post/quick-go-test.html)
- [Go Reflect 提高反射性能](https://geektutu.com/post/hpg-reflect.html)
- [SQLite 常用命令速查表](https://geektutu.com/post/cheat-sheet-sqlite.html)