---
title: 动手写ORM框架 - GeeORM第二天 对象表结构映射
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。使用反射(reflect)获取任意 struct 对象的名称和字段，映射为数据中的表；使用 dialect 隔离不同数据库之间的差异，便于扩展；数据表的创建(create)、删除(drop)。
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
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第二篇。

- 使用反射(reflect)获取任意 struct 对象的名称和字段，映射为数据中的表。
- 使用 dialect 隔离不同数据库之间的差异，便于扩展。
- 数据表的创建(create)、删除(drop)。