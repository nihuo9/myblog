---
title: "Docker的基本使用"
subtitle: ""
date: 2021-09-24T17:16:28+08:00
lastmod: 2021-09-24T17:16:28+08:00
draft: false
author: ""
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["容器"]
categories: ["工具"]

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
## 安装
### 官方脚本自动安装
```shell
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
```

### 手动安装

### 卸载旧版本
Docker 的旧版本被称为 docker，docker.io 或 docker-engine 。如果已安装，请卸载它们：
```shell
$ sudo apt-get remove docker docker-engine docker.io containerd runc
```

### 使用Docker仓库安装
在新主机上首次安装 Docker Engine-Community 之前，需要设置 Docker 仓库。之后，您可以从仓库安装和更新 Docker 。

**设置仓库**  
更新apt包索引
```shell
$ sudo apt-get update
```

安装 apt 依赖包，用于通过HTTPS来获取仓库:
```shell
$ sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
```

添加Docker的官方GPG密钥：
```shell
$ curl -fsSL https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
```

9DC8 5822 9FC7 DD38 854A E2D8 8D81 803C 0EBF CD88 通过搜索指纹的后8个字符，验证您现在是否拥有带有指纹的密钥。
```shell
$ sudo apt-key fingerprint 0EBFCD88
   
pub   rsa4096 2017-02-22 [SCEA]
      9DC8 5822 9FC7 DD38 854A  E2D8 8D81 803C 0EBF CD88
uid           [ unknown] Docker Release (CE deb) <docker@docker.com>
sub   rsa4096 2017-02-22 [S]
```

使用以下指令设置稳定版仓库
```shell
$ sudo add-apt-repository \
   "deb [arch=amd64] https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu/ \
  $(lsb_release -cs) \
  stable"
```

**安装 Docker Engine-Community**  
更新 apt 包索引。
```shell
$ sudo apt-get update
```

安装最新版本的 Docker Engine-Community 和 containerd ，或者转到下一步安装特定版本：
```shell
$ sudo apt-get install docker-ce docker-ce-cli containerd.io
```

要安装特定版本的 Docker Engine-Community，请在仓库中列出可用版本，然后选择一种安装。列出您的仓库中可用的版本：
```shell
$ apt-cache madison docker-ce

  docker-ce | 5:18.09.1~3-0~ubuntu-xenial | https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu  xenial/stable amd64 Packages
  docker-ce | 5:18.09.0~3-0~ubuntu-xenial | https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu  xenial/stable amd64 Packages
  docker-ce | 18.06.1~ce~3-0~ubuntu       | https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu  xenial/stable amd64 Packages
  docker-ce | 18.06.0~ce~3-0~ubuntu       | https://mirrors.ustc.edu.cn/docker-ce/linux/ubuntu  xenial/stable amd64 Packages
  ...
```

使用第二列中的版本字符串安装特定版本，例如 5:18.09.1~3-0~ubuntu-xenial。
```shell
$ sudo apt-get install docker-ce=<VERSION_STRING> docker-ce-cli=<VERSION_STRING> containerd.io
```

测试 Docker 是否安装成功，输入以下指令，打印出以下信息则安装成功:
```shell
$ sudo docker run hello-world

Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
1b930d010525: Pull complete                                                                                                                                  Digest: sha256:c3b4ada4687bbaa170745b3e4dd8ac3f194ca95b2d0518b417fb47e5879d9b5f
Status: Downloaded newer image for hello-world:latest


Hello from Docker!
This message shows that your installation appears to be working correctly.


To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.


To try something more ambitious, you can run an Ubuntu container with:
 $ docker run -it ubuntu bash


Share images, automate workflows, and more with a free Docker ID:
 https://hub.docker.com/


For more examples and ideas, visit:
 https://docs.docker.com/get-started/
```

### 使用Shell脚本进行安装
Docker 在 get.docker.com 和 test.docker.com 上提供了方便脚本，用于将快速安装 Docker Engine-Community 的边缘版本和测试版本。脚本的源代码在 docker-install 仓库中。 不建议在生产环境中使用这些脚本，在使用它们之前，您应该了解潜在的风险：
* 脚本需要运行 root 或具有 sudo 特权。因此，在运行脚本之前，应仔细检查和审核脚本。
* 这些脚本尝试检测 Linux 发行版和版本，并为您配置软件包管理系统。此外，脚本不允许您自定义任何安装参数。从 Docker 的角度或您自己组织的准则和标准的角度来看，这可能导致不支持的配置。
* 这些脚本将安装软件包管理器的所有依赖项和建议，而无需进行确认。这可能会安装大量软件包，具体取决于主机的当前配置。
* 该脚本未提供用于指定要安装哪个版本的 Docker 的选项，而是安装了在 edge 通道中发布的最新版本。
* 如果已使用其他机制将 Docker 安装在主机上，请不要使用便捷脚本。

本示例使用 get.docker.com 上的脚本在 Linux 上安装最新版本的Docker Engine-Community。要安装最新的测试版本，请改用 test.docker.com。在下面的每个命令，取代每次出现 get 用 test。
```shell
$ curl -fsSL https://get.docker.com -o get-docker.sh
$ sudo sh get-docker.sh
```

如果要使用 Docker 作为非 root 用户，则应考虑使用类似以下方式将用户添加到 docker 组：
```shell
$ sudo usermod -aG docker your-user
```

### 镜像加速
国内从 DockerHub 拉取镜像有时会遇到困难，此时可以配置镜像加速器。Docker 官方和国内很多云服务商都提供了国内加速器服务，例如：
*  科大镜像：https://docker.mirrors.ustc.edu.cn/
*  网易：https://hub-mirror.c.163.com/
*  阿里云：https://<你的ID>.mirror.aliyuncs.com
*  七牛云加速器：https://reg-mirror.qiniu.com

当配置某一个加速器地址之后，若发现拉取不到镜像，请切换到另一个加速器地址。国内各大云服务商均提供了 Docker 镜像加速服务，建议根据运行 Docker 的云平台选择对应的镜像加速服务。

阿里云镜像获取地址：https://cr.console.aliyun.com/cn-hangzhou/instances/mirrors，登陆后，左侧菜单选中镜像加速器就可以看到你的专属地址了

**Ubuntu16.04+、Debian8+、CentOS7**  
对于使用 systemd 的系统，请在 /etc/docker/daemon.json 中写入如下内容（如果文件不存在请新建该文件）：
```shell
{"registry-mirrors":["https://reg-mirror.qiniu.com/"]}
```

之后重新启动服务：
```shell
$ sudo systemctl daemon-reload
$ sudo systemctl restart docker
```

检查加速器是否生效配置加速器之后，如果拉取镜像仍然十分缓慢，请手动检查加速器配置是否生效，在命令行执行 docker info，如果从结果中看到了如下内容，说明配置成功。
```shell
$ docker info
Registry Mirrors:
    https://reg-mirror.qiniu.com
```

## 使用容器

1. 查看所有命令

```shell
$ docker

Usage:  docker [OPTIONS] COMMAND

A self-sufficient runtime for containers

Options:
      --config string      Location of client config files (default "/home/ubuntu/.docker")
  -c, --context string     Name of the context to use to connect to the daemon (overrides DOCKER_HOST env var and
                           default context set with "docker context use")
  -D, --debug              Enable debug mode
  -H, --host list          Daemon socket(s) to connect to
  -l, --log-level string   Set the logging level ("debug"|"info"|"warn"|"error"|"fatal") (default "info")
      --tls                Use TLS; implied by --tlsverify
      --tlscacert string   Trust certs signed only by this CA (default "/home/ubuntu/.docker/ca.pem")
      --tlscert string     Path to TLS certificate file (default "/home/ubuntu/.docker/cert.pem")
      --tlskey string      Path to TLS key file (default "/home/ubuntu/.docker/key.pem")
      --tlsverify          Use TLS and verify the remote
  -v, --version            Print version information and quit

Management Commands:
  app*        Docker App (Docker Inc., v0.9.1-beta3)
  builder     Manage builds
  buildx*     Build with BuildKit (Docker Inc., v0.6.1-docker)
  config      Manage Docker configs
  container   Manage containers
  context     Manage contexts
  image       Manage images
  manifest    Manage Docker image manifests and manifest lists
  network     Manage networks
  node        Manage Swarm nodes
  plugin      Manage plugins
  scan*       Docker Scan (Docker Inc., v0.8.0)
  secret      Manage Docker secrets
  service     Manage services
  stack       Manage Docker stacks
  swarm       Manage Swarm
  system      Manage Docker
  trust       Manage trust on Docker images
  volume      Manage volumes

Commands:
  attach      Attach local standard input, output, and error streams to a running container
  build       Build an image from a Dockerfile
  commit      Create a new image from a container's changes
  cp          Copy files/folders between a container and the local filesystem
  create      Create a new container
  diff        Inspect changes to files or directories on a container's filesystem
  events      Get real time events from the server
  exec        Run a command in a running container
  export      Export a container's filesystem as a tar archive
  history     Show the history of an image
  images      List images
  import      Import the contents from a tarball to create a filesystem image
  info        Display system-wide information
  inspect     Return low-level information on Docker objects
  kill        Kill one or more running containers
  load        Load an image from a tar archive or STDIN
  login       Log in to a Docker registry
  logout      Log out from a Docker registry
  logs        Fetch the logs of a container
  pause       Pause all processes within one or more containers
  port        List port mappings or a specific mapping for the container
  ps          List containers
  pull        Pull an image or a repository from a registry
  push        Push an image or a repository to a registry
  rename      Rename a container
  restart     Restart one or more containers
  rm          Remove one or more containers
  rmi         Remove one or more images
  run         Run a command in a new container
  save        Save one or more images to a tar archive (streamed to STDOUT by default)
  search      Search the Docker Hub for images
  start       Start one or more stopped containers
  stats       Display a live stream of container(s) resource usage statistics
  stop        Stop one or more running containers
  tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE
  top         Display the running processes of a container
  unpause     Unpause all processes within one or more containers
  update      Update configuration of one or more containers
  version     Show the Docker version information
  wait        Block until one or more containers stop, then print their exit codes

Run 'docker COMMAND --help' for more information on a command.
```

2. 获取镜像

如果我们本地没有 ubuntu 镜像，我们可以使用 docker pull 命令来载入 ubuntu 镜像：
```shell
$ docker pull ubuntu
```

3. 启动容器

以下命令使用 ubuntu 镜像启动一个容器，参数为以命令行模式进入该容器：
```shell
$ docker run -it --name firstName ubuntu /bin/bash
```
说明：
* -i: 交互式操作
* -t: 终端
* ubuntu: 镜像名称
* /bin/bash: 运行的命令
* --name: 容器命名

退出终端，输入exit命令

4. 查看容器

```shell
docker ps
```

参数：  
* -a：查看所有容器
* -l：查看最后一次创建的容器

5. 启动已经停止的容器

```shell
docker start <CONTAINER ID>
```

6. 后台运行一个容器

在大部分的场景下，我们希望 docker 的服务是在后台运行的，我们可以过 -d 指定容器的运行模式。
```shell
$ docker run -itd --name ubuntu-test ubuntu /bin/bash
```

7. 停止一个容器

```shell
$ docker stop <CONTAINER ID>
```

8. 进入容器

在使用 -d 参数时，容器启动后会进入后台。此时想要进入容器，可以通过以下指令进入：
* docker attach
* docker exec：推荐大家使用 docker exec 命令，因为此退出容器终端，不会导致容器的停止

9. 导出和导入容器

导出容器:
```shell
$ docker export 1e560fca3906 > ubuntu.tar
```

导入容器:
可以使用 docker import 从容器快照文件中再导入为镜像，以下实例将快照文件 ubuntu.tar 导入到镜像 test/ubuntu:v1:
```shell
$ cat docker/ubuntu.tar | docker import - test/ubuntu:v1
```

此外，也可以通过指定 URL 或者某个目录来导入，例如：
```shell
$ docker import http://example.com/exampleimage.tgz example/imagerepo
```

10. 删除容器

```shell
$ docker rm -f 1e560fca3906
```

下面的命令可以清理掉所有处于终止状态的容器。
```shell
$ docker container prune 
```

11. 运行一个web应用

前面我们运行的容器并没有一些什么特别的用处。接下来让我们尝试使用 docker 构建一个 web 应用程序。我们将在docker容器中运行一个 Python Flask 应用来运行一个web应用。
```shell
docker pull training/webapp  # 载入镜像
docker run -d -P training/webapp python app.py
```

说明：  
* -d:让容器在后台运行。
* -P:将容器内部使用的网络端口随机映射到我们使用的主机上。

我们也可以通过 -p 参数来设置不一样的端口：
```shell
$ docker run -d -p 5000:5000 training/webapp python app.py
```

通过 docker ps 命令可以查看到容器的端口映射，docker 还提供了另一个快捷方式 docker port，使用 docker port 可以查看指定 （ID 或者名字）容器的某个确定端口映射到宿主机的端口号。上面我们创建的 web 应用容器 ID 为 bf08b7f2cd89 名字为 wizardly_chandrasekhar。我可以使用 docker port bf08b7f2cd89 或 docker port wizardly_chandrasekhar 来查看容器端口的映射情况。
```shell
$ docker port bf08b7f2cd89
5000/tcp -> 0.0.0.0:5000
```

12. 查看日志

```shell
$ docker logs -f bf08b7f2cd89
 * Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)
192.168.239.1 - - [09/May/2016 16:30:37] "GET / HTTP/1.1" 200 -
192.168.239.1 - - [09/May/2016 16:30:37] "GET /favicon.ico HTTP/1.1" 404 -
```

说明：  
* -f: 让 docker logs 像使用 tail -f 一样来输出容器内部的标准输出。

13. 查看容器内部运行的进程

```shell
$ docker top wizardly_chandrasekhar
UID     PID         PPID          ...       TIME                CMD
root    23245       23228         ...       00:00:00            python app.py
```

14. 检查容器

```shell
$ docker inspect wizardly_chandrasekhar
[
    {
        "Id": "bf08b7f2cd897b5964943134aa6d373e355c286db9b9885b1f60b6e8f82b2b85",
        "Created": "2018-09-17T01:41:26.174228707Z",
        "Path": "python",
        "Args": [
            "app.py"
        ],
        "State": {
            "Status": "running",
            "Running": true,
            "Paused": false,
            "Restarting": false,
            "OOMKilled": false,
            "Dead": false,
            "Pid": 23245,
            "ExitCode": 0,
            "Error": "",
            "StartedAt": "2018-09-17T01:41:26.494185806Z",
            "FinishedAt": "0001-01-01T00:00:00Z"
        },
......
```

15. 停止容器

```shell
$ docker stop wizardly_chandrasekhar   
```

16. 重启正在运行的容器

```shell
docker restart wizardly_chandrasekhar
```

17. 移除容器

```shell
$ docker rm wizardly_chandrasekhar 
```

删除容器时，容器必须是停止状态，否则会报错误

## 使用镜像

1. 列出镜像列表

```shell
$ docker images           
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
ubuntu              14.04               90d5884b1ee0        5 days ago          188 MB
php                 5.6                 f40e9e0f10c8        9 days ago          444.8 MB
nginx               latest              6f8d099c3adc        12 days ago         182.7 MB
mysql               5.6                 f2e8d6c772c0        3 weeks ago         324.6 MB
```

说明：  
* REPOSITORY：表示镜像的仓库源
* TAG：镜像的标签
* IMAGE ID：镜像ID
* CREATED：镜像创建时间
* SIZE：镜像大小

同一仓库源可以有多个 TAG，代表这个仓库源的不同个版本，如 ubuntu 仓库源里，有 15.10、14.04 等多个不同的版本，我们使用 REPOSITORY:TAG 来定义不同的镜像。所以，我们如果要使用版本为15.10的ubuntu系统镜像来运行容器时，命令如下：
```shell
$ docker run -t -i ubuntu:15.10 /bin/bash 
root@d77ccb2e5cca:/#
```

2. 获取一个新镜像

当我们在本地主机上使用一个不存在的镜像时 Docker 就会自动下载这个镜像。如果我们想预先下载这个镜像，我们可以使用 docker pull 命令来下载它。
```shell
$ docker pull ubuntu:13.10
13.10: Pulling from library/ubuntu
6599cadaf950: Pull complete 
23eda618d451: Pull complete 
f0be3084efe9: Pull complete 
52de432f084b: Pull complete 
a3ed95caeb02: Pull complete 
Digest: sha256:15b79a6654811c8d992ebacdfbd5152fcf3d165e374e264076aa435214a947a3
Status: Downloaded newer image for ubuntu:13.10
```

3. 查找镜像

我们可以从 Docker Hub 网站来搜索镜像，Docker Hub 网址为： https://hub.docker.com/。我们也可以使用 docker search 命令来搜索镜像。比如我们需要一个 httpd 的镜像来作为我们的 web 服务。我们可以通过 docker search 命令搜索 httpd 来寻找适合我们的镜像。
```shell
$ docker search httpd
NAME                                    DESCRIPTION                                     STARS     OFFICIAL   AUTOMATED
httpd                                   The Apache HTTP Server Project                  3693      [OK]       
centos/httpd-24-centos7                 Platform for running Apache httpd 2.4 or bui…   40                   
centos/httpd                                                                            34                   [OK]
polinux/httpd-php                       Apache with PHP in Docker (Supervisor, CentO…   5                    [OK]
solsson/httpd-openidc                   mod_auth_openidc on official httpd image, ve…   2                    [OK]
hypoport/httpd-cgi                      httpd-cgi                                       2                    [OK]
```

说明：  
* NAME: 镜像仓库源的名称
* DESCRIPTION: 镜像的描述
* OFFICIAL: 是否 docker 官方发布
* stars: 类似 Github 里面的 star，表示点赞、喜欢的意思。
* AUTOMATED: 自动构建。

4. 拖取镜像

我们决定使用上图中的 httpd 官方版本的镜像，使用命令 docker pull 来下载镜像。
```shell
$ docker pull httpd
Using default tag: latest
latest: Pulling from library/httpd
8b87079b7a06: Pulling fs layer 
a3ed95caeb02: Download complete 
0d62ec9c6a76: Download complete 
a329d50397b9: Download complete 
ea7c1f032b5c: Waiting 
be44112b72c7: Waiting
```

5. 删除镜像

```shell
$ docker rmi hello-world
```

6. 创建镜像

当我们从 docker 镜像仓库中下载的镜像不能满足我们的需求时，我们可以通过以下两种方式对镜像进行更改。
* 从已经创建的容器中更新镜像，并且提交这个镜像
* 使用 [Dockerfile](#Dockerfile) 文件来创建一个新的镜像

我们使用命令 docker build ，从零开始来创建一个新的镜像。为此，我们需要创建一个 Dockerfile 文件，其中包含一组指令来告诉 Docker 如何构建我们的镜像。
```Dockerfile
FROM    centos:6.7
MAINTAINER      Fisher "fisher@sudops.com"

RUN     /bin/echo 'root:123456' |chpasswd
RUN     useradd runoob
RUN     /bin/echo 'runoob:123456' |chpasswd
RUN     /bin/echo -e "LANG=\"en_US.UTF-8\"" >/etc/default/local
EXPOSE  22
EXPOSE  80
CMD     /usr/sbin/sshd -D
```

每一个指令都会在镜像上创建一个新的层，每一个指令的前缀都必须是大写的。第一条FROM，指定使用哪个镜像源RUN 指令告诉docker 在镜像内执行命令，安装了什么。。。然后，我们使用 Dockerfile 文件，通过 docker build 命令来构建一个镜像。
```shell
$ docker build -t runoob/centos:6.7 .
Sending build context to Docker daemon 17.92 kB
Step 1 : FROM centos:6.7
 ---&gt; d95b5ca17cc3
Step 2 : MAINTAINER Fisher "fisher@sudops.com"
 ---&gt; Using cache
 ---&gt; 0c92299c6f03
Step 3 : RUN /bin/echo 'root:123456' |chpasswd
 ---&gt; Using cache
 ---&gt; 0397ce2fbd0a
Step 4 : RUN useradd runoob
......
```

说明：  
* -t ：指定要创建的目标镜像名
* . ：Dockerfile 文件所在目录，可以指定Dockerfile 的绝对路径

7. 更新镜像

更新镜像之前，我们需要使用镜像来创建一个容器。 
```shell
$ docker run -t -i ubuntu:15.10 /bin/bash
```

在运行的容器内使用 apt-get update 命令进行更新。在完成操作之后，输入 exit 命令来退出这个容器。此时 ID 为 e218edb10161 的容器，是按我们的需求更改的容器。我们可以通过命令 docker commit 来提交容器副本。
```shell
$ docker commit -m="has update" -a="xxx" e218edb10161 runoob/ubuntu:v2
sha256:70bf1840fd7c0d2d8ef0a42a817eb29f854c1af8f7c59fc03ac7bdee9545aff8
```

说明：  
* -m: 提交的描述信息
* -a: 指定镜像作者
* e218edb10161：容器 ID
* runoob/ubuntu:v2: 指定要创建的目标镜像名

8. 设置镜像标签

我们可以使用 docker tag 命令，为镜像添加一个新的标签。
```shell
$ docker tag 860c279d2fec runoob/centos:dev
```
docker tag 镜像ID，这里是 860c279d2fec ,用户名称、镜像源名(repository name)和新的标签名(tag)。使用 docker images 命令可以看到，ID为860c279d2fec的镜像多一个标签。

## 容器连接

### 网络端口映射
我们创建了一个 python 应用的容器。
```shell
$ docker run -d -P training/webapp python app.py
fce072cc88cee71b1cdceb57c2821d054a4a59f67da6b416fceb5593f059fc6d
```

我们也可以使用 -p 标识来指定容器端口绑定到主机端口。
```shell
$ docker run -d -p 5000:5000 training/webapp python app.py
33e4523d30aaf0258915c368e66e03b49535de0ef20317d3f639d40222ba6bc0
```

两种方式的区别是:
* -P :是容器内部端口随机映射到主机的高端口。
* -p : 是容器内部端口绑定到指定的主机端口。

另外，我们可以指定容器绑定的网络地址，比如绑定 127.0.0.1。
```shell
$ docker run -d -p 127.0.0.1:5001:5000 training/webapp python app.py
95c6ceef88ca3e71eaf303c2833fd6701d8d1b2572b5613b5a932dfdfe8a857c
```

这样我们就可以通过访问 127.0.0.1:5001 来访问容器的 5000 端口。上面的例子中，默认都是绑定 tcp 端口，如果要绑定 UDP 端口，可以在端口后面加上 /udp。
```shell
$ docker run -d -p 127.0.0.1:5000:5000/udp training/webapp python app.py
6779686f06f6204579c1d655dd8b2b31e8e809b245a97b2d3a8e35abe9dcd22a
```

docker port 命令可以让我们快捷地查看端口的绑定情况。
```shell
$ docker port adoring_stonebraker 5000
127.0.0.1:5001
```

### Docker 容器互联
端口映射并不是唯一把 docker 连接到另一个容器的方法。docker 有一个连接系统允许将多个容器连接在一起，共享连接信息。docker 连接会创建一个父子关系，其中父容器可以看到子容器的信息。

下面创建一个新的Docker网络：
```shell
$ docker network create -d bridge test-net
```

说明：  
* -d: 参数指定 Docker 网络类型，有 bridge、overlay。其中 overlay 网络类型用于 Swarm mode

查看网络：
```shell
docker network ls
```

运行一个容器并连接到新建的test-net网络：
```shell
$ docker run -itd --name test1 --network test-net ubuntu /bin/bash
```

打开新的终端，再运行一个容器并加入到 test-net 网络:
```shell
$ docker run -itd --name test2 --network test-net ubuntu /bin/bash
```

下面通过 ping 来证明 test1 容器和 test2 容器建立了互联关系。如果 test1、test2 容器内中无 ping 命令，则在容器内执行以下命令安装 ping

如果你有多个容器之间需要互相连接，推荐使用 Docker Compose

### 配置DNS
我们可以在宿主机的 /etc/docker/daemon.json 文件中增加以下内容来设置全部容器的 DNS：
```json
{
  "dns" : [
    "114.114.114.114",
    "8.8.8.8"
  ]
}
```

设置后，启动容器的 DNS 会自动配置为 114.114.114.114 和 8.8.8.8。配置完，需要重启 docker 才能生效。查看容器的 DNS 是否生效可以使用以下命令，它会输出容器的 DNS 信息：
```shell
$ docker run -it --rm  ubuntu  cat etc/resolv.conf
```

如果只想在指定的容器设置 DNS，则可以使用以下命令：
```shell
$ docker run -it --rm -h host_ubuntu  --dns=114.114.114.114 --dns-search=test.com ubuntu
```

说明：  
* --rm：容器退出时自动清理容器内部的文件系统。
* -h HOSTNAME 或者 --hostname=HOSTNAME： 设定容器的主机名，它会被写到容器内的 /etc/hostname 和 /etc/hosts。
* --dns=IP_ADDRESS： 添加 DNS 服务器到容器的 /etc/resolv.conf 中，让容器用这个服务器来解析所有不在 /etc/hosts 中的主机名。
* --dns-search=DOMAIN： 设定容器的搜索域，当设定搜索域为 .example.com 时，在搜索一个名为 host 的主机时，DNS 不仅搜索 host，还会搜索 host.example.com。  
如果在容器启动时没有指定 --dns 和 --dns-search，Docker 会默认用宿主主机上的 /etc/resolv.conf 来配置容器的 DNS

## Dockerfile
Dockerfile 是一个用来构建镜像的文本文件，文本内容包含了一条条构建镜像所需的指令和说明。

### 使用Dockerfile定制镜像
1. 下面以定制一个 nginx 镜像（构建好的镜像内会有一个 /usr/share/nginx/html/index.html 文件）

在一个空目录下，新建一个名为 Dockerfile 文件，并在文件内添加以下内容：
```dockerfile
FROM nginx
RUN echo '这是一个本地构建的nginx镜像' > /usr/share/nginx/html/index.html
```

2. FROM 和 RUN 指令的作用

FROM：定制的镜像都是基于 FROM 的镜像，这里的 nginx 就是定制需要的基础镜像。后续的操作都是基于 nginx。

RUN：用于执行后面跟着的命令行命令。有以下俩种格式：

shell 格式：
```dockerfile
RUN <命令行命令>
# <命令行命令> 等同于，在终端操作的 shell 命令。
```

exec 格式：
```dockerfile
RUN ["可执行文件", "参数1", "参数2"]
# 例如：
# RUN ["./test.php", "dev", "offline"] 等价于 RUN ./test.php dev offline
```

注意：Dockerfile 的指令每执行一次都会在 docker 上新建一层。所以过多无意义的层，会造成镜像膨胀过大。例如
```dockerfile
FROM centos
RUN yum install wget
RUN wget -O redis.tar.gz "http://download.redis.io/releases/redis-5.0.3.tar.gz"
RUN tar -xvf redis.tar.gz
## 以上执行会创建 3 层镜像。可简化为以下格式：
FROM centos
RUN yum install wget \
    && wget -O redis.tar.gz "http://download.redis.io/releases/redis-5.0.3.tar.gz" \
    && tar -xvf redis.tar.gz
```

3. 开始构建镜像

在 Dockerfile 文件的存放目录下，执行构建动作。以下示例，通过目录下的 Dockerfile 构建一个 nginx:v3（镜像名称:镜像标签）。
```shell
$ docker build -t nginx:v3 .
```

`.`表示上下文路径，上下文路径，是指 docker 在构建镜像，有时候想要使用到本机的文件（比如复制），docker build 命令得知这个路径后，会将路径下的所有内容打包。

解析：由于 docker 的运行模式是 C/S。我们本机是 C，docker 引擎是 S。实际的构建过程是在 docker 引擎下完成的，所以这个时候无法用到我们本机的文件。这就需要把我们本机的指定目录下的文件一起打包提供给 docker 引擎使用。

如果未说明最后一个参数，那么默认上下文路径就是 Dockerfile 所在的位置。

注意：上下文路径下不要放无用的文件，因为会一起打包发送给 docker 引擎，如果文件过多会造成过程缓慢。

### 指令详解
1. COPY

复制指令，从上下文目录中复制文件或者目录到容器里指定路径。

格式：
```dockerfile
COPY [--chown=<user>:<group>] <源路径1>...  <目标路径>
COPY [--chown=<user>:<group>] ["<源路径1>",...  "<目标路径>"]
```

`[--chown=<user>:<group>]`：可选参数，用户改变复制到容器内文件的拥有者和属组。

`<源路径>`：源文件或者源目录，这里可以是通配符表达式，其通配符规则要满足 Go 的 filepath.Match 规则。例如：
```dockerfile
COPY hom* /mydir/
COPY hom?.txt /mydir/
```

`<目标路径>`：容器内的指定路径，该路径不用事先建好，路径不存在的话，会自动创建。

2. ADD

ADD 指令和 COPY 的使用格类似（同样需求下，官方推荐使用 COPY）。功能也类似，不同之处如下：

*  ADD 的优点：在执行 <源文件> 为 tar 压缩文件的话，压缩格式为 gzip, bzip2 以及 xz 的情况下，会自动复制并解压到 <目标路径>。
*  ADD 的缺点：在不解压的前提下，无法复制 tar 压缩文件。会令镜像构建缓存失效，从而可能会令镜像构建变得比较缓慢。具体是否使用，可以根据是否需要自动解压来决定。

3. CMD

类似于 RUN 指令，用于运行程序，但二者运行的时间点不同:

* CMD 在docker run 时运行。
* RUN 是在 docker build。

作用：为启动的容器指定默认要运行的程序，程序运行结束，容器也就结束。CMD 指令指定的程序可被 docker run 命令行参数中指定要运行的程序所覆盖。

注意：如果 Dockerfile 中如果存在多个 CMD 指令，仅最后一个生效。

格式：
```dockerfile
CMD <shell 命令> 
CMD ["<可执行文件或命令>","<param1>","<param2>",...] 
CMD ["<param1>","<param2>",...]  # 该写法是为 ENTRYPOINT 指令指定的程序提供默认参数
```

推荐使用第二种格式，执行过程比较明确。第一种格式实际上在运行的过程中也会自动转换成第二种格式运行，并且默认可执行文件是 sh。

4. ENTRYPOINT

类似于 CMD 指令，但其不会被 docker run 的命令行参数指定的指令所覆盖，而且这些命令行参数会被当作参数送给 ENTRYPOINT 指令指定的程序。

但是, 如果运行 docker run 时使用了 --entrypoint 选项，将覆盖 CMD 指令指定的程序。

优点：在执行 docker run 的时候可以指定 ENTRYPOINT 运行所需的参数。

注意：如果 Dockerfile 中如果存在多个 ENTRYPOINT 指令，仅最后一个生效。

格式：
```dockerfile
ENTRYPOINT ["<executeable>","<param1>","<param2>",...]
```

可以搭配 CMD 命令使用：一般是变参才会使用 CMD ，这里的 CMD 等于是在给 ENTRYPOINT 传参，以下示例会提到。

示例：

假设已通过 Dockerfile 构建了 nginx:test 镜像：
```dockerfile
FROM nginx

ENTRYPOINT ["nginx", "-c"] # 定参
CMD ["/etc/nginx/nginx.conf"] # 变参 
```

不传参运行:
```shell
$ docker run  nginx:test
```

容器内会默认运行以下命令，启动主进程。
```shell
nginx -c /etc/nginx/nginx.conf
```

传参运行：
```shell
$ docker run  nginx:test -c /etc/nginx/new.conf
```

容器内会默认运行以下命令，启动主进程(/etc/nginx/new.conf:假设容器内已有此文件)
```shell
nginx -c /etc/nginx/new.conf
```

5. ENV

设置环境变量，定义了环境变量，那么在后续的指令中，就可以使用这个环境变量。

格式：
```dockerfile
ENV <key> <value>
ENV <key1>=<value1> <key2>=<value2>...
```

以下示例设置 NODE_VERSION = 7.2.0 ， 在后续的指令中可以通过 $NODE_VERSION 引用：
```dockerfile
ENV NODE_VERSION 7.2.0

RUN curl -SLO "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-x64.tar.xz" \
  && curl -SLO "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc"
```

6. ARG

构建参数，与 ENV 作用一致。不过作用域不一样。ARG 设置的环境变量仅对 Dockerfile 内有效，也就是说只有 docker build 的过程中有效，构建好的镜像内不存在此环境变量。

构建命令 docker build 中可以用 --build-arg <参数名>=<值> 来覆盖。

格式：
```dockerfile
ARG <参数名>[=<默认值>]
```

7. VOLUME

定义匿名数据卷。在启动容器时忘记挂载数据卷，会自动挂载到匿名卷。

作用：

* 避免重要的数据，因容器重启而丢失，这是非常致命的。
* 避免容器不断变大。

格式：
```dockerfile
VOLUME ["<路径1>", "<路径2>"...]
VOLUME <路径>
```

在启动容器 docker run 的时候，我们可以通过 -v 参数修改挂载点

8. EXPOSE

仅仅只是声明端口。

作用：

* 帮助镜像使用者理解这个镜像服务的守护端口，以方便配置映射。
* 在运行时使用随机端口映射时，也就是 docker run -P 时，会自动随机映射 EXPOSE 的端口。

格式：
```dockerfile
EXPOSE <端口1> [<端口2>...]
```

9. WORKDIR

指定工作目录。用 WORKDIR 指定的工作目录，会在构建镜像的每一层中都存在。（WORKDIR 指定的工作目录，必须是提前创建好的）。

docker build 构建镜像过程中的，每一个 RUN 命令都是新建的一层。只有通过 WORKDIR 创建的目录才会一直存在。

格式：
```dockerfile
WORKDIR <工作目录路径>
```

10. USER

用于指定执行后续命令的用户和用户组，这边只是切换后续命令执行的用户（用户和用户组必须提前已经存在）。

格式：
```dockerfile
USER <用户名>[:<用户组>]
```

11. HEALTHCHECK

用于指定某个程序或者指令来监控 docker 容器服务的运行状态。

格式：
```dockerfile
HEALTHCHECK [选项] CMD <命令>：设置检查容器健康状况的命令
HEALTHCHECK NONE：如果基础镜像有健康检查指令，使用这行可以屏蔽掉其健康检查指令

HEALTHCHECK [选项] CMD <命令> : 这边 CMD 后面跟随的命令使用，可以参考 CMD 的用法。
```

12. ONBUILD

用于延迟构建命令的执行。简单的说，就是 Dockerfile 里用 ONBUILD 指定的命令，在本次构建镜像的过程中不会执行（假设镜像为 test-build）。当有新的 Dockerfile 使用了之前构建的镜像 FROM test-build ，这时执行新镜像的 Dockerfile 构建时候，会执行 test-build 的 Dockerfile 里的 ONBUILD 指定的命令。

格式：
```dockerfile
ONBUILD <其它指令>
```

13. LABEL

LABEL 指令用来给镜像添加一些元数据（metadata），以键值对的形式，语法格式如下：
```dockerfile
LABEL <key>=<value> <key>=<value> <key>=<value> ...
```

## 参考

### 持久化
Docker的数据持久化主要有两种方式`bind mount`和`volume`。

Docker的数据持久化即使数据不随着container的结束而结束，数据存在于host机器上——要么存在于host的某个指定目录中（使用bind mount），要么使用docker自己管理的volume（/var/lib/docker/volumes下）。

1. bind mount

bind mount自docker早期便开始为人们使用了，用于将host机器的目录mount到container中。但是bind mount在不同的宿主机系统时不可移植的，比如Windows和Linux的目录结构是不一样的，bind mount所指向的host目录也不能一样。这也是为什么bind mount不能出现在Dockerfile中的原因，因为这样Dockerfile就不可移植了。

将host机器上当前目录下的host-data目录mount到container中的/container-data目录：
```shell
docker run -it -v $(pwd)/host-dava:/container-data alpine sh
```

注意：
* host机器的目录路径必须为全路径(准确的说需要以/或~/开始的路径)
* 如果host机器上的目录不存在，docker会自动创建该目录
* 如果container中的目录不存在，docker会自动创建该目录
* 如果container中的目录已经有内容，那么docker会使用host上的目录将其覆盖掉

2. volume

volume也是绕过container的文件系统，直接将数据写到host机器上，只是volume是被docker管理的，docker下所有的volume都在host机器上的指定目录下/var/lib/docker/volumes。

将my-volume挂载到container中的/mydata目录：
```shell
docker run -it -v my-volume:/mydata alpine sh
```

然后可以查看到给my-volume的volume：
```shell
docker volume inspect my-volume
[
    {
        "CreatedAt": "2018-03-28T14:52:49Z",
        "Driver": "local",
        "Labels": null,
        "Mountpoint": "/var/lib/docker/volumes/my-volume/_data",
        "Name": "my-volume",
        "Options": {},
        "Scope": "local"
    }
]
```

可以看到，volume在host机器的目录为`/var/lib/docker/volumes/my-volume/_data`。此时，如果my-volume不存在，那么docker会自动创建my-volume，然后再挂载。

也可以不指定host上的volume：
```shell
docker run -it -v /mydata alpine sh
```

此时docker将自动创建一个匿名的volume，并将其挂载到container中的/mydata目录。匿名volume在host机器上的目录路径类似于：`/var/lib/docker/volumes/300c2264cd0acfe862507eedf156eb61c197720f69e7/_data`。

除了让docker帮我们自动创建volume，我们也可以自行创建：
```shell
docker volume create my-volume-2
```

然后将这个已有的my-volume-2挂载到container中:
```shell
docker run -it -v my-volume-2:/mydata alpine sh
```

需要注意的是，与bind mount不同的是，如果volume是空的而container中的目录有内容，那么docker会将container目录中的内容拷贝到volume中，但是如果volume中已经有内容，则会将container中的目录覆盖。请参考[这里](https://medium.com/@yaofei/docker-volume-what-i-learned-27134081d6d9)。

3. Dockerfile中的VOLUME

在Dockerfile中，我们也可以使用VOLUME指令来申明contaienr中的某个目录需要映射到某个volume：
```dockerfile
VOLUME /foo
```

这表示，在docker运行时，docker会创建一个匿名的volume，并将此volume绑定到container的/foo目录中，如果container的/foo目录下已经有内容，则会将内容拷贝的volume中。也即，Dockerfile中的`VOLUME /foo`与`docker run -v /foo alpine`的效果一样。

Dockerfile中的VOLUME使每次运行一个新的container时，都会为其自动创建一个匿名的volume，如果需要在不同container之间共享数据，那么我们依然需要通过`docker run -it -v my-volume:/foo`的方式将/foo中数据存放于指定的my-volume中。因此，VOLUME /foo在某些时候会产生歧义，如果不了解的话将导致问题。
