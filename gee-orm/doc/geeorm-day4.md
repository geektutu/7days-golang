---
title: 动手写ORM框架 - GeeORM第四天 链式操作与更新删除
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。通过链式(chain)操作，支持查询条件(where, order by, limit 等)的叠加；实现记录的更新(update)和删除(delete)功能。
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
- chain operation
- delete from
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第四篇。

- 通过链式(chain)操作，支持查询条件(where, order by, limit 等)的叠加。
- 实现记录的更新(update)和删除(delete)功能。