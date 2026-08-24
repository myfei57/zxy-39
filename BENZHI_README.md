# PipeWatch 油气长输管道 SCADA 监控平台

PipeWatch 是面向油气长输管道泵站与阀室的 SCADA 监控平台，覆盖管线/区段命名空间、
站场注册与控制模式、压力/流量/温度变送器采集、周期扫描与批次、越限报警与确认、
联锁关阀与释放、控制指令下发与回执、历史窗口聚合、读数配额与全链路审计。

## 构建与运行

```bash
go build -mod=vendor ./...
go test -mod=vendor ./...
go vet -mod=vendor ./...
```

启动控制台服务：

```bash
go run -mod=vendor ./cmd/pipewatch -addr :8080 -data ./data
```

服务启动后提供以下主要 API：

- `GET /healthz` 健康检查
- `GET /api/stations` 站场列表
- `POST /api/stations` 注册站场
- `POST /api/stations/{id}/mode` 切换自动/手动模式
- `POST /api/stations/{id}/commands` 下发控制指令
- `GET /api/gauges` 变送器列表
- `POST /api/gauges` 注册变送器
- `POST /api/gauges/{id}/failover` 切换通信通道
- `POST /api/scan/run` 手动触发一次扫描周期
- `GET /api/alarms` 报警列表
- `POST /api/alarms/{id}/confirm` 确认报警
- `GET /api/interlocks` 联锁与洪泛抑制状态
- `POST /api/interlocks/{stationID}/retry` 联锁重试
- `POST /api/interlocks/{stationID}/release` 联锁释放
- `GET /api/history` 历史聚合
- `GET /api/audit` 审计记录
- `GET /api/quota` 配额列表

所有数据以 JSON 文件形式持久化在 `-data` 指定的目录下，不依赖外部数据库。
