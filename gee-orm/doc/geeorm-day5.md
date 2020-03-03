---
title: 动手写ORM框架 - GeeORM第五天 实现钩子(Hooks)
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。通过反射(reflect)获取结构体绑定的钩子(hooks)，并调用；支持增删查改(CRUD)前后调用钩子。
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
- hooks
- BeforeUpdate
image: post/geeorm/geeorm_sm.jpg
github: https://github.com/geektutu/7days-golang
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第五篇。

- 通过反射(reflect)获取结构体绑定的钩子(hooks)，并调用。
- 支持增删查改(CRUD)前后调用钩子。