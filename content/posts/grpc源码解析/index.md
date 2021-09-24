---
title: "grpc源码解析"
subtitle: ""
date: 2021-06-26T16:25:05+08:00
lastmod: 2021-06-26T16:25:05+08:00
draft: true
author: ""
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["Go", "grpc"]
categories: ["Go源码解析"]

hiddenFromHomePage: false
hiddenFromSearch: false
enableComment: true

resources:
- name: "featured-image"
  src: "featured-image.png"

featuredImage: ""
featuredImagePreview: ""
images: [""]

toc:
  enable: true
math:
  enable: false

license: ""
---

<!--more-->

## grpclog
### internal
内部包有两个变量:
```Go
// Logger is the logger used for the non-depth log functions.
var Logger LoggerV2
// DepthLogger is the logger used for the depth log functions.
var DepthLogger DepthLoggerV2
```
`Logger`作为grpclog包的默认日志接口，`DepthLogger`是一个带有层次的日志接口，另外内部包还定义了一个`PrefixLogger`结构体，通过设置`prefix`字段，可以在输出的时候带上前缀。

该包还实现了4个函数`InfoDepth`、`WarningDepth`、`ErrorDepth`、`FatalDepth`，举例说明：
```Go
func InfoDepth(depth int, args ...interface{}) {
	if DepthLogger != nil {
		DepthLogger.InfoDepth(depth, args...)
	} else {
		Logger.Infoln(args...)
	}
}
```
默认情况下会使用层次日志来输出日志信息，如果没有实现层次日志就用普通的日志输出。

### loggerv2
该文件中定义了`Loggerv2`接口，并且`loggerT`结构体实现了该接口：
```Go
type loggerT struct {
	m []*log.Logger
	v int
}
```
其中m储存了多个标准库的日志结构体指针，v代表当前日志层次。grpclog包初始化时会调用`SetLoggerV2(newLoggerV2())`来初始化一个`loggerT`作为默认日志。`newLoggerV2`函数如下：
```Go
func newLoggerV2() LoggerV2 {
	errorW := ioutil.Discard
	warningW := ioutil.Discard
	infoW := ioutil.Discard

	logLevel := os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL")
	switch logLevel {
	case "", "ERROR", "error": // If env is unset, set level to ERROR.
		errorW = os.Stderr
	case "WARNING", "warning":
		warningW = os.Stderr
	case "INFO", "info":
		infoW = os.Stderr
	}

	var v int
	vLevel := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if vl, err := strconv.Atoi(vLevel); err == nil {
		v = vl
	}
	return NewLoggerV2WithVerbosity(infoW, warningW, errorW, v)
}
```
该函数定义了3个值`ioutil.Discard`的变量，向该值读写数据实际上不会做任何事，然后通过读取环境变量`GRPC_GO_LOG_SEVERITY_LEVEL`来获取错误级别，默认错误级别是`errorW`，这时会让`errorW`变量的值变为`os.Stderr`表明输出到错误流，另外还会读取`GRPC_GO_LOG_VERBOSITY_LEVEL`来获取默认的日志层次。最后调用`NewLoggerV2WithVerbosity`来创建一个`loggerT`：
```Go
func NewLoggerV2WithVerbosity(infoW, warningW, errorW io.Writer, v int) LoggerV2 {
	var m []*log.Logger
	m = append(m, log.New(infoW, severityName[infoLog]+": ", log.LstdFlags))
	m = append(m, log.New(io.MultiWriter(infoW, warningW), severityName[warningLog]+": ", log.LstdFlags))
	ew := io.MultiWriter(infoW, warningW, errorW) // ew will be used for error and fatal.
	m = append(m, log.New(ew, severityName[errorLog]+": ", log.LstdFlags))
	m = append(m, log.New(ew, severityName[fatalLog]+": ", log.LstdFlags))
	return &loggerT{m: m, v: v}
}
```

### component
该文件实现了一个带组件名称的日志，实现的结构体如下：
```Go
type componentData struct {
	name string
}
```
`name`就是组件名称，使用该日志输出的时候会在前面先输出该名称，实现接口其实是内部包日志的一个包装。例如：
```Go
func (c *componentData) InfoDepth(depth int, args ...interface{}) {
	args = append([]interface{}{"[" + string(c.name) + "]"}, args...)
	grpclog.InfoDepth(depth+1, args...)
}
func (c *componentData) Info(args ...interface{}) {
	c.InfoDepth(1, args...)
}
```
我们可以通过`Component`函数来进行获取或者新建一个组件日志：
```Go
func Component(componentName string) DepthLoggerV2 {
	if cData, ok := cache[componentName]; ok {
		return cData
	}
	c := &componentData{componentName}
	cache[componentName] = c
	return c
}
```

## attributes
这个包定义了一个通用的key/value存储，被很多gRPC组件使用。
```Go
type Attributes struct {
	m map[interface{}]interface{}
}
```
主要数据结构就是`Attributes`由一个map组成用来索引数据。使用`New`用来创建一个新的`Attributes`
```Go
func New(kvs ...interface{}) *Attributes {
	if len(kvs)%2 != 0 {
		panic(fmt.Sprintf("attributes.New called with unexpected input: len(kvs) = %v", len(kvs)))
	}
	a := &Attributes{m: make(map[interface{}]interface{}, len(kvs)/2)}
	for i := 0; i < len(kvs)/2; i++ {
		a.m[kvs[i*2]] = kvs[i*2+1]
	}
	return a
}
```
该函数首先会判断给的参数是否是偶数个，然后通过循环给新建的map进行初始化赋值。  
`Attributes`只有两个方法：
```Go
func (a *Attributes) WithValues(kvs ...interface{}) *Attributes {
	if a == nil {
		return New(kvs...)
	}
	if len(kvs)%2 != 0 {
		panic(fmt.Sprintf("attributes.New called with unexpected input: len(kvs) = %v", len(kvs)))
	}
	n := &Attributes{m: make(map[interface{}]interface{}, len(a.m)+len(kvs)/2)}
	for k, v := range a.m {
		n.m[k] = v
	}
	for i := 0; i < len(kvs)/2; i++ {
		n.m[kvs[i*2]] = kvs[i*2+1]
	}
	return n
}

func (a *Attributes) Value(key interface{}) interface{} {
	if a == nil {
		return nil
	}
	return a.m[key]
}
```
其中`WithValues`用来储存新的k/v对，`Value`用来检索键值。

## credentials
credentials包实现了gRPC库支持的各种证书，gRPC库封装了客户端使用服务器进行身份验证和做出各种断言所需的所有状态。

### internal-credentials
```Go
type requestInfoKey struct{}
```
requestInfoKey被用来作为上下文中存储请求信息的键，于此对应的有两个函数，分别用来储存请求信息，和读取请求信息：
```Go
func NewRequestInfoContext(ctx context.Context, ri interface{}) context.Context {
	return context.WithValue(ctx, requestInfoKey{}, ri)
}
func RequestInfoFromContext(ctx context.Context) interface{} {
	return ctx.Value(requestInfoKey{})
}
```
另外还有一个`clientHandshakeInfoKey`与`requestInfoKey`具有相同的定义，用来处理客户端握手信息。

```Go
// file:spiffe.go
func SPIFFEIDFromState(state tls.ConnectionState) *url.URL {
	if len(state.PeerCertificates) == 0 || len(state.PeerCertificates[0].URIs) == 0 {
		return nil
	}
	return SPIFFEIDFromCert(state.PeerCertificates[0])
}

func SPIFFEIDFromCert(cert *x509.Certificate) *url.URL {
	if cert == nil || cert.URIs == nil {
		return nil
	}
	var spiffeID *url.URL
	for _, uri := range cert.URIs {
		if uri == nil || uri.Scheme != "spiffe" || uri.Opaque != "" || (uri.User != nil && uri.User.Username() != "") {
			continue
		}
		// From this point, we assume the uri is intended for a SPIFFE ID.
		if len(uri.String()) > 2048 {
			logger.Warning("invalid SPIFFE ID: total ID length larger than 2048 bytes")
			return nil
		}
		if len(uri.Host) == 0 || len(uri.Path) == 0 {
			logger.Warning("invalid SPIFFE ID: domain or workload ID is empty")
			return nil
		}
		if len(uri.Host) > 255 {
			logger.Warning("invalid SPIFFE ID: domain length larger than 255 characters")
			return nil
		}
		// A valid SPIFFE certificate can only have exactly one URI SAN field.
		if len(cert.URIs) > 1 {
			logger.Warning("invalid SPIFFE ID: multiple URI SANs")
			return nil
		}
		spiffeID = uri
	}
	return spiffeID
}
```
`SPIFFEIDFromState`获取tls连接中对方的证书，如果存在，使用`SPIFFEIDFromCert`从证书中解析SPIFFEID,
有关SPIFFEID的概念，请看：[](https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/)。

### 接口和基本类型定义
```Go
type PerRPCCredentials interface {
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	RequireTransportSecurity() bool
}
```
PerRPCCredentials是为需要附加安全信息到每个RPC的证书通用接口，`GetRequestMetadata`获取请求的元数据，`RequireTransportSecurity`指示是否证书要求传输层安全。

**安全等级**  
一共定义了3个安全等级：
```Go
const (
	InvalidSecurityLevel SecurityLevel = iota
	NoSecurity
	IntegrityOnly
	PrivacyAndIntegrity
)
```
依此是不安全、仅仅提供完整性保证、提供完整性和隐私保护。
安全等级被保存在`CommonAuthInfo`结构体中，该结构体会被其它的实现所内嵌。

`TransportCredentialsl`是gRPC拨号协议和支持的传输层安全协议的通用接口，定义如下:
```Go
type TransportCredentials interface {
	// ClientHandshake为客户端执行由rawConn上对应的身份认证协议指定的身份认证握手。它返回经过
	// 身份认证的连接和关于这个连接对应的认证信息。认证信息应该被嵌入CommonAuthInfo
	// 以返回关于证书的额外信息。实现必须使用提供的上下文来实现定时取消。如果返回的是一个临时性
	// 错误（例如io.EOF、context.DeadlineExceeded、err.Temporary()==true）gRPC将尝试重新
	// 连接。此外传递给这个调用的上下文中的ClientHandshakeInfo数据是可用的。
	// 如果返回的net.Conn已经关闭，必须把提供的net.Conn也关闭。
	ClientHandshake(context.Context, string, net.Conn) (net.Conn, AuthInfo, error)
	ServerHandshake(net.Conn) (net.Conn, AuthInfo, error)
	Info() ProtocolInfo
	Clone() TransportCredentials
	OverrideServerName(string) error
}
```

### tls
证书结构体如下：
```Go
type tlsCreds struct {
	config *tls.Config
}
```
只有一个tls配置的字段，该结构体实现了`TransportCredentials`接口，这里主要看`ClientHandshake`和`ServerHandshake`方法的实现：
```Go
// 实现TransportCredentials
func (c *tlsCreds) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (_ net.Conn, _ AuthInfo, err error) {
	// 如果使用多端点，使用本地配置避免清除服务器名称
	cfg := credinternal.CloneTLSConfig(c.config)
	if cfg.ServerName == "" {
		// 从authority分割主机名和端口号
		serverName, _, err := net.SplitHostPort(authority)
		if err != nil {
			// 如果authority没有端口号，就把服务器名设置为authority原本的值
			serverName = authority
		}
		// 设置服务器名
		cfg.ServerName = serverName
	}
	// 通过客户端连接和tls配置，返回一个新的TLS客户端连接
	conn := tls.Client(rawConn, cfg)
	errChannel := make(chan error, 1)
	go func() {
		// 进行握手
		errChannel <- conn.Handshake()
		close(errChannel)
	}()
	// 在这等待握手完成，或者上下文中给定的定时时间结束
	select {
	case err := <-errChannel:
		if err != nil {
			// 握手失败关闭连接
			conn.Close()
			return nil, nil, err
		}
	case <-ctx.Done():
		conn.Close()
		return nil, nil, ctx.Err()
	}
	tlsInfo := TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: CommonAuthInfo{
			// 安全等级设置为最高提供隐私和完整性保护
			SecurityLevel: PrivacyAndIntegrity,
		},
	}
	// 获取SPIFFEID
	// 下面两句可以合并为：tlsInfo.SPIFFEID = credinternal.SPIFFEIDFromState(conn.ConnectionState())
	id := credinternal.SPIFFEIDFromState(conn.ConnectionState())
	if id != nil {
		tlsInfo.SPIFFEID = id
	}
	// 尝试将原本连接和加密后的连接包装为syscallConn
	return credinternal.WrapSyscallConn(rawConn, conn), tlsInfo, nil
}

func (c *tlsCreds) ServerHandshake(rawConn net.Conn) (net.Conn, AuthInfo, error) {
	// 根据原本连接和tls配置返回一个新的服务端tls连接
	conn := tls.Server(rawConn, c.config)
	// 进行握手
	if err := conn.Handshake(); err != nil {
		conn.Close()
		return nil, nil, err
	}
	tlsInfo := TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: CommonAuthInfo{
			SecurityLevel: PrivacyAndIntegrity,
		},
	}
	id := credinternal.SPIFFEIDFromState(conn.ConnectionState())
	if id != nil {
		tlsInfo.SPIFFEID = id
	}
	return credinternal.WrapSyscallConn(rawConn, conn), tlsInfo, nil
}
```
这两个方法都会使用`tls`库将网络连接转换为tls连接，然后通过`Handshake`方法进行握手，如果握手成功，会返回一个包装了原始连接和tls连接的新连接：
```Go
func WrapSyscallConn(rawConn, newConn net.Conn) net.Conn {
	// 以提供对底层文件描述符或句柄的访问
	sysConn, ok := rawConn.(syscall.Conn)
	if !ok {
		return newConn
	}
	return &syscallConn{
		Conn:    newConn,
		sysConn: sysConn,
	}
}
```
还会返回一个`AuthInfo`接口实现，具体类型是`TLSInfo`:
```Go
type TLSInfo struct {
	// TLS连接状态
	State tls.ConnectionState
	// 主要就是一个安全级别的信息
	CommonAuthInfo
	SPIFFEID *url.URL
}
```
`TLSInfo`AuthInfo接口实现：
```Go
func (t TLSInfo) AuthType() string {
	return "tls"
}
```
`TLSInfo`ChannelzSecurityInfo接口实现：
```Go
type TLSChannelzSecurityValue struct {
	ChannelzSecurityValue
	// 加密套件的名字
	StandardName      string
	// 本地证书
	LocalCertificate  []byte
	// 远程证书
	RemoteCertificate []byte
}

func (t TLSInfo) GetSecurityValue() ChannelzSecurityValue {
	v := &TLSChannelzSecurityValue{
		// 获取加密套件的名称
		StandardName: cipherSuiteLookup[t.State.CipherSuite],
	}
	if len(t.State.PeerCertificates) > 0 {
		// 如果连接对面发来证书，复制下
		v.RemoteCertificate = t.State.PeerCertificates[0].Raw
	}
	return v
}
```
创造客户端TLS证书可以根据CA证书和服务器名，也可以直接从CA文件和服务器名：
```Go
func NewClientTLSFromCert(cp *x509.CertPool, serverNameOverride string) TransportCredentials {
	return NewTLS(&tls.Config{ServerName: serverNameOverride, RootCAs: cp})
}

func NewClientTLSFromFile(certFile, serverNameOverride string) (TransportCredentials, error) {
	b, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	// 新建一个证书池
	cp := x509.NewCertPool()
	// 从PEM文件附加证书到证书池中
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}
	return NewTLS(&tls.Config{ServerName: serverNameOverride, RootCAs: cp}), nil
}
```
服务端要求服务器证书或者服务器证书文件以及私钥文件：
```Go
func NewServerTLSFromCert(cert *tls.Certificate) TransportCredentials {
	return NewTLS(&tls.Config{Certificates: []tls.Certificate{*cert}})
}

func NewServerTLSFromFile(certFile, keyFile string) (TransportCredentials, error) {
	// 需要服务端证书和私钥
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}}), nil
}
```
这4个函数最终都借助`NewTLS`函数来创造证书：
```Go
func NewTLS(c *tls.Config) TransportCredentials {
	tc := &tlsCreds{credinternal.CloneTLSConfig(c)}
	tc.config.NextProtos = credinternal.AppendH2ToNextProtos(tc.config.NextProtos)
	return tc
}
```
`AppendH2ToNextProtos`将http2协议添加到下一个协议字段。


## serviceconfig
该包定义了服务配置有关的接口：
```Go
type Config interface {
	isServiceConfig()
}

type LoadBalancingConfig interface {
	isLoadBalancingConfig()
}
```
其中`Config`是普通的配置接口，`LoadbalancingConfig`是负载均衡配置的接口，另外该包还定义了一个结构体：
```Go
type ParseResult struct {
	Config Config
	Err    error
}
```
该结构体作为解析配置的结果类型。要求Config或者Err两个字段有一个不为nil。

### internal
内部包中实现了负载均衡的配置结构`BalancerConfig`和RPC方法的配置`MethodConfig`。
```Go
type BalancerConfig struct {
	Name   string
	Config externalserviceconfig.LoadBalancingConfig
}
```
负载均衡的配置中`Name`表示该配置的名称，`Config`是真正的配置，该结构体有两个方法：
```Go
func (bc *BalancerConfig) MarshalJSON() ([]byte, error) {
	if bc.Config == nil {
		// If config is nil, return empty config `{}`.
		return []byte(fmt.Sprintf(`[{%q: %v}]`, bc.Name, "{}")), nil
	}
	c, err := json.Marshal(bc.Config)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`[{%q: %s}]`, bc.Name, c)), nil
}
```
MarshalJSON将名字作为map的键，配置作为map的值，把`BalancerConfig`编码为一个`[]map[string]json.RawMessage`
```Go
func (bc *BalancerConfig) UnmarshalJSON(b []byte) error {
	var ir intermediateBalancerConfig
	err := json.Unmarshal(b, &ir)
	if err != nil {
		return err
	}

	for i, lbcfg := range ir {
		...
		var (
			name    string
			jsonCfg json.RawMessage
		)
		for name, jsonCfg = range lbcfg {
		}

		// 同过名字取得负载均衡的建造器
		builder := balancer.Get(name)
		if builder == nil {
			continue
		}
		bc.Name = name

		// 看看该负载均衡器是否实现了ConfigParser
		parser, ok := builder.(balancer.ConfigParser)
		if !ok {
			...
			return nil
		}
		cfg, err := parser.ParseConfig(jsonCfg)
		if err != nil {
			...
			return
		}
		bc.Config = cfg
		return nil
	}
	return ...
}
```
`UnmarshalJSON`借助`intermediateBalancerConfig`（`[]map[string]json.RawMessage`的别名）从字节切片中恢复配置，因为`Config`字段是一个接口，我们必须要找到该配置的原本类型，所以先根据获取到的配置名称，从balancer包获取到负载均衡的建造器，然后看看该建造器是否实现了`ConfigParser`接口，如果实现了，就用`ParseConfig`方法来恢复配置。

方法配置如下：
```Go
type MethodConfig struct {
	// 指示多个RPC发送到此方法时是否应该等待默认情况下连接就绪
	WaitForReady *bool
	// 默认超时时间。实际使用的截止时间将是这里指定的值和应用程序通过gRPC客户端API设置的值的最小值。
	// 如果其中一个没有设置，那么将使用另一个。如果两者都没有设置，则RPC没有截止时间
	Timeout *time.Duration
	// 客户端到服务器一次请求允许的最大负载字节大小
	MaxReqSize *int
	MaxRespSize *int
	RetryPolicy *RetryPolicy
}
```
其中定义了一些方法的处理策略，其中`RetryPolicy`代表的是重试的策略，其结构体如下：
```Go
type RetryPolicy struct {
	// 最大重试次数
	MaxAttempts int

	// Exponential backoff parameters. The initial retry attempt will occur at
	// random(0, initialBackoff). In general, the nth attempt will occur at
	// random(0,
	//   min(initialBackoff*backoffMultiplier**(n-1), maxBackoff)).
	//
	// These fields are required and must be greater than zero.
	// 指数退避参数
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64

	// 能够进行重试的状态代码集合
	RetryableStatusCodes map[codes.Code]bool
}
```

## backoff
backoff包完成指数退避算法的工作。  
导出的包中主要定义了配置相关的内容：
```Go
type Config struct {
	// 第一次失败退避的时间
	BaseDelay time.Duration
	// 失败后重试的乘法因子应该大于1
	Multiplier float64
	// 退避随机化因子
	Jitter float64
	// 最大退避延迟
	MaxDelay time.Duration
}
```
默认配置如下：
```Go
var DefaultConfig = Config{
	BaseDelay:  1.0 * time.Second,
	Multiplier: 1.6,
	Jitter:     0.2,
	MaxDelay:   120 * time.Second,
}
```

### internal
退避算法的实现在内部包中，实现的结构体是`Exponential`，定义如下：
```Go
type Exponential struct {
	Config grpcbackoff.Config
}
```
结构体中只有一个配置字段，这个配置就是我们前面说的那个。这个结构体只有一个方法，实现了`Strategy`接口：
```Go
func (bc Exponential) Backoff(retries int) time.Duration {
	if retries == 0 {
		return bc.Config.BaseDelay
	}
	backoff, max := float64(bc.Config.BaseDelay), float64(bc.Config.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.Config.Multiplier
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + bc.Config.Jitter*(grpcrand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}
```
`retries`参数就是重试的次数，如果重试次数为0说明是第一次，那么直接返回配置中的基础延时，另外会根据重试次数让基础延时`Config.BaseDelay`与乘数因子`Config.Multiplier`进行多次相乘直到重试次数为0或者延时时间超过了最大延时上限，最后一步需要再让退避时间乘于一个随机因子，该随机因子为`1 + Config.Jitter * (rand * 2 - 1)`，其中rand来自`grpcrand`包，实际上是调用了`rand.New(rand.NewSource(time.Now().UnixNano())).Float64()`。