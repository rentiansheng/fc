# inject 
用来管理服务依赖注入关系，通过依赖注入分析interface 具体实现

主要提供signature及wire 方式

## signature

默认实现，主要判断struct 是否满足interface 约束签名来实现， 该方式噪点比较多

## wire

通过wire方式， 通过wire_gen.go 找到对应的实现
