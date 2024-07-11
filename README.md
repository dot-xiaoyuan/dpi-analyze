# dpi-analyze
deep packet inspection analyze

### 项目结构
```tree
├── cmd
│   └── dpi
│       └── main.go 程序入口
├── config
│   └── config.yaml 配置文件
├── go.mod
├── internal
│   ├── config
│   │   └── config.go 配置加载和管理
│   ├── database  
│   │   ├── mongodb
│   │   │   └── mongodb.go 
│   │   └── redis
│   │       └── redis.go
│   ├── detector
│   │   └── detector.go
│   ├── logger
│   │   └── logger.go
│   ├── parser
│   │   └── parser.go
│   └── storage
├── pkg
│   ├── capture
│   │   └── capture.go
│   ├── protocol
│   └── utils
│       └── utils.go
├── scripts
│   └── build.sh
└── testdata

```