---
title: 动手写ORM框架 - GeeORM第三天 记录新增和查询
date: 2020-03-08 01:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。实现新增(insert)记录的功能；使用反射(reflect)将数据库的记录转换为对应的结构体实例，实现查询(select)功能。
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
- insert into
- select from
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day3 记录新增和查询
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第三篇。

- 实现新增(insert)记录的功能。
- 使用反射(reflect)将数据库的记录转换为对应的结构体实例，实现查询(select)功能。**代码约150行**

## 1 Clause 构造 SQL 语句

从第三天开始，GeeORM 需要涉及一些较为复杂的操作，例如查询操作。查询语句一般由很多个子句(clause) 构成。SELECT 语句的构成通常是这样的：

```sql
SELECT col1, col2, ...
    FROM table_name
    WHERE [ conditions ]
    GROUP BY col1
    HAVING [ conditions ]
```

也就是说，如果想一次构造出完整的 SQL 语句是比较困难的，因此我们将构造 SQL 语句这一部分独立出来，放在子package clause 中实现。

首先在 `clause/generator.go` 中实现各个子句的生成规则。

[day3-save-query/clause/generator.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day3-save-query/clause)


```go
package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
}

func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars

}

func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}
```

然后在 `clause/clause.go` 中实现结构体 `Clause` 拼接各个独立的子句。

[day3-save-query/clause/clause.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day3-save-query/clause)

```go
package clause

import "strings"

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int
const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
```

- `Set` 方法根据 `Type` 调用对应的 generator，生成该子句对应的 SQL 语句。
- `Build` 方法根据传入的 `Type` 的顺序，构造出最终的 SQL 语句。

在 `clause_test.go` 实现对应的测试用例：

```go
func testSelect(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{"Tom", 3}) {
		t.Fatal("failed to build SQLVars")
	}
}

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
}
```

## 2 实现 Insert 功能

首先为 Session 添加成员变量 clause

```go
// session/raw.go
type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	sqlVars  []interface{}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}
```

clause 已经支持生成简单的插入(INSERT) 和 查询(SELECT) 的 SQL 语句，那么紧接着我们就可以在 session 中实现对应的功能了。

INSERT 对应的 SQL 语句一般是这样的：

```sql
INSERT INTO table_name(col1, col2, col3, ...) VALUES
    (A1, A2, A3, ...),
    (B1, B2, B3, ...),
    ...
```

在 ORM 框架中期望 Insert 的调用方式如下：

```go
s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
u1 := &User{Name: "Tom", Age: 18}
u2 := &User{Name: "Sam", Age: 25}
s.Insert(u1, u2, ...)
```

也就是说，我们还需要一个步骤，根据数据库中列的顺序，从对象中找到对应的值，按顺序平铺。即 `u1`、`u2` 转换为 `("Tom", 18), ("Same", 25)` 这样的格式。

因此在实现 Insert 功能之前，还需要给 `Schema` 新增一个函数 `RecordValues` 完成上述的转换。

[day3-save-query/schema/schema.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day3-save-query/schema)

```go
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
```

在 session 文件夹下新建 record.go，用于实现记录增删查改相关的代码。

[day3-save-query/session/record.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day3-save-query/session)

```go
package session

import (
	"geeorm/clause"
	"reflect"
)

func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
```

后续所有构造 SQL 语句的方式都将与 `Insert` 中构造 SQL 语句的方式一致。分两步：

- 1）多次调用 `clause.Set()` 构造好每一个子句。
- 2）调用一次 `clause.Build()` 按照传入的顺序构造出最终的 SQL 语句。

构造完成后，调用 `Raw().Exec()` 方法执行。

## 3 实现 Find 功能

期望的调用方式是这样的：传入一个切片指针，查询的结果保存在切片中。

```go
s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
var users []User
s.Find(&users);
```

Find 功能的难点和 Insert 恰好反了过来。Insert 需要将已经存在的对象的每一个字段的值平铺开来，而 Find 则是需要根据平铺开的字段的值构造出对象。同样，也需要用到反射(reflect)。

```go
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}
```

Find 的代码实现比较复杂，主要分为以下几步：

- 1) `destSlice.Type().Elem()` 获取切片的单个元素的类型 `destType`，使用 `reflect.New()` 方法创建一个 `destType` 的实例，作为 `Model()` 的入参，映射出表结构 `RefTable()`。
- 2）根据表结构，使用 clause 构造出 SELECT 语句，查询到所有符合条件的记录 `rows`。
- 3）遍历每一行记录，利用反射创建 `destType` 的实例 `dest`，将 `dest` 的所有字段平铺开，构造切片 `values`。
- 4）调用 `rows.Scan()` 将该行记录每一列的值依次赋值给 values 中的每一个字段。
- 5）将 `dest` 添加到切片 `destSlice` 中。循环直到所有的记录都添加到切片 `destSlice` 中。

## 4 测试

在 session 文件夹下新建 `record_test.go`，创建测试用例。

> `User` 和 `NewSession()` 的定义位于 raw_test.go 中。

[day3-save-query/session/record_test.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day3-save-query/session)

```go
package session

import "testing"

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)

func testRecordInit(t *testing.T) *Session {
	t.Helper()
	s := NewSession().Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
	return s
}

func TestSession_Insert(t *testing.T) {
	s := testRecordInit(t)
	affected, err := s.Insert(user3)
	if err != nil || affected != 1 {
		t.Fatal("failed to create record")
	}
}

func TestSession_Find(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	if err := s.Find(&users); err != nil || len(users) != 2 {
		t.Fatal("failed to query all")
	}
}
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go Test 单元测试简明教程](https://geektutu.com/post/quick-go-test.html)
- [SQLite 常用命令速查表](https://geektutu.com/post/cheat-sheet-sqlite.html)
- [Laws Of Reflection - golang.org](https://blog.golang.org/laws-of-reflection)