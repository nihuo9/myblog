---
title: "HUGO目录结构"
subtitle: ""
date: 2021-05-19T18:28:13+08:00
lastmod: 2021-05-19T18:28:13+08:00
draft: false
author: ""
description: ""

page:
    theme: "wide"

authorComment: ""

tags: ["HUGO"]
categories: ["HUGO"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.png"
---

本文介绍HUGO自动生成的目录结构
<!--more-->
我们使用hugo new site命令搭建了一个项目目录结构，目录的基本结构如下：

```
.
├── archetypes
├── config.toml
├── content
├── data
├── layouts
├── static
└── themes
```

## 目录结构解释
**archetypes**  
  当你使用`hugo new`命令创造文件时，会以这个文件夹下的对应文件作为模板生成你的文件，你可以用来设置你的front matter

**assets**   
  保存了你使用Hugo管道时需要使用的文件。只有使用了.Permalink或.RelPermalink的文件才会被发布到public目录。注意:默认情况下不创建assets目录。

**config**  
  Hugo带有大量配置指令。config目录是这些指令存储为JSON、YAML或TOML文件的地方。注意:config目录默认不创建。

**content**  
  你的网站的主要内容都在这个文件夹下。这个文件夹下的每个顶层文件夹都被Hugo作为一个内容部分。如果你的网站有三个主要的部分：博客、文章和教程，那么你可以在这个目录下生成content/blog、content/articles和content/tutorials。

**data**  
  这个目录用于存储配置文件，Hugo在生成您的网站时可以使用这些文件。可以放置YAML、JSON、TOML文件在这。除了添加到此文件夹中的文件外，还可以创建从动态内容中提取的数据模板。

**layout**  
  这个文件夹主要是放置生成HTML文件的模板。

**static**  
  存放所有的静态内容，例如：images、CSS、javascript等。当Hugo生成你的网站时所有在这个文件夹下的资源都会被相应的复制到public文件夹下。