---
title: 动手写ORM框架 - GeeORM第六天 支持事务(Transaction) 
date: 2020-03-03 23:00:00
description: 7天用 Go语言/golang 从零实现 ORM 框架 GeeORM 教程(7 days implement golang object relational mapping framework from scratch tutorial)，动手写 ORM 框架，参照 gorm, xorm 的实现。介绍 SQLite 数据库中的事务(transaction)；封装事务，用户自定义回调函数实现原子操作。
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
published: false
---

本文是[7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html)的第六篇。

- 介绍 SQLite 数据库中的事务(transaction)。
- 封装事务，用户自定义回调函数实现原子操作。