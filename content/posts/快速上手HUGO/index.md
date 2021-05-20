---
title: "快速上手HUGO"
subtitle: ""
date: 2021-05-18T22:27:15+08:00
lastmod: 2021-05-18T22:27:15+08:00
draft: false
author: ""
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["HUGO", "WEB"]
categories: ["HUGO"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.png"
---

本文简单介绍HUGO的安装和使用
<!--more-->

## 安装
使用源码安装：
```shell
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/gohugoio/hugo.git
cd hugo
go install --tags extended
```

Snap:
```shell
snap install hugo --channel=extended
```
{{< admonition note >}}
snap软件可以用snap run 运行
{{< /admonition >}}

apt-get:
```shell
sudo apt-get install hugo
```

## 创建一个新站
```shell
hugo new site yoursite
```

## 添加一个主题
可以从<https://themes.gohugo.io/>中选择一个，，首先从github下载并添加到你的站点目录下
```shell
cd yoursite
git init
git submodule add https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke
```

## 添加你的内容
```shell
hugo new posts/my-first-post.md
```
hugo会自动在content目录下创建相应的文件，生成的文件内容如下：
```yaml
---
title: "My First Post"
date: 2019-03-26T08:47:11+01:00
draft: true
---
```
你可以在这个文件下面编辑你的内容

## 运行HUGO服务器
```shell
hugo server -D
```
-D 强制生成草稿，就是说草稿也会显示出来
hugo 会在本地1313端口开启服务，打开浏览器就可以查看你的网站了

## 生成静态页面
```shell
hugo -D
```
hugo会在./public目录下生成你站点的静态文件，这样你只需要提供该目录的文件服务就可以运行你的网站了，简单的可以运行如下GO代码
```Go
package main

import (
	"log"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Static("/", "public")
  log.Fatal(r.Run(你的服务器地址))
}
```