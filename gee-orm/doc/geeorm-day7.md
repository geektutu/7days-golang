---
title: 动手写ORM框架 - GeeORM第七天 数据库迁移(Migrate)
date: 2020-03-08 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。结构体(struct)变更时，数据库表的字段(field)自动迁移(migrate)；仅支持字段新增与删除，不支持字段类型变更。
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
- migrate
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day7 数据库迁移
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第七篇。

- 结构体(struct)变更时，数据库表的字段(field)自动迁移(migrate)。
- 仅支持字段新增与删除，不支持字段类型变更。**代码约70行**

## 1 使用 SQL 语句 Migrate

数据库 Migrate 一直是数据库运维人员最为头痛的问题，如果仅仅是一张表增删字段还比较容易，那如果涉及到外键等复杂的关联关系，数据库的迁移就会变得非常困难。

GeeORM 的 Migrate 操作仅针对最为简单的场景，即支持字段的新增与删除，不支持字段类型变更。

在实现 Migrate 之前，我们先看看如何使用原生的 SQL 语句增删字段。

### 1.1 新增字段

```sql
ALTER TABLE table_name ADD COLUMN col_name, col_type;
```

大部分数据支持使用 `ALTER` 关键字新增字段，或者重命名字段。

### 1.2 删除字段

> 参考 [sqlite delete or add column - stackoverflow](https://stackoverflow.com/questions/8442147/how-to-delete-or-add-column-in-sqlite)

对于 SQLite 来说，删除字段并不像新增字段那么容易，一个比较可行的方法需要执行下列几个步骤：

```sql
CREATE TABLE new_table AS SELECT col1, col2, ... from old_table
DROP TABLE old_table
ALTER TABLE new_table RENAME TO old_table;
```

- 第一步：从 `old_table` 中挑选需要保留的字段到 `new_table` 中。
- 第二步：删除 `old_table`。
- 第三步：重命名 `new_table` 为 `old_table`。

## 2 GeeORM 实现 Migrate

按照原生的 SQL 命令，利用之前实现的事务，在 `geeorm.go` 中实现 Migrate 方法。

```go
// difference returns a - b
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

// Migrate table
func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
```

- `difference` 用来计算前后两个字段切片的差集。新表 - 旧表 = 新增字段，旧表 - 新表 = 删除字段。
- 使用 `ALTER` 语句新增字段。
- 使用创建新表并重命名的方式删除字段。

## 3 测试

在 `geeorm_test.go` 中添加 Migrate 的测试用例：

```go
type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestEngine_Migrate(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text PRIMARY KEY, XXX integer);").Exec()
	_, _ = s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	engine.Migrate(&User{})

	rows, _ := s.Raw("SELECT * FROM User").QueryRows()
	columns, _ := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"Name", "Age"}) {
		t.Fatal("Failed to migrate table User, got columns", columns)
	}
}
```

- 首先假设原有的 `User` 包含两个字段 `Name` 和 `XXX`，在一次业务变更之后，`User` 结构体的字段变更为 `Name` 和 `Age`。
- 即需要删除原有字段 `XXX`，并新增字段 `Age`。
- 调用 `Migrate(&User{})` 之后，新表的结构为 `Name`，`Age`

## 4 总结

GeeORM 的整体实现比较粗糙，比如数据库的迁移仅仅考虑了最简单的场景。实现的特性也比较少，比如结构体嵌套的场景，外键的场景，复合主键的场景都没有覆盖。ORM 框架的代码规模一般都比较大，如果想尽可能地逼近数据库，就需要大量的代码来实现相关的特性；二是数据库之间的差异也是比较大的，实现的功能越多，数据库之间的差异就会越突出，有时候为了达到较好的性能，就不得不为每个数据做特殊处理；还有些 ORM 框架同时支持关系型数据库和非关系型数据库，这就要求框架本身有更高层次的抽象，不能局限在 SQL 这一层。

GeeORM 仅 800 左右的代码是不可能做到这一点的。不过，GeeORM 的目的并不是实现一个可以在生产使用的 ORM 框架，而是希望尽可能多地介绍 ORM 框架大致的实现原理，例如

- 在框架中如何屏蔽不同数据库之间的差异；
- 数据库中表结构和编程语言中的对象是如何映射的；
- 如何优雅地模拟查询条件，链式调用是个不错的选择；
- 为什么 ORM 框架通常会提供 hooks 扩展的能力；
- 事务的原理和 ORM 框架如何集成对事务的支持；
- 一些难点问题，例如数据库迁移。
- ...

基于这几点，我觉得 GeeORM 的目的达到了。

## 附 推荐阅读

- [Go Test 单元测试简明教程](https://geektutu.com/post/quick-go-test.html)
- [SQLite 常用命令速查表](https://geektutu.com/post/cheat-sheet-sqlite.html)
- [sqlite delete or add column - stackoverflow](https://stackoverflow.com/questions/8442147/how-to-delete-or-add-column-in-sqlite)
