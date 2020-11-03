---
title: 动手写ORM框架 - GeeORM第六天 支持事务(Transaction) 
date: 2020-03-08 21:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。介绍数据库中的事务(transaction)；封装事务，用户自定义回调函数实现原子操作。
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
- transaction
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day6 支持事务
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第六篇。

- 介绍数据库中的事务(transaction)。
- 封装事务，用户自定义回调函数实现原子操作。**代码约100行**

## 1 事务的 ACID 属性

> 数据库事务(transaction)是访问并可能操作各种数据项的一个数据库操作序列，这些操作要么全部执行,要么全部不执行，是一个不可分割的工作单位。事务由事务开始与事务结束之间执行的全部数据库操作组成。

举一个简单的例子，转账。A 转账给 B 一万元，那么数据库至少需要执行 2 个操作：

- 1）A 的账户减掉一万元。
- 2）B 的账户增加一万元。

这两个操作要么全部执行，代表转账成功。任意一个操作失败了，之前的操作都必须回退，代表转账失败。一个操作完成，另一个操作失败，这种结果是不能够接受的。这种场景就非常适合利用数据库事务的特性来解决。

如果一个数据库支持事务，那么必须具备 ACID 四个属性。

- 1）原子性(Atomicity)：事务中的全部操作在数据库中是不可分割的，要么全部完成，要么全部不执行。
- 2）一致性(Consistency): 几个并行执行的事务，其执行结果必须与按某一顺序 串行执行的结果相一致。
- 3）隔离性(Isolation)：事务的执行不受其他事务的干扰，事务执行的中间结果对其他事务必须是透明的。
- 4）持久性(Durability)：对于任意已提交事务，系统必须保证该事务对数据库的改变不被丢失，即使数据库出现故障。

## 2 SQLite 和 Go 标准库中的事务

SQLite 中创建一个事务的原生 SQL 长什么样子呢？

```sql
sqlite> BEGIN;
sqlite> DELETE FROM User WHERE Age > 25;
sqlite> INSERT INTO User VALUES ("Tom", 25), ("Jack", 18);
sqlite> COMMIT;
```

`BEGIN` 开启事务，`COMMIT` 提交事务，`ROLLBACK` 回滚事务。任何一个事务，均以 `BEGIN` 开始，`COMMIT` 或 `ROLLBACK` 结束。

Go 语言标准库 database/sql 提供了支持事务的接口。用一个简单的例子，看一看 Go 语言标准是如何支持事务的。

```go
package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, _ := sql.Open("sqlite3", "gee.db")
	defer func() { _ = db.Close() }()
	_, _ = db.Exec("CREATE TABLE IF NOT EXISTS User(`Name` text);")

	tx, _ := db.Begin()
	_, err1 := tx.Exec("INSERT INTO User(`Name`) VALUES (?)", "Tom")
	_, err2 := tx.Exec("INSERT INTO User(`Name`) VALUES (?)", "Jack")
	if err1 != nil || err2 != nil {
		_ = tx.Rollback()
		log.Println("Rollback", err1, err2)
	} else {
		_ = tx.Commit()
		log.Println("Commit")
	}
}
```

Go 语言中实现事务和 SQL 原生语句其实是非常接近的。调用 `db.Begin()` 得到 `*sql.Tx` 对象，使用 `tx.Exec()` 执行一系列操作，如果发生错误，通过 `tx.Rollback()` 回滚，如果没有发生错误，则通过 `tx.Commit()` 提交。

## 3 GeeORM 支持事务

GeeORM 之前的操作均是执行完即自动提交的，每个操作是相互独立的。之前直接使用 `sql.DB` 对象执行 SQL 语句，如果要支持事务，需要更改为 `sql.Tx` 执行。在 Session 结构体中新增成员变量 `tx *sql.Tx`，当 `tx` 不为空时，则使用 `tx` 执行 SQL 语句，否则使用 `db` 执行 SQL 语句。这样既兼容了原有的执行方式，又提供了对事务的支持。

[day6-transaction/session/raw.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day6-transaction/session)

```go
type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect
	tx       *sql.Tx
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	sqlVars  []interface{}
}

// CommonDB is a minimal function set of db
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB returns tx if a tx begins. otherwise return *sql.DB
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
```

新建文件 `session/transaction.go` 封装事务的 Begin、Commit 和 Rollback 三个接口。

[day6-transaction/session/transaction.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day6-transaction/session)

```go
package session

import "geeorm/log"

func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) Rollback() (err error) {
	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	return
}
```

- 调用 `s.db.Begin()` 得到 `*sql.Tx` 对象，赋值给 s.tx。
- 封装的另一个目的是统一打印日志，方便定位问题。


最后一步，在 `geeorm.go` 中为用户提供傻瓜式/一键式使用的接口。

[day6-transaction/geeorm.go](https://github.com/geektutu/7days-golang/tree/master/gee-orm/day6-transaction)

```go
type TxFunc func(*session.Session) (interface{}, error)

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return f(s)
}
```

> Transaction 的实现参考了 [stackoverflow](https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback)

用户只需要将所有的操作放到一个回调函数中，作为入参传递给 `engine.Transaction()`，发生任何错误，自动回滚，如果没有错误发生，则提交。

## 4 测试

在 `geeorm_test.go` 中添加测试用例看看 Transaction 如何工作的吧。

```go
func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("failed to connect", err)
	}
	return engine
}

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}
```

首先是 rollback 的用例：

```go
func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return nil, errors.New("Error")
	})
	if err == nil || s.HasTable() {
		t.Fatal("failed to rollback")
	}
}
```

- 在这个用例中，如何执行成功，则会创建一张表 `User`，并插入一条记录。
- 故意返回了一个自定义 error，最终事务回滚，表创建失败。

接下来是 commit 的用例：

```go
func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return
	})
	u := &User{}
	_ = s.First(u)
	if err != nil || u.Name != "Tom" {
		t.Fatal("failed to commit")
	}
}
```

- 创建表和插入记录均成功执行，最终通过 `s.First()` 方法查询到插入的记录。

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go Test 单元测试简明教程](https://geektutu.com/post/quick-go-test.html)
- [SQLite 常用命令速查表](https://geektutu.com/post/cheat-sheet-sqlite.html)