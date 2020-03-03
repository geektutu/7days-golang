---
title: 动手写ORM框架 - GeeORM第七天 数据库迁移(Migrate)
date: 2020-03-03 23:00:00
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
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第七篇。

- 结构体(struct)变更时，数据库表的字段(field)自动迁移(migrate)。
- 仅支持字段新增与删除，不支持字段类型变更。