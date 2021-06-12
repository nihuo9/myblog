---
title: "GO 1.15 以上版本解决GRPC X509 Common Name field, use SANs or temporarily enable Common Name matching"
subtitle: ""
date: 2021-06-06T15:22:32+08:00
lastmod: 2021-06-06T15:22:32+08:00
draft: true
author: "nihuo"
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: []
categories: []
questions: ["Go"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.png"

toc:
  enable: true
math:
  enable: false

license: ""
---

<!--more-->
GO1.15   X509 被砍了（不能用了） ，需要用到SAN证书，下面就介绍一下SAN证书生成

1：首先你的有OPENSSL，网上下载一个自己安装就可以了，

2：生成普通的key:     openssl genrsa -des3 -out server.key 2048     （记住你的密码以后需要用到，别忘了！！！）

3：生成ca的crt： openssl req -new -x509 -key server.key -out ca.crt -days 3650

4：生成csr：openssl req -new -key server.key -out server.csr 

以上就是基础，然后通过更改openssl.cfg（我是windos ,Linux是openssl.cnf）

1：把openssl.cfg 拷贝到当前目录（main.go 的目录就行）

2：找到 [ CA_default ],打开 copy_extensions = copy

3：找到[ req ],打开 req_extensions = v3_req # The extensions to add to a certificate request

4：找到[ v3_req ],添加 subjectAltName = @alt_names

5：添加新的标签 [ alt_names ] , 和标签字段  

DNS.1 = *.org.haha.com
DNS.2 = *.haha.com

6：生成证书私钥test.key：

  openssl genpkey -algorithm RSA -out test.key

7：通过私钥test.key生成证书请求文件test.csr：

openssl req -new -nodes -key test.key -out test.csr -days 3650 -subj "/C=cn/OU=myorg/O=mycomp/CN=myname" -config ./openssl.cfg -extensions v3_req

8：test.csr是上面生成的证书请求文件。ca.crt/server.key是CA证书文件和key，用来对test.csr进行签名认证。这两个文件在第一部分生成。

9：生成SAN证书：

openssl x509 -req -days 365 -in test.csr -out test.pem -CA ca.crt -CAkey server.key -CAcreateserial -extfile ./openssl.cfg -extensions v3_req

10：然后就可以用在 GO 1.15 以上版本的GRPC通信了，服务器加载代码:

creds, err := credentials.NewServerTLSFromFile("test.pem", "test.key")

11：客户端加载代码

creds,err := credentials.NewClientTLSFromFile("test.pem","*.org.haha.com")