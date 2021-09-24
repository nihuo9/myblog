---
title: "Go语言grpc入门"
subtitle: ""
date: 2021-06-06T18:43:52+08:00
lastmod: 2021-06-06T18:43:52+08:00
draft: false
author: ""
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["grpc", "Go"]
categories: ["Go"]

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

本文主要是借鉴了《Go语言高级编程》的内容，其中有些内容已经过时了，需要做些修改
<!--more-->
## Protobuf

使用指导看这里：[proto3指导](https://oshirisu.site/posts/proto3%E6%8C%87%E5%AF%BC/)

## gRPC入门
从Protobuf的角度看，gRPC只不过是一个针对服务接口生成代码的生成器。接下来我们来看看如何使用gRPC。  
1. 安装protoc
```shell
# 1.
$ apt install -y protobuf-compiler
$ protoc --version #确保版本号是3+

# 2.
$ wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
$ unzip protoc-3.15.8-linux-x86_64.zip -d $HOME/.local
$ export PATH="$PATH:$HOME/.local/bin"
```

2. 安装Go的protocol编译器插件
```shell
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

3. 编写proto文件
首先创建hello.proto文件，定义HelloService接口：
```proto
syntax = "proto3";
package hello;

option go_package = "grpctest/hello";

message String {
	string value = 1;
}

service HelloService {
	rpc Hello(String) returns (String);
}
```

4. 编译proto文件
```shell
# paths=source_relative表示输出文件同输入文件在同一个文件路径下
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative hello.proto
```

5. 服务端代码
```Go
type HelloServiceImpl struct {
	hello.UnimplementedHelloServiceServer
}

func (h *HelloServiceImpl) Hello(ctx context.Context, request *hello.String) (reply *hello.String, err error) {
	reply = &hello.String{Value: "hello," + request.GetValue() }
	return
}

func main() {
	grpcServer := grpc.NewServer()
	hello.RegisterHelloServiceServer(grpcServer, new(HelloServiceImpl))

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer.Serve(lis)
}
```
gRPC通过`context.Context`参数，为每个方法调用提供了上下文支持，客户端在调用方法时，可以通过可选的`grpc.CallOption`类型的参数提供额外的上下文信息。

6. 客户端代码
```Go
func main() {
	conn, err := grpc.Dial(":1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := hello.NewHelloServiceClient(conn)
	reply, err := client.Hello(context.Background(), &hello.String{Value: "grpc"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply.GetValue())
}
```
gRPC和标准库的RPC框架有一个区别，即gRPC生成的接口并不支持异步调用。不过，我们可以在多个Groutine之间安全地共享gRPC底层的`HTTP/2`连接，因此可以通过在另一个Goroutine阻塞调用的方法模拟异步调用。

### gRPC流
RPC是远程过程调用，因此每次掉用的函数参数和返回值不能太大，否则将严重影响每次掉用的响应时间。因此传统的RPC方法调用对上传和下载较大数据量的场景并不合适。同时传统RPC模式也不适合时间不确定的订阅和发布模式。为此，gRPC框架提供了流的特性。  
流有方向，从客户端到服务器，或者从服务器到客户端，我们先使用双向流来定义一个`Channel()`方法：
```proto
service HelloService {
	rpc Hello(String) returns (String);
	rpc Channel(stream String) returns (stream String);
}
```
重新编译proto文件后可以看到`HelloServiceServer`接口的变化:
```Go
type HelloServiceServer interface {
	Hello(context.Context, *String) (*String, error)
	Channel(HelloService_ChannelServer) error
	mustEmbedUnimplementedHelloServiceServer()
}
```
多了一个`Channel(HelloService_ChannelServer) error`方法的要求，其中`HelloService_ChannelServer`是一个接口：
```Go
type HelloService_ChannelServer interface {
	Send(*String) error
	Recv() (*String, error)
	grpc.ServerStream
}
```
其中`Send`用来发送数据，`Recv`用来接收数据。客户端的接口`HelloServiceClient`也有变化:
```Go
type HelloServiceClient interface {
	Hello(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
	Channel(ctx context.Context, opts ...grpc.CallOption) (HelloService_ChannelClient, error)
}
```
其中新增加了一个方法`Channel`，这个方法会返回`HelloService_ChannelClient`接口
```Go
type HelloService_ChannelClient interface {
	Send(*String) error
	Recv() (*String, error)
	grpc.ClientStream
}
```
与服务器一样，`Send`发送数据，`Recv`接收数据，我们很容易想到所谓的流是什么意思。现在来我们来根据这样一个对应关系来实现流式RPC服务。首先在服务端的实现中新增一个方法：
```Go
func (h *HelloServiceImpl) Channel(stream hello.HelloService_ChannelServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		reply := &hello.String{Value : "hello:" + args.GetValue() }

		err = stream.Send(reply)
		if err != nil {
			return err
		}
	}
}
```
服务端在循环中接收客户端发来的请求，如果遇到`io.EOF`表示客户端关闭连接，生成返回的数据通过流发送回客户端，双向流数据的发送和接收是完全独立的行为。  

客户端代码如下：
```Go
func main() {
	conn, err := grpc.Dial(":1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := hello.NewHelloServiceClient(conn)
	stream, err := client.Channel(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	
	// 发送请求
	go func() {
		for {
			if err := stream.Send(&hello.String{Value: "stream"}); err != nil {
				log.Fatal(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// 接收响应
	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println(reply.GetValue())
	}
}
```
首先需要调用`Channel`方法获取到流对象，然后我们通过开启一个goroutine来发送请求，同时在主goroutine中读取响应。

## 发布和订阅模式
发布和订阅是一个常见的设计模式，下面是基于pubsub包实现的本地发布和订阅代码：
```Go
import (
	"github.com/docker/docker/pkg/pubsub"
)

func main() {
	p := pubsub.NewPublish(100 * time.Millisecond, 10)

	golang := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "golang:") {
				return true
			}
		}
		return false
	})

	docker := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "docker:") {
				return true
			}
		}
		return false
	})

	go p.Publish("hi")
	go p.Publish("golang: https://golang.org")
	go p.Publish("docker: https://www.docker.com/")
	time.Sleep(1)

	go func() {
		fmt.Println("golang topic:", <-golang)
	}()

	go func() {
		fmt.Println("docker topic:", <-docker)
	}()

	// wait for other goroutine
	// ...
}
```
其中`pubsub.NewPublisher`构造一个发布对象，`p.SubscribeTopic()`可以通过函数筛选感兴趣的主题进行订阅。  
现在尝试基于gRPC和pubsub包，提供一个跨网络的发布和订阅系统。首先通过Protobuf定义一个发布和订阅的服务接口：
```proto
service PubsubService {
	rpc Publish(String) returns (String);
	rpc Subscribe(String) returns (stream String);
}
```
接着在服务端实现服务：
```Go
// import xpubsub "github.com/docker/docker/pkg/pubsub"

type PubsubServiceImpl struct {
	pubsub.UnimplementedPubsubServiceServer
	pub		*xpubsub.Publisher
}

func NewPubsubServiceImpl() *PubsubServiceImpl {
	return &PubsubServiceImpl{
		pub: xpubsub.NewPublisher(100*time.Millisecond, 10),
	}
}

func (p *PubsubServiceImpl)Publish(c context.Context, req *pubsub.String) (*pubsub.String, error) {
	if p.pub == nil {
		return nil, errors.New("Publish: No Publisher")
	}
	p.pub.Publish(req.GetValue())
	return &pubsub.String{Value: fmt.Sprintf("Publish: %s", req.GetValue())}, nil
}

func (p *PubsubServiceImpl)Subscribe(req *pubsub.String, stream pubsub.PubsubService_SubscribeServer) error {
	ch := p.pub.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, req.GetValue()) {
				return true
			}
		}
		return false
	})

	for v := range ch {
		if err := stream.Send(&pubsub.String{Value: v.(string)}); err != nil {
			return err
		}
	}

	return nil
}
```
客户端实现：
```Go
var methodName = flag.String("m", "", "Method name. Publish/Subscribe")
var methodArg = flag.String("arg", "", "Method arg. A string value.")

func main() {
	flag.Parse()

	conn, err := grpc.Dial(":1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pubsub.NewPubsubServiceClient(conn)

	if *methodName == "Publish" {
		_, err = client.Publish(context.Background(), &pubsub.String{Value: *methodArg})
		if err != nil {
			log.Print(err)
		}
		fmt.Println("Publish:", *methodArg)
	} else if *methodName == "Subscribe" {
		stream, err := client.Subscribe(context.Background(), &pubsub.String{Value: *methodArg})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Subscribe:", *methodArg)

		for {
			reply, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			fmt.Println(reply.GetValue())
		}
	}
}
```
我们通过`flag`包来从命令行控制调用rpc方法。在一个终端开启客户端订阅：
```shell
$ ./client -m "Subscribe" -arg "golang:"
```
在另一个终端开启客户端进行发布：
```shell
$ ./client -m "Publish" -arg "golang: Hello Docker"
```
可以看到，第一个终端收到了服务端发来的订阅消息：
```shell
$ ./client -m "Subscribe" -arg "golang:"
Subscribe: golang:
golang: Hello Docker
```

## 证书认证
gRPC建立在HTTP/2协议之上对TLS提供了很好的支持，在之前我们创建客户端时通过`grpc.WithInsecure()`选项跳过了对服务器证书的验证。没有启用证书的gRPC服务和客户端进行的是明文通信，信息面临被任何第三方监听的风险，为了保证gRPC通信的安全，我们接下来将启用TLS来对通信进行加密。有关TLS的内容可以看文章[SSL与TLS原理详解](https://oshirisu.site/posts/ssl%E4%B8%8Etls%E5%8E%9F%E7%90%86%E8%AF%A6%E8%A7%A3/)。  
### 生成证书
1. 为了测试我们采用OpenSSL自签名证书的办法生成服务端证书已经客户端证书，首先生成CA（Certificate Authority）,就是一个根证书，命令如下：
```shell
openssl req -x509 -newkey rsa:2048 -keyout ca.key \
-subj "/CN=ca.grpc.com" -days 5000 -out ca.crt
```
`x509`：表示作为一个根证书自签名  
`newkey`：创建一个新的私钥，后面跟的就是私钥类型，这里是`rsa:2048`  
`keyout`：私钥名  
`subj`：项目的一些信息，`CN`表示通用名称  
`days`：有效日期  
`out`：输出证书名  

2. 生成服务端私钥，和客户端私钥，这一步也可以直接使用`newkey`子命令来实现：
```shell
openssl genrsa -out server.key 2048
openssl genrsa -out client.key 2048
```

3. 因为新的Go语言库要求使用SAN(Subject Alternative Name)证书，使用了 SAN 字段的 SSL 证书，可以扩展此证书支持的域名，使得一个证书可以支持多个不同域名的解析，所以我们需要生成SAN证书请求文件：
```shell
## server
openssl req -new -subj "/C=FJ/L=China/O=server/CN=server.grpc.io" \
-key server.key -out server.csr -config ./openssl.cnf -extensions v3_req

## client
openssl req -new -subj "/C=FJ/L=China/O=server/CN=client.grpc.io" \
-key client.key -out client.csr -config ./openssl.cnf -extensions v3_req
```
其中配置文件openssl.cnf，从`/etc/ssl/openssl.cnf`复制到当前文件夹下，并且需要做以下修改
```shell
#1：找到 [ CA_default ],打开 copy_extensions = copy
#2：找到[ req ],打开 req_extensions = v3_req # The extensions to add to a certificate request
#3：找到[ v3_req ],添加 subjectAltName = "DNS:example1.site, DNS:example2.site"
#或者subjectAltName = @alt_names，然后在文件中添加新的标签 [ alt_names ] , 和标签字段  
#DNS.1 = example1.site
#DNS.2 = example2.site
```

4. 使用CA签名SAN证书：
```shell
# server
openssl x509 -req -sha256 -CA ca.crt \
-CAkey ca.key -CAcreateserial -days 365 \
-in server.csr -out server.crt -extfile ./openssl.cnf -extensions v3_req

# client
openssl x509 -req -sha256 -CA ca.crt \
-CAkey ca.key -CAcreateserial -days 365 \
-in client.csr -out client.crt -extfile ./openssl.cnf -extensions v3_req
```

最后文件夹下的文件如下：
```shell
$ ls
ca.crt  ca.key  ca.srl  client.crt  client.csr  client.key  openssl.cnf  server.crt  server.csr  server.key
```

### 启用TLS
我们对之前实现的发布订阅系统增加TLS支持，首先服务器中代码变为如下：
```Go
func main() {
	certificate, err := tls.LoadX509KeyPair("../crt/server.crt", "../crt/server.key")
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../crt/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs: certPool,
	})

	grpcServer := grpc.NewServer(grpc.Creds(creds))

	...
}
```
代码中先加载服务端证书，然后读取CA证书构建`certPool`，通过服务端证书以及CA证书，启用了一个TLS服务。  
客户端代码也同服务端变化大体相同：
```Go
func main() {
	...

	certificate, err := tls.LoadX509KeyPair("../crt/client.crt", "../crt/client.key")
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../crt/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName: "server.grpc.io",
		RootCAs: certPool,
	})

	conn, err := grpc.Dial(":1234", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pubsub.NewPubsubServiceClient(conn)

	...
}
```
需要注意的是客户端需要在tls配置结构体中，设置连接的服务器名称。

## Token认证
gRPC还为每个gRPC方法调用提供了认证支持，这样就可以基于用户Token对不同的方法访问进行权限管理。  
要实现对每个gRPC方法进行认证，需要实现`grpc.PerRPCCredentials`接口：
```Go
// PerRPCCredentials defines the common interface for the credentials which need to
// attach security information to every RPC (e.g., oauth2).
type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing
	// tokens if required. This should be called by the transport layer on
	// each request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status
	// for the RPC. uri is the URI of the entry point for the request.
	// When supported by the underlying implementation, ctx can be used for
	// timeout and cancellation. Additionally, RequestInfo data will be
	// available via ctx to this call.
	// TODO(zhaoq): Define the set of the qualified keys instead of leaving
	// it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}
```
其中`GetRequestMetadata`方法返回认证需要的必要信息。`RequireTransportSecurity`方法表示是否要求底层使用安全链接。在真实环境中必需要求底层使用安全的链接，否则信息有泄露和被篡改的风险。  
我们可以在客户端中创建一个Authentication类型，用于实现用户名和密码的认证：
```Go
type Authentication struct {
	User		string
	Password	string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"user": a.User, "password": a.Password}, nil
}

func (a *Authentication) RequireTransportSecurity() bool {
	return false
}
```
在`GetRequestMetadata`方法中，返回认证的信息包括user和password两个信息。简单起便，`RequireTransportSecurity`不要求底层使用安全链接。客户端代码如下：
```Go
func main() {
	...

	auth := Authentication {
		User: "gopher",
		Password: "1234",
	}

	conn, err := grpc.Dial(":1234", grpc.WithInsecure(), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	...
}
```
因为这里没有启用安全链接，所以需要传入`grpc.WithInsecure()`选项表示忽略证书认证。  
然后在gRPC服务端的每个方法中通过`Auth`函数对客户端进行身份认证。
```Go
var authInfos = map[string]string {
	"gopher": "12345",
}

func Auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing credentials")
	}

	var appid string
	var appkey string

	if val, ok := md["user"]; ok {
		appid = val[0]
	}
	if val, ok := md["password"]; ok {
		appkey = val[0]
	}

	if val, ok := authInfos[appid]; ok {
		if appkey == val {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "invalid password")
	}
	return grpc.Errorf(codes.Unauthenticated, "invalid user name")
}
```
首先通过` metadata.FromIncomingContext`获取上下文中的元信息，然后取得`user`和`password`字段的信息，然后再根据服务端的存储信息进行对比，如果存在该用户且密码正确就认证通过，否则发送错误信息。  
然后我们就在每个方法中调用该函数进行认证，这里以pubsub服务为例：
```Go
func (p *PubsubServiceImpl)Publish(c context.Context, req *pubsub.String) (*pubsub.String, error) {
	...

	if err := Auth(c); err != nil {
		return nil, err
	}

	...
}

func (p *PubsubServiceImpl)Subscribe(req *pubsub.String, stream pubsub.PubsubService_SubscribeServer) error {
	if err := Auth(stream.Context()); err != nil {
		return err
	}

	...
}
```
这里需要注意的是流式的RPC方法需要通过`stream.Context`方法获取上下文。

## 截取器
gRPC中的`grpc.UnaryInterceptor`和`grpc.StreamInterceptor`选项分别对普通方法和流方法提供了截取器的支持。  
上面的两个选项都需要实现一个截取器函数：
```Go
// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server. info
// contains all the information of this RPC the interceptor can operate on. And handler is the wrapper
// of the service method implementation. It is the responsibility of the interceptor to invoke handler
// to complete the RPC.
type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)

// StreamServerInterceptor provides a hook to intercept the execution of a streaming RPC on the server.
// info contains all the information of this RPC the interceptor can operate on. And handler is the
// service method implementation. It is the responsibility of the interceptor to invoke handler to
// complete the RPC.
type StreamServerInterceptor func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error
```
我们在服务端实现这两个函数：
```Go
func unaryFilter(ctx context.Context, req interface{}, 
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {
	log.Println("unaryFilter:", info)
	return handler(ctx, req)
}

func streamFilter(srv interface{}, ss grpc.ServerStream, 
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("streamFilter:", info)
	return handler(srv, ss)
}
```
其中`UnaryServerInfo`和`StreamServerInfo`分别表示普通方法的信息和流方法的信息：
```Go
type UnaryServerInfo struct {
	// Server is the service implementation the user provides. This is read-only.
	Server interface{}
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
}

type StreamServerInfo struct {
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
	// IsClientStream indicates whether the RPC is a client streaming RPC.
	IsClientStream bool
	// IsServerStream indicates whether the RPC is a server streaming RPC.
	IsServerStream bool
}
```
我们在截取函数中首先打印改信息，然后再调用RPC方法，这个截取器就起到了一个日志的作用。  
要使用截取器函数，我们只需要在创建gRPC服务器时传入选项即可：  
```Go
grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryFilter), 
	grpc.StreamInterceptor(streamFilter))
```
另外gRPC也可以实现链式的截取器，只需要把`UnaryInterceptor`换成`ChainUnaryInterceptor`，`StreamInterceptor`换成`ChainStreamInterceptor`即可，然后可以在选项函数中传入多个截取器函数。  
我们在服务器端实现对客户端Token认证的截取器函数，其中Token认证见上一章节：
```Go
func unaryCheckFilter(ctx context.Context, req interface{}, 
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {
	if err := Auth(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func streamCheckFilter(srv interface{}, ss grpc.ServerStream, 
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := Auth(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}
```
然后，在创建gRPC服务器时传入选项：
```Go
grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(unaryFilter, unaryCheckFilter), 
grpc.ChainStreamInterceptor(streamFilter, streamCheckFilter))
```

## grpcurl工具
Protobuf本身具有反射功能，可以在运行时获取对象的Proto文件。gRPC同样也提供了一个名为`reflection`的反射包，用于为gRPC服务提供查询。
### 启用反射服务
reflection包中只有一个`Register()`函数，用于将`grpc.Server`注册到反射服务中。reflection包文档给出了简单的使用方法：
```Go
import "google.golang.org/grpc/reflection"

s := grpc.NewServer()
pb.RegisterYourOwnServer(s, &server{})

// Register reflection service on gRPC server.
reflection.Register(s)

s.Serve(lis)
```

### 查看服务列表
grpcurl是Go语言开源社区开发的工具，安装步骤如下：
```shell
go get github.com/fullstorydev/grpcurl/...
go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```
grpcurl最常使用的是`list`命令，用于获取服务或服务方法的列表。例如，`grpcurl localhost:1234 list`命令将获取本地1234端口上的gRPC服务的列表。在使用grpcurl时，需要通过参数`-cert`和`-key`设置公钥和私钥文件，对于没有启用TLS协议的gRPC服务，通过参数`-plaintext`忽略TLS证书的验证过程。如果是Unix套接字协议，则需要指定`-unix`参数。 
现在我们启动一个gRPC服务，服务的proto文件如下：
```proto
syntax = "proto3";
package hello;
import "google/api/annotations.proto";

option go_package = "grpctest/hello";

message String {
	string value = 1;
}

service HelloService {
	rpc Hello(String) returns (String) {
		option (google.api.http) = {
			get: "/get/{value}"
		};
	}
}
```
运行服务器，然后使用grpcurl查看服务列表：
```shell
$ grpcurl -plaintext localhost:1234 list
grpc.reflection.v1alpha.ServerReflection
hello.HelloService
```
其中`hello.HelloService`表示在hello包中定义了一个`HelloService`服务。`ServerReflection`服务是`grpc.reflection.v1alpha`包中注册的反射服务。通过`ServerReflectiona`服务可以查询包括本身在内的全面gRPC服务信息。

### 服务的方法列表和类型信息
继续使用list命令还可以查看服务的方法列表：
```shell
$ grpcurl -plaintext localhost:1234 list hello.HelloService
hello.HelloService.Hello
```
输出打印除了`HelloService`服务中有一个`Hello`方法。如果想看该方法在proto文件中的定义，可以使用`describe`子命令：
```shell
$ grpcurl -plaintext localhost:1234 describe hello.HelloService
hello.HelloService is a service:
service HelloService {
  rpc Hello ( .hello.String ) returns ( .hello.String );
}
```
另外我们也可以使用该命令查看proto文件中定义的类型信息：
```shell
grpcurl -plaintext localhost:1234 describe hello.String
hello.String is a message:
message String {
  string value = 1;
}
```

### 调用方法
下面的命令通过参数`-d`传入一个JSON字符串作为输入参数，调用了`HelloService.Hello`方法：
```shell
$ grpcurl -plaintext -d '{"value": "gopher"}' \
> localhost:1234 hello.HelloService/Hello
{
  "value": "hello,gopher"
}
```
另外如果`-d`后面跟的是一个`@`字符表示从标准输入读取输入：
```shell
$ grpcurl -plaintext -d @ localhost:1234 hello.HelloService/Hello <<EOF
> {"value": "gopher"}
> EOF
{
  "value": "hello,gopher"
}
```