# go-utils

常用 Go 工具库集合，包含生产级的泛型缓存、时间计算等实用组件。

## 工具列表

| 工具 | 作用 | 最低 Go 版本 |
|------|------|:---:|
| [double-cache](./double-cache) | 泛型双缓存库，基于 two-cache 模式，后台定时刷新，无锁读取 | 1.21 |
| [time-anchor](./time-anchor) | 基于时间锚点的周期刻度计算工具 | 1.18 |
| [ticker-mission](./ticker-mission) | 基于 Redis 分布式锁的周期任务调度器，支持幂等启动 | 1.21 |

## 设计理念

- **独立模块**：各工具无内部依赖，可单独引用
- **生产可用**：强调线程安全与边界条件处理
- **接口简洁**：最小化 API surface，降低接入成本

## 工作区结构

```
go-utils/
├── go.work              # 工作区文件
├── double-cache/        # 泛型双缓存
├── time-anchor/         # 时间锚点计算
├── ticker-mission/      # 周期任务调度器
└── README.md
```

## 安装

各工具独立安装：

```bash
# double-cache
go get github.com/CppToGo/go-utils/double-cache

# time-anchor
go get github.com/CppToGo/go-utils/time-anchor

# ticker-mission
go get github.com/CppToGo/go-utils/ticker-mission
```

## 开发

使用 Go workspace 模式管理多模块：

```bash
go work sync        # 同步工作区依赖
go build ./...      # 构建所有模块
go test ./...       # 测试所有模块
```
