---
title: "解析Go语言gob库"
subtitle: ""
date: 2021-06-12T17:35:32+08:00
lastmod: 2021-06-12T17:35:32+08:00
draft: false
author: "nihuo"
description: ""

page:
    theme: "wide"

upd: ""
authorComment: ""

tags: ["Go", "gob"]
categories: ["Go源码解析"]

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

gob是Go语言类型编码的一种，在rpc中发挥着很重要的作用，本文简单的分析了Go语言gob库的代码实现。
<!--more-->
## 简述
gob库总的来说可以分成3个部分组成：“类型定义”：“编码器”、“解码器”，如下图所示
<div style="text-align:center">

![图1. gob文件组成](ch1_1.png)
</div>

## 类型定义
### userTypeInfo
```Go
type userTypeInfo struct {
	user        reflect.Type 	// 用户递交的类型
	base        reflect.Type 	// 解引用后的类型
	indir       int         	// 由user到base需要解引用的次数
	externalEnc int          // 使用的外部编码器 Gob、Binary、Text
	externalDec int          // 使用的外部解码器
	encIndir    int8         // 由user到实现编码器类型需要解引用的次数，-1代表取地址
	decIndir    int8         // 由user到实现解码器类型需要解引用的次数
}
```
userTypeInfo结构体表示的是用户类型的一些信息，包括类型本身的反射类型、基础类型（非引用类型）的反射类型、还有是否实现了外部的编解码器等。  
可以通过`validUserType`来从反射类型中获取userTypeInfo
```Go
// 代码有省略
func validUserType(rt reflect.Type) (*userTypeInfo, error) {
	ut := new(userTypeInfo)
	ut.base = rt
	ut.user = rt

	// 快慢指针检查是否存在循环引用
	slowpoke := ut.base
	for {
		pt := ut.base
		if pt.Kind() != reflect.Ptr { // 如果pt的类型不是指针就可以直接退出了
			break
		}
		ut.base = pt.Elem() // 解引用一次
		if ut.base == slowpoke {
			// 循环引用返回错误
			return nil, errors.New("can't represent recursive pointer type " + ut.base.String())
		}
		// ut.base每解引用两次，slowpoke解引用一次
		if ut.indir%2 == 0 {
			slowpoke = slowpoke.Elem()
		}
		ut.indir++
	}

	// 判断类型是否有实现编码接口
	if ok, indir := implementsInterface(ut.user, gobEncoderInterfaceType); ok {
		ut.externalEnc, ut.encIndir = xGob, indir
	} 

	if ok, indir := implementsInterface(ut.user, gobDecoderInterfaceType); ok {
		ut.externalDec, ut.decIndir = xGob, indir
	} 

	// 如果rt存在就返回rt，不存在就保存ut并返回ut
	ui, _ := userTypeCache.LoadOrStore(rt, ut)
	return ui.(*userTypeInfo), nil
}
```
validUserType会新建一个userTypeInfo然后反射类型的Elem方法来解引用获取基本类型，其中使用了快慢指针来避免出现指针的互相引用，原理就是快指针移动速度是慢指针的两倍如果出现闭环那么快指针将会在某一时刻与慢指针相等。另外该函数还会检查用户类型是否实现了某个编解码器的接口。

### gobType 
```Go
// 表示gob类型的id
type typeId int32

// gob类型接口
type gobType interface {
	id() typeId
	setId(id typeId)
	name() string
	string() string
	safeString(seen map[typeId]bool) string
}
```
typeId是int32的别名，实际作用是用来索引一个gob类型，用户自定义的类型id从64开始。gobType表示的是一个gob内部类型的接口定义，在文件内通过实现该接口定义了Go类型，例如：
```Go
type CommonType struct {
	Name string
	Id   typeId
}

func (t *CommonType) id() typeId { return t.Id }

func (t *CommonType) setId(id typeId) { t.Id = id }

func (t *CommonType) string() string { return t.Name }

func (t *CommonType) safeString(seen map[typeId]bool) string {
	return t.Name
}

func (t *CommonType) name() string { return t.Name }
```
这个CommonType是通用的接口实现，基础类型会直接在文件内会通过调用`bootstrapType`函数生成一个gob类型，并且在函数内部调用`setTypeId`给该类型设置一个typeId，不同类型之间的区别是名称和typeId。
除此之外还有几个特殊的结构体也会通过以`CommondType`匿名成员的方式实现gobType
```Go
type arrayType struct {
	CommonType
	// 数组元素的类型id
	Elem typeId
	Len  int
}

// 编码器类型
type gobEncoderType struct {
	CommonType
}

type mapType struct {
	CommonType
	Key  typeId
	Elem typeId
}

type sliceType struct {
	CommonType
	Elem typeId
}

// 结构体类型
type fieldType struct {
	Name string
	Id   typeId
}
type structType struct {
	CommonType
	Field []*fieldType
}
```
其它的所有类型（包括用户自定义）都可以通过以上gobType来构建一个新的gobType，例如在初始化过程中会通过`mustGetTypeInfo`->`buildTypeInfo`->`getBaseType`->`getType`->`newTypeObject`的流程生成内部使用的类型对应的gob类型及typeId。
其中newTypeObject函数定义如下，这里我们省略了很多类型，重点分析下slice和结构体：
```Go
func newTypeObject(name string, ut *userTypeInfo, rt reflect.Type) (gobType, error) {
if ut.externalEnc != 0 {
		return newGobEncoderType(name), nil
	}
	var err error
	var type0, type1 gobType

	switch t := rt; t.Kind() {
  case reflect.Bool:
		return tBool.gobType(), nil
  // 省略...
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return tBytes.gobType(), nil
		}
		st := newSliceType(name)
		types[rt] = st
		type0, err = getBaseType(t.Elem().Name(), t.Elem())
		if err != nil {
			return nil, err
		}
		st.init(type0)
		return st, nil

	case reflect.Struct:
		st := newStructType(name)
		types[rt] = st
		idToType[st.id()] = st
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !isSent(&f) {
				continue
			}
			typ := userType(f.Type).base
			tname := typ.Name()
			// Name只会返回reflect包内定义的类型名称，其它类型会返回空字符串
			if tname == "" {
				t := userType(f.Type).base
				tname = t.String()
			}
			gt, err := getBaseType(tname, f.Type)
			if err != nil {
				return nil, err
			}

			if gt.id() == 0 {
				setTypeId(gt)
			}
			st.Field = append(st.Field, &fieldType{f.Name, gt.id()})
		}
		return st, nil

	default:
		return nil, errors.New("gob NewTypeObject can't handle type: " + rt.String())
	}
}
```
可以看到该函数首先会判断是否实现了外部编码器如果实现了就转交给`newGobEncoderType`函数实现，其它情况会对用户类类型进行反射，根据不同类型种类来创建gobType，如果是基本类型就直接返回预定义的gobType，比如bool类型就返回`tBytes.gobType()`。如果类型种类是slice首先会判断是否是`reflect.Uint8`，因为`[]uint8`就相当于`[]byte`，而`[]byte`是作为基本类型实现的所以可以直接返回`tBytes.gobType()`，其他slice会调用`newSliceType`来创建一个新的`sliceType`,然后递归调用了`getbaseType`来获取slice元素的gobType，最后再调用sliceType的init方法，该方法设置了sliceType类型的typeId并设置了其Elem成员的TypeId。 
 
结构体也与slice类似，首先调用`newStructType`新建一个`structType`，然后通过反射的方法遍历结构体字段，`isSend`函数判断该字段是否导出且有效的，如果有效就通过`GetBaseType`得到每个字段的gobType，将获得的gobType的typeId附加到`structType`的Field成员上

### wireType & typeInfo
```Go
type wireType struct {
	ArrayT           *arrayType
	SliceT           *sliceType
	StructT          *structType
	MapT             *mapType
	GobEncoderT      *gobEncoderType
	BinaryMarshalerT *gobEncoderType
	TextMarshalerT   *gobEncoderType
}

type typeInfo struct {
	id      typeId
	encInit sync.Mutex    // 编码时在buildEncEngine中使用保护encoder
	encoder atomic.Value  // 储存*encEngine 
	wire    *wireType
}
```
wireType是一种特殊的gobType，用来表示我们发送或接收某个类型的信息。我们编码一个值时，会先发送一个(-id，wireType)对。wireType会将对应类型的结构体字段初始化，其它字段仍为空。  
typeInfo用来作为类型的信息，主要是保存了类型对应的wireType以及encEngine（编码引擎），encEngine保存了如何把类型编码的操作，具体可以看[编码器](#编码器)。typeInfo通过函数`buildTypeInfo`来创建：
```Go
func buildTypeInfo(ut *userTypeInfo, rt reflect.Type) (*typeInfo, error) {
	typeLock.Lock()
	defer typeLock.Unlock()

	// 锁住后进行二次检查，如果存在就直接返回
	if info := lookupTypeInfo(rt); info != nil {
		return info, nil
	}

	gt, err := getBaseType(rt.Name(), rt)
	if err != nil {
		return nil, err
	}
	info := &typeInfo{id: gt.id()}

	if ut.externalEnc != 0 {
		// 如果存在外部编码器
		// 得到rt的类型信息，会调用newGobEncoderType创建一个编码器类型
		userType, err := getType(rt.Name(), ut, rt)
		if err != nil {
			return nil, err
		}
		gt := userType.id().gobType().(*gobEncoderType)
		switch ut.externalEnc {
			// 选择编码器类型
		case xGob:
			info.wire = &wireType{GobEncoderT: gt}
		case xBinary:
			info.wire = &wireType{BinaryMarshalerT: gt}
		case xText:
			info.wire = &wireType{TextMarshalerT: gt}
		}
		rt = ut.user
	} else {
		t := info.id.gobType()
		switch typ := rt; typ.Kind() {
		case reflect.Array:
			info.wire = &wireType{ArrayT: t.(*arrayType)}
		case reflect.Map:
			info.wire = &wireType{MapT: t.(*mapType)}
		case reflect.Slice:
			if typ.Elem().Kind() != reflect.Uint8 {
				info.wire = &wireType{SliceT: t.(*sliceType)}
			}
		case reflect.Struct:
			info.wire = &wireType{StructT: t.(*structType)}
		}
	}

	newm := make(map[reflect.Type]*typeInfo)
	// 加载旧的map到m中
	m, _ := typeInfoMap.Load().(map[reflect.Type]*typeInfo)
	for k, v := range m {
		// 复制oldm到newm中
		newm[k] = v
	}
  // 新增一条类型信息
	newm[rt] = info
	// 保存回去
	typeInfoMap.Store(newm)
	return info, nil
}
```
buildTypeInfo在上锁后首先进行二次检查，如果发现类型信息已经存在就直接返回，另外会通过反射类型获得对应的gob类型并创建一个新的typeInfo给id赋值。如果存在外部编码器，那么这个类型信息的wireType就根据编码器类型来决定，因为要找到实现编码器的类型需要从用户给的类型开始，所以重新调用`getType`取得用户给的类型对应的gobType，将这个gobType赋值给wireType对应的字段。其他类型就根据其反射类型来设置wireType。最后需要把把这个类型加入到类型信息缓存中。

### 缓存
type文件中有很多缓存数据方便数据查找
```GO
var userTypeCache sync.Map  // map[reflect.Type]*userTypeInfo

var types = make(map[reflect.Type]gobType) // 普通类型到gobType的映射

var idToType = make(map[typeId]gobType) // typeId到gobType的映射

var builtinIdToType map[typeId]gobType // 内建的typeId到gob类型的映射

var typeInfoMap atomic.Value // 反射类型到typeInfo 原子值是一个指向map[reflect.Type]*typeInfo.的指针

// 解码用，类型名到反射类型的映射
var nameToConcreteType sync.Map // map[string]reflect.Type
// 编码用，反射类型到类型名的映射
var concreteTypeToName sync.Map // map[reflect.Type]string
```
userTypeCache主要用在`validUserType`中，缓存了反射类型到userTypeInfo的信息。

types用于`getType`以及`newTypeObject`中，缓存了普通类型到gobType的信息

idToType在`setTypeId`储存，typeId的`gobType`方法通过该缓存查找id对应的gobType，解码器`typeString`获取类型名时也会用到

builtinIdToType在初始化了内建类型后直接从idToType复制而来。在解码器编译解码指令时，用来根据typeId获取gobType。

typeInfoMap在`buildTypeInfo`中储存，`lookupTypeInfo`使用缓存通过反射类型查找类型信息

nameToConcreteType和concreteTypeToName前者是类型名到反射类型的缓存后者相反，通过函数`Register`或者`RegisterName`可以把类型的信息添加进缓存，在type.go文件中会为所有的内建类型调用`Register`注册信息。nameToConcreteType在解码接口时用来根据编码的名字得到类型，如果你发送的接口不是内建类型需要主动调用`Register`函数来进行注册。concreteTypeToName的作用就是在编码接口时根据反射类型获取类型名字，然后把这个名字作为字符串编码。

## 编码器
### 编码器及编码器状态
编码器的数据结构如下所示，w是一个Writer接口的slice，编码器最终通过该slice的最后一个Writer接口写入编码数据，sent用来记录该编码器发送过的类型，因为已经发送的类型在接收方也会有响应的记录所以不需要重复发送，countState是一个专门用来发送消息长度的编码器状态，freeList是空闲的编码器状态链表，编码时会先尝试从该链表取得一个编码器状态，如果取不到才会新建，byteBuf是用来作为编码时的缓存。编码器状态记录了编码器实际执行编码的情况，这里提下fieldnum，该字段用来表示编码的数据在结构体中的位置，非结构体数据的fieldnum为0，实际上编码器状态保存的fieldnum是当前字段相比于上个字段的差值，因为差值一般很小所以通过压缩0编码可以节省更多的空间。每一个编码器状态都需要一个encBuffer，encBuffer用来保存编码时的缓存数据，可以通过`writeByte`、`Write`、`WriteString`方法来写入不同的数据。
```Go
type Encoder struct {
	mutex      sync.Mutex              
	w          []io.Writer             // 编码数据写入的地方
	sent       map[reflect.Type]typeId // 已经发送过的类型
	countState *encoderState           // 专门用来发送消息长度的编码器状态
	freeList   *encoderState           // 空闲的encoderState的链表，避免再分配
	byteBuf    encBuffer               // 顶层的encoderState缓存
	err        error
}

type encoderState struct {
	enc      *Encoder				// 对应的编码器
	b        *encBuffer
	sendZero bool                 	// 是否发送零值元素
	fieldnum int                  	// 最后被写的字段号
	buf      [1 + uint64Size]byte 	// 被encoder使用的缓存用来编码整数用，避免分配
	next     *encoderState          
}

type encBuffer struct {
	data    []byte
	scratch [64]byte
}
```
编码器状态有3个方法，`encodeUint`、`encodeInt`、`update`，其中`encodeUint`用来编码无符号整数，当要编码的值小于0x7F就直接作为一个字节写入缓存中，其他情况下会通过binary编码的大端模式进行编码原来的整数值会被编码为一个[]byte，编码完成后会计算整数的前导0，将所有非0的字节的数量的负数转换为uint8写入最后一个0字节中，这也是编码开始时buf从索引1开始的原因。`encodeInt`会把有符号数转换为无符号数然后用`encodeUint`来进行编码，具体方法是把有符号数如果是负数得先取反然后左移1位，再把符号位写在最低位上，负数取反是方便解码时通过取反操作恢复符号位。`update`用来更新fieldnum。
```Go
func (state *encoderState) encodeUint(x uint64) {
	if x <= 0x7F {
		state.b.writeByte(uint8(x))
		return
	}

	binary.BigEndian.PutUint64(state.buf[1:], x)
	bc := bits.LeadingZeros64(x) >> 3      
	state.buf[bc] = uint8(bc - uint64Size) 

	state.b.Write(state.buf[bc : uint64Size+1])
}

func (state *encoderState) encodeInt(i int64) {
	var x uint64
	if i < 0 {
		x = uint64(^i<<1) | 1
	} else {
		x = uint64(i << 1)
	}

	state.encodeUint(x)
}

func (state *encoderState) update(instr *encInstr) {
	if instr != nil {
		state.encodeUint(uint64(instr.field - state.fieldnum))
		state.fieldnum = instr.field
	}
}
```
### 编码引擎
encOp声明了编码操作的格式，encInstr是一条编码指令其中op就是编码操作，field记录了该编码指令对应的结构体字段id，index则是结构体索引，indir表示的是在结构体中这个值需要解引用多少次，encEngine就是一组编码指令，用来具体的编码一个数据类型。
```Go
type encEngine struct {
	instr []encInstr
}

type encInstr struct {
	op    encOp
	field int   
	index []int 
	indir int   
}

type encOp func(i *encInstr, state *encoderState, v reflect.Value)
```
在encode.go中定义了一些基本数据类型的编码操作，我们来看几个例子
```Go
func floatBits(f float64) uint64 {
	// 将float64以uint64的形式表示 
	u := math.Float64bits(f)
	// 翻转，为了使得指数端在前使得整型浮点数更精密地发送
	return bits.ReverseBytes64(u)
}
func encFloat(i *encInstr, state *encoderState, v reflect.Value) {
	f := v.Float()
	if f != 0 || state.sendZero {
		bits := floatBits(f)
		state.update(i)
		state.encodeUint(bits)
	}
}

func encUint8Array(i *encInstr, state *encoderState, v reflect.Value) {
	b := v.Bytes()
	if len(b) > 0 || state.sendZero {
		state.update(i)
		state.encodeUint(uint64(len(b)))
		state.b.Write(b)
	}
}
```
encFloat编码一个浮点数，首先它会借助`math.Float64bits`来将浮点数转换为uint64，接着它会把这个整型值反转目的是使得浮点数的指数端在前使得编码后的值更小，在按uint编码前还会使用编码器状态的`update`更新下fieldnum，像浮点数这种非结构体类型该值为0。encUint8Array编码一个字节数组，当数组长度大于0或者允许发送零值时，该函数首先会更新编码器状态的fieldnum，然后把数组长度编码为一个无符号整数，最后直接通过编码器缓存的`Write`把整个数组写入进去。  

编码引擎由函数`compileEnc`来进行创建，对于非结构体和实现了外部编码器的类型，编码引擎的指令就是通过`encOpFor`获得的单独一条，而结构体会遍历其所有字段然后分别通过`encOpFor`创建操作，并把该字段的索引和字段号都写进指令中。注意这一条语句`engine.instr = append(engine.instr, encInstr{encStructTerminator, 0, nil, 0})`，每个结构体后面都会编码一个整数0。
```Go
func compileEnc(ut *userTypeInfo, building map[*typeInfo]bool) *encEngine {
	srt := ut.base
	engine := new(encEngine)
	seen := make(map[reflect.Type]*encOp)
	rt := ut.base
	if ut.externalEnc != 0 {
		rt = ut.user
	}
	if ut.externalEnc == 0 && srt.Kind() == reflect.Struct {
		for fieldNum, wireFieldNum := 0, 0; fieldNum < srt.NumField(); fieldNum++ {
			f := srt.Field(fieldNum)
			if !isSent(&f) {
				continue
			}
			op, indir := encOpFor(f.Type, seen, building)
			engine.instr = append(engine.instr, encInstr{*op, wireFieldNum, f.Index, indir})
			wireFieldNum++
		}
		if srt.NumField() > 0 && len(engine.instr) == 0 {
			errorf("type %s has no exported fields", rt)
		}
		engine.instr = append(engine.instr, encInstr{encStructTerminator, 0, nil, 0})
	} else {
		engine.instr = make([]encInstr, 1)
		op, indir := encOpFor(rt, seen, building)
		engine.instr[0] = encInstr{*op, singletonField, nil, indir}
	}
	return engine
}
```
`encOpFor`首先会尝试从表`encOpTable`获得类型的编码操作，如果获取不到就根据反射类型的种类来分别创建。
```Go
// 有删减
func encOpFor(rt reflect.Type, inProgress map[reflect.Type]*encOp, building map[*typeInfo]bool) (*encOp, int) {
	ut := userType(rt)
	if ut.externalEnc != 0 {
		return gobEncodeOpFor(ut)
	}
	if opPtr := inProgress[rt]; opPtr != nil {
		return opPtr, ut.indir
	}

	typ := ut.base
	indir := ut.indir
	k := typ.Kind()
	var op encOp

	if int(k) < len(encOpTable) {
		op = encOpTable[k]
	}
	if op == nil {
		inProgress[rt] = &op
		switch t := typ; t.Kind() {
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 {
				op = encUint8Array
				break
			}
			elemOp, elemIndir := encOpFor(t.Elem(), inProgress, building)
			helper := encSliceHelper[t.Elem().Kind()]
			op = func(i *encInstr, state *encoderState, slice reflect.Value) {
				if !state.sendZero && slice.Len() == 0 {
					return
				}
				state.update(i)
				state.enc.encodeArray(state.b, slice, *elemOp, elemIndir, slice.Len(), helper)
			}
		case reflect.Struct:
			getEncEngine(userType(typ), building)
			info := mustGetTypeInfo(typ)
			op = func(i *encInstr, state *encoderState, sv reflect.Value) {
				state.update(i)
				enc := info.encoder.Load().(*encEngine)
				state.enc.encodeStruct(state.b, enc, sv)
			}
		case reflect.Interface:
			op = func(i *encInstr, state *encoderState, iv reflect.Value) {
				if !state.sendZero && (!iv.IsValid() || iv.IsNil()) {
					return
				}
				state.update(i)
				state.enc.encodeInterface(state.b, iv)
			}
		}
	}
	if op == nil {
		errorf("can't happen: encode type %s", rt)
	}
	return &op, indir
}
```
我们以slice、struct、interface类型来进行说明

* slice：对于slice类型的数据来说首先会获取其元素的类型，如果元素类型是uint8那么可以直接通过`encUint8Array`来进行编码，其它情况会递归获取其元素的编码操作，然后获取一个slice编码的帮助函数，这个帮助函数由文件`encode.go`通过`go:generate go run encgen.go -output enc_helpers.go`的声明自动生成存放在文件`enc_helpers.go`中，主要是一些内建类型的slice和数组的编码函数，例如：
```Go
func encBoolSlice(state *encoderState, v reflect.Value) bool {
	slice, ok := v.Interface().([]bool)
	if !ok {
		return false
	}
	for _, x := range slice {
		if x != false || state.sendZero {
			if x {
				state.encodeUint(1)
			} else {
				state.encodeUint(0)
			}
		}
	}
	return true
}
```
获取到帮助函数后会创建一个匿名函数赋值给op，这个匿名函数会调用状态的`update`函数来更新fieldnum，并调用`(enc *Encoder) encodeArray`来进行真正的编码工作。
```Go
func (enc *Encoder) encodeArray(b *encBuffer, value reflect.Value, op encOp, elemIndir int, length int, helper encHelper) {
	state := enc.newEncoderState(b)
	defer enc.freeEncoderState(state)
	state.fieldnum = -1
	state.sendZero = true
	state.encodeUint(uint64(length))
	if helper != nil && helper(state, value) {
		return
	}

	for i := 0; i < length; i++ {
		elem := value.Index(i)
		if elemIndir > 0 {
			elem = encIndirect(elem, elemIndir)
			if !valid(elem) {
				errorf("encodeArray: nil element")
			}
		}
		op(nil, state, elem)
	}
}
```
该函数会先创建一个新的编码状态避免影响前一个编码状态的fieldnum，然后把数组的长度进行编码，然后看看helper函数是否能够完成编码工作，如果不能，就遍历元素然后执行元素的编码操作。

* 结构体：如果字段是结构体就需要递归地调用`getEncEngine`获取结构体字段的编码引擎，如果结构体字段的编码引擎成功创建会储存在对应类型信息的`encoder`字段上。接着会创建一个匿名函数赋值给op，工作流程同slice的变化是调用了`(enc *Encoder) encodeStruct`方法
```Go
func (enc *Encoder) encodeStruct(b *encBuffer, engine *encEngine, value reflect.Value) {
	if !valid(value) {
		return
	}
	state := enc.newEncoderState(b)
	defer enc.freeEncoderState(state)
	state.fieldnum = -1
	for i := 0; i < len(engine.instr); i++ {
		instr := &engine.instr[i]
		if i >= value.NumField() {
			instr.op(instr, state, reflect.Value{})
			break
		}
		field := value.FieldByIndex(instr.index)
		if instr.indir > 0 {
			field = encIndirect(field, instr.indir)
			if !valid(field) {
				conjiang
			}
		}
		instr.op(instr, state, field)
	}
}
```
方法中会遍历结构体的编码引擎然后分别执行编码操作，这里注意编码器状态的fieldnum被设置成-1，这样编码结构体第一个字段时，根据差值编码会将编码器状态的fieldnum设置成1，在编码完所有字段后还会执行一个编码整数0的操作作为结构体结束的标志。

* 接口：接口基本工作流程也一样，最终会调用`(enc *Encoder) encodeInterface`来进行编码
```Go
func (enc *Encoder) encodeInterface(b *encBuffer, iv reflect.Value) {
	elem := iv.Elem()
	if elem.Kind() == reflect.Ptr && elem.IsNil() {
		errorf("gob: cannot encode nil pointer of type %s inside interface", iv.Elem().Type())
	}
	state := enc.newEncoderState(b)
	state.fieldnum = -1
	state.sendZero = true
	if iv.IsNil() {
		state.encodeUint(0)
		return
	}

	ut := userType(iv.Elem().Type())
	namei, ok := concreteTypeToName.Load(ut.base)
	if !ok {
		errorf("type not registered for interface: %s", ut.base)
	}
	name := namei.(string)

	// 发送具体类型的名称到缓存中
	state.encodeUint(uint64(len(name)))
	state.b.WriteString(name)
	// 发送类型描述符和类型id
	enc.sendTypeDescriptor(enc.writer(), state, ut)
	enc.sendTypeId(state, ut)
	// 把缓存设置为写入对象
	enc.pushWriter(b)
	// 从缓存池中取出一个encBuffer作为新的缓存
	data := encBufferPool.Get().(*encBuffer)
	// 向缓存中填充长度所需字节
	data.Write(spaceForLength)
	// 把接口的动态值编码到缓存中
	enc.encode(data, elem, ut)
	if enc.err != nil {
		error_(enc.err)
	}
	// 恢复之前的写入对象
	enc.popWriter()
	// 把缓存data中的数据写入缓存b
	enc.writeMessage(b, data)
	data.Reset()
	encBufferPool.Put(data)
	if enc.err != nil {
		error_(enc.err)
	}
	enc.freeEncoderState(state)
}
```
这个函数操作多了一点，首先会根据`concreteTypeToName`查找类型名称，如果找到了就往缓存里写入这个名称，接着直接发送动态值的类型描述，然后又将类型id写入缓存中，之后把缓存设置为新的写入对象并且从编码缓存池中取出一个新的缓存，然后把接口的动态值编码到新的缓存中，恢复之前的写入对象，把新的缓存中的内容写进旧的缓存完成编码，总的来说，接口被编码为： 动态值的类型描述符 + name_len + name + 类型id + 动态值编码的长度 + 动态值。

### 编码流程
用户创建完一个编码器后，通过`Encode`方法进行编码，而`Encode`方法其实是调用了`EncodeValue`
```GO
func (enc *Encoder) EncodeValue(value reflect.Value) error {
	if value.Kind() == reflect.Invalid {
		return errors.New("gob: cannot encode nil value")
	}
	if value.Kind() == reflect.Ptr && value.IsNil() {
		panic("gob: cannot encode nil pointer of type " + value.Type().String())
	}

	enc.mutex.Lock()
	defer enc.mutex.Unlock()

	enc.w = enc.w[0:1]

	ut, err := validUserType(value.Type())
	if err != nil {
		return err
	}

	enc.err = nil
	enc.byteBuf.Reset()
	enc.byteBuf.Write(spaceForLength)
	state := enc.newEncoderState(&enc.byteBuf)

	enc.sendTypeDescriptor(enc.writer(), state, ut)
	enc.sendTypeId(state, ut)
	if enc.err != nil {
		return enc.err
	}

	enc.encode(state.b, value, ut)
	if enc.err == nil {
		enc.writeMessage(enc.writer(), state.b)
	}

	enc.freeEncoderState(state)
	return enc.err
}
```
该函数会先通过`validUserType`获取用户类型的信息，然后通过`sendTypeDescriptor`发送类型描述符，如果不是基础类型或者接口的话，其最终调用的是`(enc *Encoder) sendActualType`方法，所有类型都会被记录到编码器的`sent`字段中
```Go
func (enc *Encoder) sendActualType(w io.Writer, state *encoderState, ut *userTypeInfo, actual reflect.Type) (sent bool) {
	if _, alreadySent := enc.sent[actual]; alreadySent {
		return false
	}
	info, err := getTypeInfo(ut)
	if err != nil {
		enc.setError(err)
		return
	}

	// 发送（-id，type）
	state.encodeInt(-int64(info.id))
	enc.encode(state.b, reflect.ValueOf(info.wire), wireTypeUserInfo)
	enc.writeMessage(w, state.b)
	if enc.err != nil {
		return
	}

	enc.sent[ut.base] = info.id
	if ut.user != ut.base {
		enc.sent[ut.user] = info.id
	}
	switch st := actual; st.Kind() {
	case reflect.Struct:
		for i := 0; i < st.NumField(); i++ {
			// 结构体就发送每个导出字段的类型
			if isExported(st.Field(i).Name) {
				enc.sendType(w, state, st.Field(i).Type)
			}
		}
	case reflect.Array, reflect.Slice:
		// 数组发送元素类型
		enc.sendType(w, state, st.Elem())
	case reflect.Map:
		// map发送键和值的类型
		enc.sendType(w, state, st.Key())
		enc.sendType(w, state, st.Elem())
	}
	return true
}
```
该方法会发送一个(-id，wireType)的值，其值都来源于用户类型对应的typeInfo接着直接调用`(enc *Encoder) writeMessage`把缓存中的值发送出去，然后根据用户类型的种类来递归地发送子元素类型，例如结构体就会发送其每个字段的类型。  
发送完类型后，`EncodeValue`会发送用户类型对应的typeId，接着就可以进行编码了。
编码一个值时，会调用到`(enc *Encoder) encode`方法，该方法首先会通过以下流程来获取一个编码引擎：`getEncEngine`->`buildEncEngine`->`compileEnc`，然后根据是否是结构体和是否实现了外部编码器来进行编码。

* 非结构体：
```Go
func (enc *Encoder) encodeSingle(b *encBuffer, engine *encEngine, value reflect.Value) {
	state := enc.newEncoderState(b)
	// 编码完把state放回enc中
	defer enc.freeEncoderState(state)
	// 单独值的fieldnum是0
	state.fieldnum = singletonField
	// 不像结构体有分帧传输，单独的值即使为零值也要生成数据，因此设置sendZero
	state.sendZero = true
	// 从engine中取得指令
	instr := &engine.instr[singletonField]
	if instr.indir > 0 {
		// 如果需要取引用就调用encIndirect取得实际的值
		value = encIndirect(value, instr.indir)
	}
	// 判断值是否有效
	if valid(value) {
		// 调用该指令的操作
		instr.op(instr, state, value)
	}
}
```

*  结构体：
```Go
func (enc *Encoder) encodeStruct(b *encBuffer, engine *encEngine, value reflect.Value) {
	if !valid(value) {
		return
	}
	state := enc.newEncoderState(b)
	defer enc.freeEncoderState(state)
	state.fieldnum = -1
	for i := 0; i < len(engine.instr); i++ {
		instr := &engine.instr[i]
		if i >= value.NumField() {
			// 编码结束 执行encStructTerminator写入一个0
			instr.op(instr, state, reflect.Value{})
			break
		}
		field := value.FieldByIndex(instr.index)
		if instr.indir > 0 {
			field = encIndirect(field, instr.indir)
			if !valid(field) {
				continue
			}
		}
		instr.op(instr, state, field)
	}
}
```
编码完成后通过`writeMessage`发送缓存中的数据，完成。

我们编码一个结构体看看其输出是什么：
```Go
type stest struct {
	ID int
	Str string
} 

var data = stest{4, "hello"}
var bs bytes.Buffer
encoder := gob.NewEncoder(&bs)

err := encoder.Encode(data)
if err != nil {
	log.Println("enc:", err)
}
fmt.Println(bs.Bytes())
```
最后输出如下：
```shell
$ ./normal -t test3
[34 255 129 3 1 1 5 115 116 101 115 116 1 255 130 0 1 2 1 2 73 68 1 4 0 1 3 83 116 114 1 12 0 0 0 12 255 130 1 8 1 5 104 101 108 108 111 0]
```
根据我们上面的分析可以很容易地理解数据的含义，其中方括号是数据，圆括号里是解释：
```
[34](message length)
[255 129](id:-65)  
[3](wireType fieldnum delta) 
	[1](structType fieldnum delta) 
		[1](CommonType fieldnum delta) [5 115 116 101 115 116](str:"stest") 
		[1](CommonType fieldnum delta) [255 130](id:65)  
		0 
	[1](structType fieldnum delta) 
		[2](array length) 
			[1](fieldtype fieldnum delta) [2 73 68](str:"ID") 
			[1](fieldtype fieldnum delta) [4](id:2) 
			0 
			[1](fieldtype fieldnum delta) [3 83 116 114](str:"Str") 
			[1](fieldtype fieldnum delta) [12](id:6) 
			0 
	0 
0 

[12](message length) 
[255 130](id:65) 
[1](stest fieldnum delta) [8](int:4) 
[1](stest fieldnum delta) [5 104 101 108 108 111](str:"hello") 
0
```
这里解释下`255 129`怎么计算出id为-65的，首先按照整数的编码方式，第一个字节是是大于0x7F的，所以第一个字节应该是保存了编码字节数的负数，255转换成int8就是-1，也就是说后面一个字节是被编码的ID，我们把129
再解码为有符号数，也就是右移一位后取反，然后得到-65。另外单独的0表示的是结构体的结束。

## 解码器
### 解码器及解码器状态
解码器中r是数据读取的地方，buf用来缓存解码数据，wireType缓存了typeId到wireType的映射，decoderCache和ignorerCache一个是解码引擎的缓存，一个是跳过操作的缓存，跳过操作就是忽略某个数据的解码内容。freeList链接着空闲的解码器状态链表，countBuf是解码整数时用到的缓存。解码器状态我们主要关注一个解码缓存b和字段号fieldnum。  decBuffer作为解码器的缓存可以通过`Read`、`ReadByte`来读取数据，每次读取数据都会使得offset字段增加，offset就是当前读取字节位置的偏移量，另外可以通过`Drop`方法来抛弃一些数据，这在跳过操作中使用到。  
```Go
type Decoder struct {
	mutex        sync.Mutex                              
	r            io.Reader                               // 读取数据的地方
	buf          decBuffer                             	 // 解码缓存
	wireType     map[typeId]*wireType                    // 远程类型缓存
	decoderCache map[reflect.Type]map[typeId]**decEngine // 本地类型解码gob类型的解码引擎缓存
	ignorerCache map[typeId]**decEngine                  // 对某个类型的跳过操作的引擎缓存
	freeList     *decoderState                           // 空闲的解码状态
	countBuf     []byte                                  // 解码整数时用到的缓存
	err          error
}

type decoderState struct {
	dec		*Decoder
	b        *decBuffer
	fieldnum int          
	next     *decoderState 
}

type decBuffer struct {
	data   []byte
	offset int
}
```
解码器状态有3个方法，`decodeUint`、`decodeInt`、`getLength`，其中`decodeUint`解码一个uint64数据，会先从缓存中读取一个字节数据，如果该字节小于0x7F那么就可以直接将该字节转换为uint64然后返回，另外说明这个字节表示的一个整数编码的字节长度的负数，需要把该字节转换为有符号数然后取反，接着从缓存中取出数据然后遍历n个字节，按照大端模式恢复整数。`decodeInt`比较简单，首先解码出一个无符号数，然后根据最后一位是否是1来判断正负，如果是负数就在右移一位后取反，正数直接右移一位。`getLength`就是`decodeUint`的包装，只不过多了一些条件的判断，以防止数据的异常。
```Go
func (state *decoderState) decodeUint() (x uint64) {
	b, err := state.b.ReadByte()
	if err != nil {
		error_(err)
	}
	if b <= 0x7f {
		return uint64(b)
	}
	n := -int(int8(b))
	if n > uint64Size {
		error_(errBadUint)
	}
	buf := state.b.Bytes()
	if len(buf) < n {
		errorf("invalid uint data length %d: exceeds input size %d", n, len(buf))
	}
	for _, b := range buf[0:n] {
		x = x<<8 | uint64(b)
	}
	state.b.Drop(n)
	return x
}

func (state *decoderState) decodeInt() int64 {
	x := state.decodeUint()
	if x&1 != 0 {
		return ^int64(x >> 1)
	}
	return int64(x >> 1)
}

func (state *decoderState) getLength() (int, bool) {
	n := int(state.decodeUint())
	if n < 0 || state.b.Len() < n || tooBig <= n {
		return 0, false
	}
	return n, true
}
```

### 解码器引擎
decOp声明了解码操作的格式，decInstr是一条解码指令，其中op就是解码指令，field表示该指令对于与结构体的字段偏移，index则是用来索引结构体，ovfl储存了发生溢出时的错误信息。decEngine就是解码引擎，其中instr是一组解码指令，numInstr表示的其中有效指令的数量。
```Go
type decOp func(i *decInstr, state *decoderState, v reflect.Value)

type decInstr struct {
	op    decOp
	field int   
	index []int 
	ovfl  error 
}

type decEngine struct {
	instr    []decInstr
	numInstr int
}
```
与编码器类似，解码器在文件`decode.go`中也有基本类型的解码操作的定义，举几个例子
```Go
func float64FromBits(u uint64) float64 {
	v := bits.ReverseBytes64(u)
	return math.Float64frombits(v)
}
func float32FromBits(u uint64, ovfl error) float64 {
	v := float64FromBits(u)
	av := v
	if av < 0 {
		av = -av
	}
	if math.MaxFloat32 < av && av <= math.MaxFloat64 {
		error_(ovfl)
	}
	return v
}
func decFloat32(i *decInstr, state *decoderState, value reflect.Value) {
	value.SetFloat(float32FromBits(state.decodeUint(), i.ovfl))
}

func decUint8Slice(i *decInstr, state *decoderState, value reflect.Value) {
	n, ok := state.getLength()
	if !ok {
		errorf("bad %s slice length: %d", value.Type(), n)
	}
	if value.Cap() < n {
		value.Set(reflect.MakeSlice(value.Type(), n, n))
	} else {
		value.Set(value.Slice(0, n))
	}
	if _, err := state.b.Read(value.Bytes()); err != nil {
		errorf("error decoding []byte: %s", err)
	}
}
```
decFloat32解码一个float32的值，将从缓存中解码的uint64的值逆序，然后调用`math.Float64frombits`恢复成float64的值，并判断这个值是否对于float32来讲是溢出的，如果一切正常会返回这个flaot64的值。  
decUint8Slice解码一个[]uint8，首先会解码一个长度n，如果用户传入的值容量小于n需要为该值设置一个容量为n的新的slice，如果大于n则重设该slice的大小为n，接着就直接调用缓存的`Read`方法从缓存中读取n个字节到用户传入的slice中。

与编码操作不同解码操作还多了一类操作——忽略操作。对于基本类型主要依靠下面两个函数来完成，函数很简单，就是解码出数据后不管了。
```Go
func ignoreUint(i *decInstr, state *decoderState, v reflect.Value) {
	state.decodeUint()
}

func ignoreTwoUints(i *decInstr, state *decoderState, v reflect.Value) {
	state.decodeUint()
	state.decodeUint()
}
```

解码引擎由`compileDec`来进行创建，函数首先检查用户类型是否是结构体，或者是否实现了外部解码器，如果不是结构体或者有外部解码器，就转调`compileSingle`生成解码引擎，该函数会新建一个instr容量为1的引擎，同时检查远程的类型id和用户传入的类型的兼容性，对于基础类型的检查就是根据其反射类型判断，远程typeId是否与内建的对应类型的typeId相等，非基础类型则递归地检查其子元素（例如数组就检查其数组元素），这里注意一点如果是结构体的话是直接返回true的。兼容性通过后就调用`decOpFor`获取该类型的解码操作，完成引擎的创建。


对于结构体类型，`compileDec`首先会根据远程typeId检查是否是本地内建的几个结构体类型，另外就从解码器的`wireType`缓存中查找，获取到类型信息即wireStruct后就可以遍历其Field字段，该字段是一个[]*fieldType，里面存放了结构体的字段信息，如果用户传入的结构体中没有该字段，或者该字段是未导出的就跳过这时通过`decIgnoreOpFor`生成一个跳过操作，正常情况下需要检查兼容性之后也同样调用`decOpFor`获取该字段的解码操作，生成指令后存放在对应的fieldnum索引上面，同时用`numInstr`字段记录有效的指令数量。  

另外`compileIgnoreSingle`方法完成单个跳过解码引擎的生成工作，直接新建一个instr容量为1的解码引擎，然后通过`decIgnoreOpFor`来生成对应的跳过操作。
```Go
func (dec *Decoder) compileDec(remoteId typeId, ut *userTypeInfo) (engine *decEngine, err error) {
	defer catchError(&err)
	rt := ut.base
	srt := rt
	if srt.Kind() != reflect.Struct || ut.externalDec != 0 {
		return dec.compileSingle(remoteId, ut)
	}
	var wireStruct *structType

	if t, ok := builtinIdToType[remoteId]; ok {
		wireStruct, _ = t.(*structType)
	} else {
		wire := dec.wireType[remoteId]
		if wire == nil {
			error_(errBadType)
		}
		wireStruct = wire.StructT
	}
	if wireStruct == nil {
		errorf("type mismatch in decoder: want struct type %s; got non-struct", rt)
	}
	engine = new(decEngine)
	engine.instr = make([]decInstr, len(wireStruct.Field))
	seen := make(map[reflect.Type]*decOp)
	for fieldnum := 0; fieldnum < len(wireStruct.Field); fieldnum++ {
		wireField := wireStruct.Field[fieldnum]
		if wireField.Name == "" {
			errorf("empty name for remote field of type %s", wireStruct.Name)
		}
		ovfl := overflow(wireField.Name)
		localField, present := srt.FieldByName(wireField.Name)

		if !present || !isExported(wireField.Name) {
			op := dec.decIgnoreOpFor(wireField.Id, make(map[typeId]*decOp))
			engine.instr[fieldnum] = decInstr{*op, fieldnum, nil, ovfl}
			continue
		}
		if !dec.compatibleType(localField.Type, wireField.Id, make(map[reflect.Type]typeId)) {
			errorf("wrong type (%s) for received field %s.%s", localField.Type, wireStruct.Name, wireField.Name)
		}
		op := dec.decOpFor(wireField.Id, localField.Type, localField.Name, seen)
		engine.instr[fieldnum] = decInstr{*op, fieldnum, localField.Index, ovfl}
		engine.numInstr++
	}
	return
}

func (dec *Decoder) compileSingle(remoteId typeId, ut *userTypeInfo) (engine *decEngine, err error) {
	rt := ut.user
	engine = new(decEngine)
	engine.instr = make([]decInstr, 1)
	name := rt.String()               

	if !dec.compatibleType(rt, remoteId, make(map[reflect.Type]typeId)) {
		remoteType := dec.typeString(remoteId)
		if ut.base.Kind() == reflect.Interface && remoteId != tInterface {
			return nil, errors.New("gob: local interface type " + name + " can only be decoded from remote interface type; received concrete type " + remoteType)
		}
		return nil, errors.New("gob: decoding into local type " + name + ", received remote type " + remoteType)
	}
	op := dec.decOpFor(remoteId, rt, name, make(map[reflect.Type]*decOp))
	ovfl := errors.New(`value for "` + name + `" out of range`)
	engine.instr[singletonField] = decInstr{*op, singletonField, nil, ovfl}
	engine.numInstr = 1
	return
}

func (dec *Decoder) compileIgnoreSingle(remoteId typeId) *decEngine {
	engine := new(decEngine)
	engine.instr = make([]decInstr, 1)
	op := dec.decIgnoreOpFor(remoteId, make(map[typeId]*decOp))
	ovfl := overflow(dec.typeString(remoteId))
	engine.instr[0] = decInstr{*op, 0, nil, ovfl}
	engine.numInstr = 1
	return engine
}
```
我们分别分析下`decIgnoreOpFor`和`decOpFor`。
* `decIgnoreOpFor`：如果能在表`decIgnoreOpMap`中根据typeId找到对应的跳过操作就可以直接返回，另外如果远程的typeId等于tInterface也就是接口对应的typeId那么就生成一个匿名的解码指令，内部调用`ignoreInterface`。其它情况，根据解码器中wireType缓存获取其wireType结构体，然后判断结构体内部那个字段不为空，如果是slice类型那么就递归调用获取其元素的跳过操作，然后借助`ignoreSlice`来完成跳过操作。
```Go
// 有省略
func (dec *Decoder) decIgnoreOpFor(wireId typeId, inProgress map[typeId]*decOp) *decOp {
	if opPtr := inProgress[wireId]; opPtr != nil {
		return opPtr
	}
	op, ok := decIgnoreOpMap[wireId]
	if !ok {
		inProgress[wireId] = &op
		if wireId == tInterface {
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.ignoreInterface(state)
			}
			return &op
		}

		wire := dec.wireType[wireId]
		switch {
		case wire.SliceT != nil:
			elemId := wire.SliceT.Elem
			elemOp := dec.decIgnoreOpFor(elemId, inProgress)
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.ignoreSlice(state, *elemOp)
			}

		case wire.StructT != nil:
			enginePtr, err := dec.getIgnoreEnginePtr(wireId)
			if err != nil {
				error_(err)
			}
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.ignoreStruct(*enginePtr)
			}

		case wire.GobEncoderT != nil, wire.BinaryMarshalerT != nil, wire.TextMarshalerT != nil:
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.ignoreGobDecoder(state)
			}
		}
	}
	if op == nil {
		errorf("bad data: ignore can't handle type %s", wireId.string())
	}
	return &op
}
```
单独看看`ignoreInterface`和`ignoreSlice`。接口被编码为“动态值的类型描述符 + name_len + name + 类型id + 动态值编码的长度 + 动态值”，类型描述符在解码值前已经被解码过了，详情看[解码流程](#解码流程)，所以我们不用管，在函数中，我们先获取类型名字的长度，然后直接跳过该长度，然后通过`decodeTypeSequence`解码一个类型id，获取完动态值编码的长度后我们直接跳过就可以了。`ignoreSlice`首先解码slice的长度，然后借助`ignoreArrayHelper`连续调用该长度个跳过元素操作。
```GO
func (dec *Decoder) ignoreInterface(state *decoderState) {
	n, ok := state.getLength()
	if !ok {
		errorf("bad interface encoding: name too large for buffer")
	}
	bn := state.b.Len()
	if bn < n {
		errorf("invalid interface value length %d: exceeds input size %d", n, bn)
	}
	state.b.Drop(n)
	id := dec.decodeTypeSequence(true)
	if id < 0 {
		error_(dec.err)
	}

	n, ok = state.getLength()
	if !ok {
		errorf("bad interface encoding: data length too large for buffer")
	}
	state.b.Drop(n)
}

func (dec *Decoder) ignoreSlice(state *decoderState, elemOp decOp) {
	dec.ignoreArrayHelper(state, elemOp, int(state.decodeUint()))
}
func (dec *Decoder) ignoreArrayHelper(state *decoderState, elemOp decOp, length int) {
	instr := &decInstr{elemOp, 0, nil, errors.New("no error")}
	for i := 0; i < length; i++ {
		if state.b.Len() == 0 {
			errorf("decoding array or slice: length exceeds input size (%d elements)", length)
		}
		elemOp(instr, state, noValue)
	}
}
```

* `decOpFor`：同`decIgnoreOpFor`的流程差不多，只是decIgnoreOpMap换成了decOpTable，然后根据用户类型的基本类型（非指针）来构建对应的解码操作。
```Go
// 有省略
func (dec *Decoder) decOpFor(wireId typeId, rt reflect.Type, name string, inProgress map[reflect.Type]*decOp) *decOp {
	ut := userType(rt)
	if ut.externalDec != 0 {
		return dec.gobDecodeOpFor(ut)
	}

	if opPtr := inProgress[rt]; opPtr != nil {
		return opPtr
	}
	typ := ut.base
	var op decOp
	k := typ.Kind()
	if int(k) < len(decOpTable) {
		op = decOpTable[k]
	}
	if op == nil {
		inProgress[rt] = &op
		switch t := typ; t.Kind() {
		case reflect.Slice:
			name = "element of " + name
			if t.Elem().Kind() == reflect.Uint8 {
				op = decUint8Slice
				break
			}
			var elemId typeId
			if tt, ok := builtinIdToType[wireId]; ok {
				elemId = tt.(*sliceType).Elem
			} else {
				elemId = dec.wireType[wireId].SliceT.Elem
			}
			elemOp := dec.decOpFor(elemId, t.Elem(), name, inProgress)
			ovfl := overflow(name)
			helper := decSliceHelper[t.Elem().Kind()]
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.decodeSlice(state, value, *elemOp, ovfl, helper)
			}

		case reflect.Struct:
			ut := userType(typ)
			enginePtr, err := dec.getDecEnginePtr(wireId, ut)
			if err != nil {
				error_(err)
			}
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				dec.decodeStruct(*enginePtr, value)
			}
		case reflect.Interface:
			op = func(i *decInstr, state *decoderState, value reflect.Value) {
				state.dec.decodeInterface(t, state, value)
			}
		}
	}
	if op == nil {
		errorf("decode can't handle type %s", rt)
	}
	return &op
}
```
我们单独分析下结构体的解码操作，对于结构体，首先会获取基本类型的用户类型信息，然后递归调用`getDecEnginePtr`构建其结构体的解码引擎，最后通过`decodeStruct`包装一个解码操作出来。在该函数内，首先会得到一个解码器状态，然后把fieldnum设置为-1，这里与编码器对应，然后遍历解码器缓存，解码出字段号的差值出来，然后恢复其字段号，根据字段号获取对应的解码指令，如果字段索引不为空，就通过反射操作取得其字段，如果该字段是个指针类型，需要调用`decAlloc`来为其分配内存，最后调用解码操作，完成一个字段的解码工作，当解码字段号差值为0时说明结构体解码完成。
```Go
func (dec *Decoder) decodeStruct(engine *decEngine, value reflect.Value) {
	state := dec.newDecoderState(&dec.buf)
	defer dec.freeDecoderState(state)
	state.fieldnum = -1
	for state.b.Len() > 0 {
		delta := int(state.decodeUint())
		if delta < 0 {
			errorf("decode: corrupted data: negative delta")
		}
		if delta == 0 {
			break
		}
		fieldnum := state.fieldnum + delta
		if fieldnum >= len(engine.instr) {
			error_(errRange)
			break
		}
		instr := &engine.instr[fieldnum]
		var field reflect.Value
		if instr.index != nil {
			field = value.FieldByIndex(instr.index)
			if field.Kind() == reflect.Ptr {
				field = decAlloc(field)
			}
		}
		instr.op(instr, state, field)
		state.fieldnum = fieldnum
	}
}

func decAlloc(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
```

### 解码流程
 首先通过`NewDecoder`创建一个新的解码器，然后调用`Decode`方法解码数据，`Decode`实际上调用了`DecodeValue`
 ```Go
 func (dec *Decoder) DecodeValue(v reflect.Value) error {
	if v.IsValid() {
		if v.Kind() == reflect.Ptr && !v.IsNil() {
		} else if !v.CanSet() {
			return errors.New("gob: DecodeValue of unassignable value")
		}
	}
	dec.mutex.Lock()
	defer dec.mutex.Unlock()
	dec.buf.Reset() 
	dec.err = nil
	id := dec.decodeTypeSequence(false)
	if dec.err == nil {
		dec.decodeValue(id, v)
	}
	return dec.err
}
 ```
 该函数首先会调用`decodeTypeSequence`来读取一个消息并获取到typeId，如果typeId小于0，说明是一个wireType还需要调用`recvType`来接收，该函数主要就是解码一个wireType然后通过typeId缓存在解码器的wireType字段中：`dec.wireType[id] = wire`。
 ```Go
 func (dec *Decoder) decodeTypeSequence(isInterface bool) typeId {
	for dec.err == nil {
		if dec.buf.Len() == 0 {
			if !dec.recvMessage() {
				break
			}
		}
		// 先从缓存接收一个整数id
		id := typeId(dec.nextInt())
		if id >= 0 {
			// id 大于0说明接下来不是wireType可以直接返回
			return id
		}

		dec.recvType(-id)
		// 该函数不仅decodeValue调用decodeInterface时也会调用
		// 如果是接口类型那么编码的动态值前面会有个动态值编码的长度，跳过
		if dec.buf.Len() > 0 {
			if !isInterface {
				dec.err = errors.New("extra data in buffer")
				break
			}
			dec.nextUint()
		}
	}
	return -1
}
 ```
 接着就开始解码值了，解码器的`decodeValue`方法会先获取到用户类型信息，然后通过wireType即远程类型信息，以及本地用户信息来创建编码引擎：`getDecEnginePtr`->`compileDec`，然后根据是否是结构体以及是否实现了外部解码器分别调用`decodeStruct`和`decodeSingle`。
 
 * 非结构体：
 ```Go
 func (dec *Decoder) decodeSingle(engine *decEngine, value reflect.Value) {
	state := dec.newDecoderState(&dec.buf)
	defer dec.freeDecoderState(state)
	state.fieldnum = singletonField
	// 解码fieldnum
	if state.decodeUint() != 0 {
		errorf("decode: corrupted data: non-zero delta for singleton")
	}
	instr := &engine.instr[singletonField]
	instr.op(instr, state, value)
}
 ```

 * 结构体：注意这个函数在[上一节](#解码器引擎)构建结构体解码指令时也有调用
 ```Go
 func (dec *Decoder) decodeStruct(engine *decEngine, value reflect.Value) {
	state := dec.newDecoderState(&dec.buf)
	defer dec.freeDecoderState(state)
	state.fieldnum = -1
	for state.b.Len() > 0 {
		delta := int(state.decodeUint())
		if delta < 0 {
			errorf("decode: corrupted data: negative delta")
		}
		if delta == 0 {
			break
		}
		fieldnum := state.fieldnum + delta
		if fieldnum >= len(engine.instr) {
			error_(errRange)
			break
		}
		instr := &engine.instr[fieldnum]
		var field reflect.Value
		if instr.index != nil {
			field = value.FieldByIndex(instr.index)
			if field.Kind() == reflect.Ptr {
				field = decAlloc(field)
			}
		}
		instr.op(instr, state, field)
		state.fieldnum = fieldnum
	}
}
 ```