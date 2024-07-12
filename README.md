# dpi-analyze
deep packet inspection analyze

### 项目结构
```tree
├── cmd
│   └── dpi
│       └── main.go     程序入口
├── config
│   └── config.yaml     配置文件
├── go.mod
├── internal
│   ├── config
│   │   └── config.go       配置加载和管理
│   ├── database  
│   │   ├── mongodb
│   │   │   └── mongodb.go 
│   │   └── redis
│   │       └── redis.go
│   ├── detector
│   │   └── detector.go     流量分析检测
│   ├── logger
│   │   └── logger.go       日志组件
│   ├── parser
│   │   └── parser.go
│   └── storage                 
├── pkg
│   ├── capture
│   │   └── capture.go      数据包捕获
│   ├── protocol
│   └── utils
│       └── utils.go        🔧
├── scripts
│   └── build.sh        打包脚本
└── testdata

```