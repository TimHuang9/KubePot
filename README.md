<img width="168" height="163" alt="logo22222" src="https://github.com/user-attachments/assets/6ae5eb48-b975-4529-96a0-9a8c2bd0a98b" />
KubePot 是一款专为 Kubernetes（K8s）

集群环境设计的蜜罐系统，核心目标是模拟 K8s 集群的关键组件与业务场景，吸引攻击者发起试探与攻击行为，进而收集攻击特征、追溯攻击源头、感知攻击趋势，为 K8s 集群的安全防护提供数据支撑与决策依据。

✨ 核心功能

- 多组件模拟：支持模拟 K8s 核心组件（如 API Server、etcd、kubelet等组件协议）、常用业务容器（如 Nginx、Redis、MySQL）及集群内部网络拓扑，高度还原真实 K8s 集群环境，提升对攻击者的吸引力。

- 全链路攻击捕获：精准捕获针对 K8s 集群的各类攻击行为，包括但不限于：未授权访问、权限提升、镜像篡改、配置泄露、DDoS 攻击、恶意命令执行、敏感路径探测等。

- 攻击数据可视化与分析：自动记录攻击源 IP、攻击时间、攻击方式、利用的漏洞、执行的命令等关键信息，支持生成攻击报告与数据统计图表，帮助安全人员快速掌握攻击态势。

- 实时告警机制：支持配置邮件、Slack、Webhook 等多种告警方式，当检测到高危攻击行为时，可实时推送告警信息，助力安全人员及时响应处置。

- 轻量可扩展：基于容器化部署，资源占用低，可快速集成到现有 K8s 集群或独立部署；支持自定义蜜罐场景与攻击规则，满足不同业务场景下的安全监测需求。

- 安全无逃逸风险：采用模拟协议

📋 适用场景

- K8s 集群安全态势感知与攻击监测

- 云原生环境下的威胁情报收集与分析

- 企业内部 K8s 安全防护演练与攻防测试

- 安全团队对新型 K8s 攻击手段的研究与追踪

🚀 安装部署

kubepot 支持两种部署方式：独立容器部署（适用于快速测试）和 K8s 集群部署（适用于生产环境监测）。

前置依赖

- Docker 19.03+ （独立部署方式）

- Kubernetes 1.18+ （集群部署方式）

- Helm 3.0+ （集群部署方式，可选）

- 网络连通性：确保蜜罐组件可被外部访问（按需开放端口），同时蜜罐可连接告警服务器与数据存储服务。

方式 1：独立容器部署（快速测试）

### 1. 拉取 kubepot 镜像
docker pull your-registry/kubepot:latest  # 替换为实际镜像地址

### 2. 启动蜜罐容器
docker run -d \
  --name kubepot \
  -p 6443:6443  # 模拟 K8s API Server 端口（可根据需求映射其他端口） \
  -v /host/path/logs:/kubepot/logs  # 挂载日志目录到宿主机 \
  -e KUBEPOT_ALERT_WEBHOOK="https://your-alert-webhook"  # 配置告警 Webhook（可选） \
  your-registry/kubepot:latest

启动成功后，可通过访问 http://宿主机IP:6443 触发蜜罐监测，日志将输出到 /host/path/logs 目录。

方式 2：K8s 集群部署（生产环境）

推荐使用 Helm  Charts 部署，支持自定义配置、滚动更新与高可用部署。

### 1. 添加 kubepot Helm 仓库（若未创建，需先自行构建）
helm repo add kubepot https://your-helm-repo/kubepot
helm repo update

### 2. 部署 kubepot 到 K8s 集群（指定命名空间，如 security）
helm install kubepot kubepot/kubepot \
  --namespace security \
  --create-namespace \
  --set service.type=NodePort  # 或 LoadBalancer，根据集群网络配置调整 \
  --set logs.persistence.enabled=true  # 开启日志持久化 \
  --set alert.webhook.enabled=true \
  --set alert.webhook.url="https://your-alert-webhook"

### 3. 查看部署状态
kubectl get pods -n security
kubectl get svc -n security  # 查看蜜罐服务暴露的端口

如需自定义蜜罐模拟组件、监听端口或告警规则，可通过修改 values.yaml 文件实现，具体参考 配置说明 部分。

🔧 使用方法

1. 查看攻击日志

独立部署方式：直接查看宿主机挂载目录下的日志文件（如 /host/path/logs/kubepot.log）。

K8s 集群部署方式：通过 kubectl 查看 Pod 日志或访问持久化日志存储服务（如 ELK 集群）。

### 查看 Pod 实时日志
kubectl logs -f <kubepot-pod-name> -n security

### 查看历史日志（若开启持久化）
kubectl exec -it <kubepot-pod-name> -n security -- cat /kubepot/logs/kubepot-history.log

2. 查看攻击报告

kubepot 支持自动生成 HTML 格式的攻击报告，默认存储在 /kubepot/reports 目录（独立部署）或持久化存储卷中（K8s 部署）。可通过以下方式获取报告：

### 独立部署：直接访问宿主机目录
ls /host/path/logs/reports

### K8s 部署：复制报告到本地
kubectl cp -n security <kubepot-pod-name>:/kubepot/reports ./local-reports

3. 配置告警规则

编辑配置文件 config.yaml（可通过挂载文件或修改 Helm values 配置），设置需要监测的高危攻击行为与告警方式：

# 示例 config.yaml 片段
alert:
  enabled: true
  methods:
    webhook:
      url: "https://your-alert-webhook"
      timeout: 5s
    email:
      smtpServer: "smtp.example.com"
      smtpPort: 465
      username: "alert@example.com"
      password: "your-email-password"
      receivers: ["security@example.com"]
  rules:
    - name: "未授权访问 API Server"
      level: "critical"  # 告警级别：critical/warning/info
      condition: "请求路径为 /api/v1/nodes 且未携带有效 Token"
      notify: true  # 是否触发告警通知

⚙️ 配置说明

kubepot 的核心配置文件为 config.yaml，包含蜜罐模拟配置、日志配置、告警配置、数据存储配置等模块，以下是关键配置项说明：

配置模块

配置项

默认值

说明

honeypot

simulateComponents

["api-server", "etcd"]

需要模拟的 K8s 组件，可选值：api-server、etcd、kubelet、kube-proxy、nginx、redis



listenPorts

{"api-server": 6443, "etcd": 2379}

各模拟组件的监听端口



fakeClusterInfo

true

是否返回伪造的 K8s 集群信息（如节点列表、Pod 列表）

logs

logPath

/kubepot/logs

日志存储路径



logLevel

info

日志级别：debug/info/warn/error

storage

type

local

数据存储类型，可选值：local（本地文件）、mysql、elasticsearch



mysql.dsn

-

MySQL 连接地址（当 storage.type 为 mysql 时必填），格式：user:password@tcp(host:port)/dbname

🏗️ 技术架构

kubepot 采用分层架构设计，整体分为以下几层：

1. 接入层：负责接收外部攻击请求，基于端口与协议分发到对应的蜜罐模拟组件；支持 TCP/UDP/HTTP/HTTPS 等多种协议。

2. 模拟层：核心层，包含多个 K8s 组件与业务场景的模拟模块，模拟真实组件的接口与响应逻辑，同时记录攻击行为细节。

3. 数据处理层：对收集到的攻击数据进行清洗、格式化与分析，生成标准化的日志与攻击报告；支持对接外部数据存储服务。

4. 告警层：基于预设的告警规则，对高危攻击行为进行实时监测与通知，支持多种告警渠道的集成。

5. 配置管理层：提供统一的配置接口，支持动态调整蜜罐模拟规则、告警策略与数据存储方式。

核心技术栈：Go 语言（核心开发）、Docker（容器化部署）、Kubernetes API（组件模拟）、Prometheus/Grafana（可选，数据可视化）。

🤝 贡献指南

欢迎各位开发者参与 kubepot 项目的贡献，无论是功能开发、Bug 修复、文档优化还是需求建议，都非常感谢！贡献流程如下：

1. Fork 本仓库到自己的 GitHub 账号下。

2. 创建新的分支（分支命名规范：feature/xxx、bugfix/xxx、docs/xxx）。

3. 在新分支上进行开发或修改。

4. 提交代码前，确保通过项目的单元测试与代码规范检查。

5. 提交 Pull Request 到本仓库的 main 分支，并详细描述修改内容与动机。

6. 等待项目维护者审核，根据审核意见进行修改，直至合并。

若有重大功能需求或架构调整，建议先在 Issues 中发起讨论，达成共识后再进行开发。

⚠️ 安全提示

- kubepot 仅用于安全监测与研究目的，请勿用于非法攻击或未经授权的测试行为。

- 部署时请确保蜜罐系统与真实业务集群做好网络隔离，避免被攻击者利用蜜罐作为跳板渗透真实环境。

- 定期清理蜜罐日志与攻击数据，避免敏感信息泄露。

📄 许可证

本项目采用 MIT 许可证 开源，详情请查看 LICENSE 文件。

📞 联系我们

若有任何问题、建议或需求，可通过以下方式联系我们：

- GitHub Issues：https://github.com/your-username/kubepot/issues

- 邮箱：your-email@example.com

- 社区：可添加开发者微信（备注：kubepot），加入技术交流群。

<img width="1960" height="828" alt="image" src="https://github.com/user-attachments/assets/e4b3046a-8638-4cd3-af54-0a639e36cbec" />

![alt text](https://github.com/TimHuang9/KubePot/blob/main/docker2375.png?raw=true)

