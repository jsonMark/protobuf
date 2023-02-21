# protobuf 插件

此仓库后续会编写一系列的类似于grpc-gateway 的一些列插件



- zeroapi ：protoc-gen-zeroapi

  概述：替代goctl生成api端代码，生成目录与文件与goctl保持一致

  为什么要替代goctl生成api：在使用go-zero做微服务的时候，我们要写一遍proto，在写一遍api，然后去分别生成rpc、与api代码，相当于相同的结构体，在proto定义一遍，api也要在定义一遍，还需要手动在api中转换一次，这样可能会降低开发效率。使用此插件，只需要在protobuf中编写一次结构体，使用同一份protobuf同时生成api、rpc的结构体，这样可以大大提升效率。

  

  使用方法：clone代码，进入plugin/protoc-gen-zeroapi ,  go build , 可以看到构建后的插件 protoc-gen-zeroapi，将protoc-gen-zeroapi 移动到 $GOPATH/bin下即可

  【注意】：需要go-zero版本>=1.4.5 

  

  使用案例 ：https://github.com/Mikaelemmmm/protoc-gen-zeroapi-demo

  

  

  



