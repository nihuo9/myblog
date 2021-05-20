---
title: "Git"
subtitle: ""
date: 2021-05-21T15:28:14+08:00
lastmod: 2021-05-21T15:28:14+08:00
draft: true
author: ""
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["Git"]
categories: ["工具"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.jpg"
---

Git教程
<!--more-->

## 简介
Git是一个开源的分布式版本控制系统，可以有效、高速地处理从很小到非常大的项目版本管理。

### 工作流程
![工作流程](1.png)

作为开发者基本的工作流程是  
1. 修改文件、添加文件等操作
2. 把所有改动的文件添加的暂存区
3. 检查改动
4. 提交暂存区文件到本地Git仓库
5. 推送到远程仓库

## 安装

yum：  
```shell
sudo yum install git
```
apt-get:
```shell
sudo apt-get install git
```

## 配置
Git可以通过git config命令来管理配置。git的配置文件存在在三个位置  
/etc/gitconfig：包含系统上每一个用户及他们仓库的通用配置。如果使用带有--system选项的 git config时，它会从此文件读写配置变量。  
~/.gitconfig;~/.config/git/：只针对当前用户。可以传递--global来读写该文件  
./.git/config：针对该仓库

### 用户信息
```shell
git config --global user.name "username"
git config --global user.email "xxx@xxx"
```
用户信息在提交更改时要用到。

### 查看配置
```shell
git config --get key
```

### 设置文本编辑器
```shell
git config --global core.editor vim
```

### 获取帮助
```shell
git config --help
git help config
```

### 设置代理
取消代理 
```shell
git config --gloabl --unset http.proxy
git config --gloabl --unset https.proxy
```
使用代理以7890端口为例
```shell
git config --global http.https://github.com.proxy http://127.0.0.1:7890
git config --global https.https://github.com.proxy https://127.0.0.1:7890
```

socket协议
```shell
git config --global http.proxy 'socks5://127.0.0.1:1080'
git config --global https.proxy 'socks5://127.0.0.1:1080'
git config --global http.https://github.com.proxy socks5://127.0.0.1:1080
#取消代理
git config --global --unset http.https://github.com.proxy
```

## 快速开始
这里远程仓库以github为例子  
1. 先在本地生成个密钥
   ```shell
   ssh-keygen -t ed25519 -C "xxx@xxx"
   ```
   完成后如下图所示：
   ![](2.png)
   记住公钥位置等等要用
2. 设置ssh认证代理
   ```shell
   eval $(ssh-agent -s)
   ```
3. 将第一步生成的密钥添加到ssh代理中
   ```shell
   ssh-add ~/.ssh/id_ed25519
   ```
4. 将公钥复制到github的设置中
   ![](3.png)
5. 用下面的命令测试下是否能认证成功
   ```shell
   ssh -T git@github.com
   ```
6. 把远程的仓库克隆到本地
   ```shell
   git clone https://github.com/nihuo9/hello-world
   ```
7. 新添加一个文件
   ```shell
   touch HelloWorld.go
   ```
8. 把文件保存到暂存区
  ```shell
  git add HelloWorld.go
  ```
9. 提交暂存区中的文件
  ```shell
  git commit -m "test"
  ```
10. 推送到远程的main分支
  ```shell
  git push origin main
  ```

## git init
`git init`命令创建一个空的Git仓库或重新初始化一个现有仓库
```shell
git init [-q | --quiet] [--bare] [--template=<template_directory>]
      [--separate-git-dir <git dir>]
      [--shared[=<permissions>]] [directory]
```
该命令创建一个空的Git仓库，一个名为.git的目录，目录结构如下：
![](4.png)
现有存储库中运行git init命令是安全的。它不会覆盖已经存在的东西。重新运行git init的主要原因是拾取新添加的模板(或者如果给出了--separate-git-dir，则将存储库移动到另一个地方)。

## git add
`git add`命令将文件添加到暂存区
```shell
git add [--verbose | -v] [--dry-run | -n] [--force | -f] [--interactive | -i] [--patch | -p]
      [--edit | -e] [--[no-]all | --[no-]ignore-removal | [--update | -u]]
      [--intent-to-add | -N] [--refresh] [--ignore-errors] [--ignore-missing]
      [--chmod=(+|-)x] [--] [<pathspec>…​]
```
此命令将要提交的文件的信息添加到索引库中(将修改添加到暂存区)，以准备为下一次提交分段的内容。它通常将现有路径的当前内容作为一个整体添加，但是通过一些选项，它也可以用于添加内容，只对所应用的工作树文件进行一些更改，或删除工作树中不存在的路径了。  
“索引”保存工作树内容的快照，并且将该快照作为下一个提交的内容。因此，在对工作树进行任何更改之后，并且在运行[git commit]()命令之前，必须使用git add命令将任何新的或修改的文件添加到索引。  
该命令可以在提交之前多次执行。它只在运行git add命令时添加指定文件的内容; 如果希望随后的更改包含在下一个提交中，那么必须再次运行git add将新的内容添加到索引。  
[git status](#git-status)命令可用于获取哪些文件具有为下一次提交分段的更改的摘要。
默认情况下，git add命令不会添加忽略的文件。 如果在命令行上显式指定了任何忽略的文件，git add命令都将失败，并显示一个忽略的文件列表。由Git执行的目录递归或文件名遍历所导致的忽略文件将被默认忽略。git add命令可以用-f(force)选项添加被忽略的文件。  

**常用选项**  
`git add -u <path>`：把`<path>`中所有跟踪文件中被修改或删除的文件信息添加到暂存区。  
`git add -A`：表示把当前目录所有文件添加到暂存区。
`git add -i <path>`：查看所有修改或者删除但没有提交的文件。

## git clone
将存储库克隆到新目录中。
```shell
git clone [--template=<template_directory>]
      [-l] [-s] [--no-hardlinks] [-q] [-n] [--bare] [--mirror]
      [-o <name>] [-b <name>] [-u <upload-pack>] [--reference <repository>]
      [--dissociate] [--separate-git-dir <git dir>]
      [--depth <depth>] [--[no-]single-branch]
      [--recurse-submodules] [--[no-]shallow-submodules]
      [--jobs <n>] [--] <repository> [<directory>]
```
将存储库克隆到新创建的目录中，为克隆的存储库中的每个分支创建远程跟踪分支(使用git branch -r可见)，并从克隆检出的存储库作为当前活动分支的初始分支。
在克隆之后，没有参数的普通git提取将更新所有远程跟踪分支，并且没有参数的git pull将另外将远程主分支合并到当前主分支(如果有的话)。
此默认配置通过在refs/remotes/origin下创建对远程分支头的引用，并通过初始化remote.origin.url和remote.origin.fetch配置变量来实现。
执行远程操作的第一步，通常是从远程主机克隆一个版本库，这时就要用到git clone命令。

**示例**  
```shell
git clone https://github.com/go-redis/redis
git clone https://github.com/go-redis/redis.git localPath
```

**常用选项**  
`-l`：从本地克隆仓库  
`-s`：从本地克隆仓库是不使用硬链接而是设置.git/objects/info/与源存储库共享对象
`-n`：不进行检出

## git status
用于查看工作目录和暂存区的状态
```shell
git status [<options>…​] [--] [<pathspec>…​]
```
显示索引文件和当前HEAD提交之间的差异，在工作树和索引文件之间有差异的路径以及工作树中没有被Git跟踪的路径。第一个是通过运行[git commit](#git-commit)来提交的; 第二个和第三个是你可以通过在运行git commit之前运行git add来提交的。

**常用选项**  
`-u[<mode>]`：显示未跟踪文件  
mode有以下取值  
* no - 不显示未跟踪文件
* normal - 显示未跟踪的文件和目录
* all - 还显示了未跟踪目录中的单个文件

## git diff
用于显示提交和工作树等之间的更改。此命令比较的是工作目录中当前文件和暂存区域快照之间的差异,也就是修改之后还没有暂存起来的变化内容。
```shell
git diff [options] [<commit>] [--] [<path>…​]
git diff [options] --cached [<commit>] [--] [<path>…​]
git diff [options] <commit> <commit> [--] [<path>…​]
git diff [options] <blob> <blob>
git diff [options] [--no-index] [--] <path> <path>
```
在工作树和索引或树之间显示更改，索引和树之间的更改，两个树之间的更改，两个blob对象之间的更改或两个文件在磁盘上的更改。

**示例**  
```shell
git diff <file> # 比较当前文件和暂存区文件差异 git diff
git diff <id1><id1><id2> # 比较两次提交之间的差异
git diff <branch1> <branch2> # 在两个分支之间比较
git diff --staged # 比较暂存区和版本库差异
git diff --cached # 比较暂存区和版本库差异
git diff --stat # 仅仅比较统计信息
git diff HEAD # 自上次提交以来工作树中的更改
```

## git commit
将暂存的当前内容与描述更改的用户和日志消息一起存储在新的提交中。
```shell
git commit [-a | --interactive | --patch] [-s] [-v] [-u<mode>] [--amend]
       [--dry-run] [(-c | -C | --fixup | --squash) <commit>]
       [-F <file> | -m <msg>] [--reset-author] [--allow-empty]
       [--allow-empty-message] [--no-verify] [-e] [--author=<author>]
       [--date=<date>] [--cleanup=<mode>] [--[no-]status]
       [-i | -o] [-S[<keyid>]] [--] [<file>…​]
```
要添加的内容可以通过以下几种方式指定：

* 在使用git commit命令之前，通过使用[git add](#git-add)对索引进行递增的“添加”更改(注意：修改后的文件的状态必须为“added”);  
* 通过使用git rm从工作树和索引中删除文件，再次使用git commit命令;  
* 通过将文件作为参数列出到git commit命令(不使用--interactive或--patch选项)，在这种情况下，提交将忽略索引中分段的更改，而是记录列出的文件的当前内容(必须已知到Git的内容) ;  
* 通过使用带有-a选项的git commit命令来自动从所有已知文件(即所有已经在索引中列出的文件)中添加“更改”，并自动从已从工作树中删除索引中的“rm”文件 ，然后执行实际提交;  
* 通过使用--interactive或--patch选项与git commit命令一起确定除了索引中的内容之外哪些文件或hunks应该是提交的一部分，然后才能完成操作。

如果提交后，然后立即发现错误，可以使用 [git reset](#git-reset) 命令恢复。

**示例**  
```shell
git add newfile.txt
git commit -m "the commit message"
git commit -a # 会先把所有已经track的文件的改动`git add`进来，然后提交。对于没有track的文件,还是需要执行`git add <file>` 命令。
git commit --amend # 增补提交，会使用与当前提交节点相同的父节点进行一次新的提交，旧的提交将会被取消。
```