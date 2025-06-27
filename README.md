# function call graph

通过 AST，SSA及初始化规则分析代码将代码关系描述出来。

将所有代码分析后归属提取出来Package,File,Struct, Interface, Function，Inject 等可以对象,



通过用户定义的入口，将函数的调用链通过Function中的静态和动态调用，加上Function 所述的Struct 的Ieject(也有可能父级函数的继承) 处理 得到调用链信息。 （只保存需要的函数和调用链路，这样报告内容比较小）



Inject:

通过分析interface 与 struct 签名或者wire 文件得到的调用初始化信息

```go
type Inject struct {
    St   uint32 `json:"struct"`
    Meta map[string][]*Inject
}

```


Struct:

函数的定义及字段的初始化信息
    
```go

type Struct struct {
    Index   uint32 `json:"index"`
    File    uint32 `json:"file"`
    Package uint32 `json:"package"`
    Name    string `json:"name"`

    Inst *types.Named `json:"-"`
    Inject *Inject `json:"inject"`
}
```



Interface :

记录所有的实现

```go
type Interface struct {
    Index   uint32           `json:"index"`
    File    uint32           `json:"file"`
    Package uint32           `json:"package"`
    Name    string           `json:"name"`
    Inst    *types.Interface `json:"-"`


    Impls []uint32 `json:"impls"`
}

```




Function：

函数body 包含的调用信息


```go
type Function struct {
    Index         uint32          `json:"index"`
    File          uint32          `json:"file"`
    Package       uint32          `json:"package"`
    CallStatic    []Call          `json:"static"`
    CallInterface []CallInterface `json:"interface"`
    Name          string          `json:"name"`
    Ready bool `json:"-"`
}

 

type CallType uint8

const (
    CallTypeStatic CallType = iota + 1
    CallTypeInterface
)

type Call struct {
    Line   uint32   `json:"line"`
    Column uint32   `json:"column"`
    Type   CallType `json:"type"`
}

type CallInterfacePath struct {
    // name  idx 互斥，优先用name
    // refactor 这里最好处理下 后续统一名字或者索引
    Name string
    Idx  int
}

```
