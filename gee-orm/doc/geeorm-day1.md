---
title: 动手写ORM框架 - GeeORM第一天 database/sql 基础
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。介绍了 SQLite 的基础操作（连接数据库，创建表、增删记录等），使用 Go 标准库 database/sql 操作 SQLite 数据库，包括执行(Exec)，查询(Query, QueryRow)。
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
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第一篇。

- SQLite 的基础操作（连接数据库，创建表、增删记录等）。
- 使用 Go 语言标准库 database/sql 连接并操作 SQLite 数据库，并简单封装。