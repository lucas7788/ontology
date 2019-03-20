# Relayer节点使用介绍

## Relayer节点介绍
为了分散共识节点的压力,并且提高外部节点同步区块的速度,将区块同步服务由专门的节点(Relayer节点)完成。
Relayer节点需要从可信节点同步区块以便向外部节点提供同步区块服务,所以在启动Relayer节点时，要提前配置好上游节点列表，上游节点必须是可信的，因为Relayer节点没有验证merkle根，无法验证区块的合法性。
Relayer节点在同步区块的时候，会将区块按顺序的保存到数据库，在向外提供同步服务时，也会按照区块高度读取数据，并且去掉了区块的反序列化操作，从而提高对外提供同步区块的性能。

## 构建开发环境

成功编译relayer节点需要以下准备：

* Golang版本在1.9及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境
* Golang所支持的操作系统

## 获取ontology relayer节点代码

```git
git clone https://github.com/laizy/ontology/tree/block-relayer
```

用第三方包管理工具glide拉取依赖库

````shell
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide install
````

如果项目有新的第三方依赖包，使用glide更新依赖库

````shell
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide update
````

编译源码
```
go build block-relayer.go
```

成功编译后会生成可以执行程序

* `block-relayer`: 节点程序/以命令行方式提供的节点控制程序

## 运行relayer节点

在运行relayer节点之前，请先准备好`peers.upstream`文件,该文件名可以变化,该文件用于配置同步区块所需要的节点ip地址和端口。
`peers.upstream`文件内容示例:
```
{
"upstream":["23.99.134.190:20338"]
}
```

执行下面的命令启动relayer节点

```
./block-relayer
```

relayer节点会默认读取当前目录下文件名是`peers.upstream`文件,如果你想指定读取的文件,请使用下面的命令启动

```
./block-relayer --upstream-file upstream节点配置文件路径
```

