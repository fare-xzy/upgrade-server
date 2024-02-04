# Quick Start
### Go-Package

> go get -u golang.org/x/crypto/ssh
>
> go get -u github.com/pkg/sftp
>
> go get -u github.com/gin-gonic/gin（轻量级web框架，github地址：https://github.com/gin-gonic/gin）
>
> go get -u github.com/rs/zerolog
> 
> go get -u gopkg.in/yaml.v3
> 
> go get -u github.com/magiconair/properties
> 
> go env -w GOPROXY=https://goproxy.cn,direct

### 打包

```shell
go build -ldflags="-H=windowsgui -s -w"
```

### 压缩

```shell
upx upgrade-server.exe
```
