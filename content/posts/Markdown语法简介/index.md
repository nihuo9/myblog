---
title: "Markdown语法简介"
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

tags: ["Markdown"]
categories: ["其它"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.png"
---

本文介绍了一些基本的MarkDown语法
<!--more-->

**Markdown**是一种轻量级标记语言。它允许人们使用纯文本格式编写文档然后转换为XHTML(或者HTML)文档。使用Markdown编写文档有以下几个好处：
1. Markdown语法简单易于学习
2. 不容易出现错误
3. 可以转换为XHTML输出
4. 保持内容和显示格式分开

## 标题
**1、使用=和-标记一级和二级标题**
```markdown
一级标题
=========

二级标题
---------
```
**2、使用#号标记**
```markdown
# 一级标题
## 二级标题
### 三级标题
#### 四级标题
##### 五级标题
###### 六级标题
```
{{< admonition note >}}
 可以添加自定义的标题ID
 ```markdown
 # 标题 {#id}
 ```
 html输出会像下面这样
 ```html
 <h1 id="id">标题</h1>
 ```
{{< /admonition >}}

## 注释
注释与HTML兼容
```html
<!--这是一个注释-->
```
你看不到我：<!--这是一个注释-->

## 水平线
HTML的`<hr>`标签可以用来分隔两个不同主题的内容，在Markdown里你可以使用下面的方法来分隔内容：
* `___`: 三个下划线
* `---`: 三个破折号
* `***`: 三个星号
渲染后你可以看到：
___

---

***

## 正文
正文在转换为html后会被`<p></p>`标签包裹
换行可以用空行或者两个以上空格+回车来完成

## 内联HTML
如果你需要一个HTML标签，可以像下面这样使用：
```markdown
段落。。。

<div class="class">
  This is <b>HTML</b>
</div>

段落。。。
```

## 字体强调
**1、粗体**
```markdown
  **内容**
  __内容__
```
渲染后可以看到：
**内容**
__内容__

**2、斜体**
```markdown
  *content*
  _content_
```
渲染后可以看到：
*content*
_content_

**3、删除线**
```markdown
  ~~内容~~
```
渲染后可以看到：
~~内容~~

## 引用
```markdown
> **Markdown** 一种轻量标记语言
```
渲染后可以看到：
> **Markdown** 一种轻量标记语言

## 列表
**1、无序列表**
无序列表有三种定义的方式：
```markdown
* 项目一
- 项目一
+ 项目一
```
例子：
```markdown
* C
* C++
* Golang
```
渲染后可以看到：
* C
* C++
* Golang

**2、有序列表**
```markdown
1. 吃饭
2. 睡觉
3. 学习
```
渲染后可以看到：
1. 吃饭
2. 睡觉
3. 学习

{{< admonition tip >}}
如果你每项都使用1.来标号，那么Markdown会自动生成号码
```markdown
1. 吃饭
1. 睡觉
1. 学习
```
渲染后可以看到：
1. 吃饭
1. 睡觉
1. 学习
{{< /admonition >}}

**3、任务列表**
任务列表创造一个带复选框的项目列表，使用方法是`-`后空格跟一个`[ ]`，如果`[ ]`里的空格换成`x`说明选中。
```markdown
- [x] 写一篇博客
- [ ] 刷一道算法题
- [ ] 看一本书
```
渲染后可以看到：
- [x] 写一篇博客
- [ ] 刷一道算法题
- [ ] 看一本书

## 代码
**1、内联代码**
在需要显示的代码片段上加上<code>`</code>
例如：
```markdown
`<section></section>` 会被`<code></code>`标签包裹
```
渲染后可以看到：
`<section></section>` 会被`<code></code>`标签包裹

**2、使用代码栅栏<code>```</code>**
{{< highlight markdown >}}
```markdown
**Hi!**
```
{{< /highlight >}}

## 表格
通过在每栏之间加管道`|`，使用破折号`-`来分隔表头和其他行，可以做一个表格。
```markdown
| 姓名 | 成绩 |
| ---  | --- |
| 张三 | 0 |
| 李四 | 58 |
| 王五 | 59 |
| 我 | 100 |
```
渲染后可以看到：
| 姓名 | 成绩 |
| ---  | --- |
| 张三 | 0 |
| 李四 | 58 |
| 王五 | 59 |
| 我 | 100 |

{{< admonition note >}}
可以通过在破折号`-`两边添加冒号来改变对齐方式，左边有右边没有就是左对齐，右边有左边没有就是右对齐，两边都有就是中间对齐。
```markdown
| 姓名 | 成绩 |
| :---  | :--- |
| 张三 | 0 |
| 李四 | 58 |
| 王五 | 59 |
| 我 | 100 |
```
输出：
| 姓名 | 成绩 |
| :---  | :--- |
| 张三 | 0 |
| 李四 | 58 |
| 王五 | 59 |
| 我 | 100 |
{{< /admonition >}}

## 链接
**1、基本链接**
```markdown
<https://baidu.com/>
[Blibili](https://www.bilibili.com/)
```
渲染后可以看到：
<https://baidu.com/>

[Blibili](https://www.bilibili.com/)

**2、给链接添加一个标题(悬浮提示)**
```markdown
[百度](https://baidu.com/ "搜索")
```
渲染后可以看到：
[百度](https://baidu.com/ "搜索")

**3、链接一个命名的锚点**
```markdown
## 目录
  * [章节1](#chapter-1)
  * [章节2](#chapter-2)
  * [章节3](#chapter-3)
```

将跳到这些章节：
```markdown
## 章节1 <a id="chapter-1"></a>
章节1内容

## 章节2 <a id="chapter-2"></a>
章节2内容

## 章节3 <a id="chapter-3"></a>
章节3内容
```

## 脚注
脚注允许你添加注释和引用，而不会打乱文档的正文。当创建脚注时，添加脚注引用的地方会出现一个带有链接的上标数字。读者可以点击链接，跳转到页面底部脚注的内容。
```markdown
这是一个数字脚注[^1]
这是一个标签脚注[^label]
[^1]: 数字脚注内容
[^label]: 标签脚注内容
```
渲染后可以看到：
这是一个数字脚注[^1]
这是一个标签脚注[^label]
[^1]: 数字脚注内容

[^label]: 标签脚注内容

## 图像
图像的语法和链接的差不多就是前面多了个`!`。
```markdown
![Minion](https://octodex.github.com/images/minion.png)
```
![Minion](https://octodex.github.com/images/minion.png)