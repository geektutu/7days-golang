---
title: 动手写ORM框架 - GeeORM第三天 插入和查询记录
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。实现新增(insert)记录的功能；使用反射(reflect)将数据库的记录转换为对应的结构体，实现查询(select)功能。
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
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第三篇。

- 实现新增(insert)记录的功能。
- 使用反射(reflect)将数据库的记录转换为对应的结构体，实现查询(select)功能。