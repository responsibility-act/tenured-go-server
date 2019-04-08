# 通讯协议生成器

   时间紧张代码真是很乱哦。

## 代码生成器组件介绍

### import 
   
   需要导入的依赖库
   
### enum 定义常量
```
enum 枚举类型名称(枚举类型) {
    枚举名称 = 枚举值
}
```
+ 枚举类型：仅支持基本类型，如果是string可以不写
+ 枚举值，如果是string类型并且名称和值相同不用谢
+ 最终生成的枚举名称：枚举类型名称+枚举名称，此操作是为了防止在两个枚举的名称相同，这样在GO里面就出现问题了

### type 结构体定义

```
//结构体注释
type 结构体名称 {

    //结构体字段注释
	字段 字段类型 empty
	
	...
}
```

+ empty 是否可以为空，
+ 上面两者互斥，不可同时存在

### error 定义错误消息

```
errors {
    错误类型(错误码,错误描述)
    ...
}
```
例如：
```
errors {
    AccountExists(2001,用户已经存在)
    AccountNotExists(2002,用户不存在)
    MobileExists(2003,手机号已被占用)
}
```

###  loadBalance 负载均衡器定义
```
loadBalance {
    负载均衡器名称 负载初始化方法
    ...
}
```
例如：
```
loadBalance {
    zone protocol.NewZoneLoadBalance
    polling protocol.NewPollingLoadBalance
}
```

+ 负载 none，round 的名称是有单独意思，不用定义均衡器，可直接使用

### 接口定义（重点）
定义组成：
```
//接口注释，可多行
service 接口名称(接口开始请求码) [loadBalance(默认负载名称)] [timeout(默认超时设置)] [executor(Fix,10,1000)]{

    //方法注释，可以多长
    方法名称(方法参数 方法参数类型,方法参数n 方法参数类型n) (返回值类型,返回值类型) [error(错误类型1,错误类型2)] [loadBalance(负载方式)] [timeout(超时时间)]
}
```
+ 方法参数可以省略，如果省略参数将会直接使用类型名称作为参数名
+ 返回值只可以是struct,[]byte，[]struct 三种类型，且组合仅为下列四中：
    
       1:   struct   
       2:   []byte 
       3:   struct,[]byte
       4:   []struct,
       其他组合将不受支持

例如：
```
//账户申请接口
service AccountService(2000) loadBalance(zone) timeout(3s) {

    ClearAll() () loadBalance(all) timeout(10s)

    //申请用户
    Apply(Account) () error(AccountExists,MobileExists) loadBalance(polling)

    //根据用户ID获取用户
    Get(id string) (Account) error(AccountNotExists)

    //查询某个状态下的用户
    Query(Search) ([]Account) loadBalance(all)
}

```