# Docker 部署文件、数据库和日志排查命令

本文用于排查 New API 使用 Docker 运行时，容器、Docker volume、数据库文件和日志的位置。命令中的 `new-api` 是示例容器名；如果实际名称不同，先用 `docker ps` 查找并替换。

## 先确认容器和 Compose 项目

```bash
# 查看运行中的容器、镜像、状态和端口
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'

# 查看所有容器（包括已停止的容器）
docker ps -a

# 查看 Compose 服务和实际容器名
docker compose ps

# 查看 Compose 最终展开后的配置（包括环境变量来源和 volume 映射）
docker compose config

# 查看容器所属 Compose 项目目录和配置文件
docker inspect new-api --format '{{index .Config.Labels "com.docker.compose.project.working_dir"}} {{index .Config.Labels "com.docker.compose.project.config_files"}}'
```

如果当前目录不是 Compose 文件所在目录，先切换到包含 `docker-compose.yml` 的目录，或者显式指定文件：

```bash
docker compose -f /path/to/docker-compose.yml ps
docker compose -f /path/to/docker-compose.yml config
```

## 查看容器内的目录和文件

```bash
# 查看容器的挂载点、环境变量、启动命令和日志驱动
docker inspect new-api

# 只查看挂载关系
docker inspect new-api --format '{{json .Mounts}}'

# 查看 /data 下的文件；数据库通常位于这里
docker exec new-api sh -lc 'find /data -maxdepth 2 -type f -printf "%p %s bytes\n" | sort'

# 查找常见 SQLite 数据库文件
docker exec new-api sh -lc 'find / -type f \( -name "one-api.db" -o -name "*.db" -o -name "*.db-wal" -o -name "*.db-shm" \) 2>/dev/null'

# 查看应用文件日志；本项目 Compose 通常将 /app/logs 映射到独立 volume
docker exec new-api sh -lc 'find /app/logs -maxdepth 2 -type f -printf "%p %s bytes\n" | sort'

# 进入容器临时排查（退出不会停止容器）
docker exec -it new-api sh
```

SQLite 使用 WAL 模式时，数据库目录中的 `one-api.db-wal` 和 `one-api.db-shm` 也属于数据库状态。备份时不要只复制 `one-api.db`。

## 查看 Docker volume 的宿主机位置

```bash
# 列出所有 volume
docker volume ls

# 查看某个 volume 的实际挂载点
docker volume inspect dmxapi_new_api_data
docker volume inspect dmxapi_new_api_logs

# 只打印 volume 名称和挂载点
docker volume inspect dmxapi_new_api_data \
  --format '{{.Name}} -> {{.Mountpoint}}'

# 查看容器挂载的 volume 名称和容器内路径
docker inspect new-api --format '{{range .Mounts}}{{.Name}}: {{.Source}} -> {{.Destination}}{{"\n"}}{{end}}'
```

在 Linux 上，`Mountpoint` 通常类似 `/var/lib/docker/volumes/<volume>/_data`。在 Docker Desktop、Colima 或 OrbStack 上，这个路径通常位于 Docker 虚拟机内部，不能直接当作 macOS 宿主机路径访问。此时优先通过 `docker exec`、`docker cp` 或 volume 备份方式读取。

## 查看应用日志

本项目有两类常见日志：应用写入 `/app/logs` 的文件日志，以及 Docker `json-file` 驱动收集的标准输出/标准错误日志。

```bash
# 实时查看容器标准输出和标准错误
docker logs -f --tail 200 new-api

# 查看最近一段时间的容器日志
docker logs --since 2h --tail 500 new-api

# 查看 Docker 日志文件路径和日志驱动
docker inspect new-api --format 'driver={{.HostConfig.LogConfig.Type}} path={{.LogPath}}'

# 查看容器内应用文件日志的最后 200 行
docker exec new-api sh -lc 'latest=$(find /app/logs -maxdepth 1 -type f -name "*.log" -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -n 1 | cut -d" " -f2-); test -n "$latest" && tail -n 200 "$latest"'

# 查看文件日志目录总大小
docker exec new-api sh -lc 'du -sh /app/logs 2>/dev/null'
```

不要把完整的环境变量或日志直接提交到工单、代码仓库或聊天记录中；其中可能包含数据库密码、API Key、Cookie 或用户输入。

## 备份数据库和日志 volume

### 备份单个文件到当前目录

适用于能明确确认数据库文件路径的情况：

```bash
mkdir -p ./backup
docker cp new-api:/data/one-api.db ./backup/one-api.db

# 如果存在 WAL 文件，一并复制
docker cp new-api:/data/one-api.db-wal ./backup/one-api.db-wal 2>/dev/null || true
docker cp new-api:/data/one-api.db-shm ./backup/one-api.db-shm 2>/dev/null || true
```

### 备份整个 volume

整个 volume 备份比只复制单个文件更稳妥，也会包含 `/data` 中的其他文件：

```bash
mkdir -p ./backup
docker run --rm \
  -v dmxapi_new_api_data:/source:ro \
  -v "$PWD/backup":/backup \
  alpine:3.20 \
  tar czf /backup/new_api_data-$(date +%Y%m%d-%H%M%S).tar.gz -C /source .

docker run --rm \
  -v dmxapi_new_api_logs:/source:ro \
  -v "$PWD/backup":/backup \
  alpine:3.20 \
  tar czf /backup/new_api_logs-$(date +%Y%m%d-%H%M%S).tar.gz -C /source .
```

备份前最好先停止会写入数据库的应用，尤其是 SQLite。停止 Compose 服务：

```bash
docker compose stop new-api
```

完成备份后再启动：

```bash
docker compose start new-api
```

### 备份 MySQL 或 PostgreSQL

如果 `SQL_DSN` 指向 MySQL 或 PostgreSQL，`/data` 中可能没有 `one-api.db`；数据库数据在数据库容器自己的 volume 中，应使用对应数据库的逻辑备份工具：

```bash
# PostgreSQL 示例
docker exec postgres pg_dump -U root -d new-api > ./backup/new-api.sql

# MySQL 示例（按实际容器名、用户和数据库名替换）
docker exec mysql mysqldump -uroot -p new-api > ./backup/new-api.sql
```

逻辑备份前请确认实际容器名、数据库名和账号，不要把密码写入 shell 历史或文档。

## 判断当前到底使用 SQLite 还是外部数据库

```bash
# 只查看数据库相关配置名和值；执行后注意脱敏，不要把输出公开
docker inspect new-api --format '{{range .Config.Env}}{{println .}}{{end}}' \
  | grep -E '^(SQL_DSN|LOG_SQL_DSN|SQLITE_PATH)='

# 查看应用启动日志中的数据库类型
docker logs new-api 2>&1 | grep -Ei 'using (SQLite|MySQL|PostgreSQL)|SQL_DSN not set'
```

判断规则：

- 未设置 `SQL_DSN`：使用 SQLite，默认数据库通常是 `/data/one-api.db`（具体以启动参数和 `SQLITE_PATH` 为准）。
- `SQL_DSN` 以 `postgres://` 或 `postgresql://` 开头：使用 PostgreSQL。
- 其他普通 SQL DSN：项目按 MySQL 处理，例如 `user:password@tcp(mysql:3306)/new-api`。
- `LOG_SQL_DSN` 只影响独立日志库，不一定等于主数据库。

## 当前 Compose 部署的实测结果

在本仓库当前 Compose 配置中，`new-api` 容器挂载关系是：

```text
dmxapi_new_api_data -> /data
dmxapi_new_api_logs -> /app/logs
```

当前容器的 `SQL_DSN` 指向 PostgreSQL，因此 `/data` 为空并不表示数据丢失；主数据在 PostgreSQL 容器的数据库 volume 中。当前应用文件日志位于 `/app/logs`，Docker 标准输出日志由 `json-file` 驱动保存。

请以你机器上 `docker inspect` 的实时输出为准，不要假设 volume 名称一定是 `dmxapi_*`；Compose 项目名改变后，volume 前缀也会改变。

## 常见误区

- `docker stop` 或删除容器不会自动删除 named volume，但 `docker compose down -v` 会删除 Compose 管理的 volume，执行前必须确认备份已完成。
- 容器内的 `/data` 不是宿主机当前目录，只有配置了 bind mount 时才会直接对应宿主机路径。
- Docker 日志和 `/app/logs` 文件日志是两套日志，排查时要分别查看。
- 如果使用 SQLite，迁移或复制前要考虑 `-wal`、`-shm` 文件；如果使用 PostgreSQL/MySQL，则不应寻找 `one-api.db` 作为主数据源。
