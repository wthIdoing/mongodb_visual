# MongoDB Visual

MongoDB Visual 是一个前后端分离的 MongoDB 可视化管理工具，适合本地 Linux 服务器部署、k8s 部署展示，以及给同学练习 MongoDB 的 database、collection、document、索引和导入导出操作。

本文档同时作为项目介绍和部署说明使用。前半部分介绍项目能力和架构，后半部分按“开发交付给运维”的方式说明配置文件、构建方式和部署方式。

## 1. 项目组成

```text
.
├── backend/           # Go + Gin 后端服务
├── frontend/          # Vue 3 + Vite 前端应用
└── docs/              # 项目说明文档
```

技术栈：

- 前端：Vue 3、Vite、TypeScript、Element Plus
- 后端：Go、Gin、MongoDB Go Driver
- 数据库：MongoDB

核心能力：

- 连接任意 MongoDB 实例
- 支持无认证 MongoDB 和用户名/密码认证 MongoDB
- 浏览 database / collection / document
- 创建和删除 database
- 创建和删除 collection
- document 新增、编辑、删除、批量删除、导入导出
- collection 备份和恢复
- 索引查看、创建、删除
- 条件查询和 JSON 查询
- 中英文切换

## 2. 环境准备

部署前先准备以下环境。

基础环境：

- Linux 服务器
- Git 或代码上传方式
- 可用的 shell 环境

构建后端需要：

- Go 1.22+

构建前端需要：

- Node.js 18+
- npm

运行服务需要：

- MongoDB 8+，也可以连接外部 MongoDB 或 k8s 内 MongoDB Service
- Nginx，推荐用于托管前端 `frontend/dist` 并监听 `80` 端口

端口规划：

- 前端生产访问端口：`80` 或 `443`
- 前端开发端口：`5173`
- 后端 API 端口：`8080`
- MongoDB 默认端口：`27017`

如果部署在云服务器或虚拟机上，需要确认防火墙或安全组已经放行实际访问端口，例如 `80`、`8080`、`27017`。

## 3. 部署架构

推荐部署方式：

```text
Browser
  |
  | 访问 80/443
  v
Nginx / Ingress
  |
  | 托管 frontend/dist
  | 反向代理 /api 到后端
  v
backend:8080
  |
  | 连接 MongoDB
  v
MongoDB:27017
```

也可以让前端直接请求后端：

```text
Browser -> frontend:80
Browser -> backend:8080
backend -> MongoDB:27017
```

两种方式都支持。生产和 k8s 场景更推荐第一种：前端通过同域 `/api` 访问后端，Nginx 或 Ingress 负责反向代理。

## 4. 后端配置

后端配置文件建议放在：

```text
backend/.env
```

可以从示例文件复制：

```bash
cd backend
cp .env.example .env
```

示例配置：

```env
MONGODB_URI=mongodb://127.0.0.1:27017/admin
MONGODB_DATABASE=admin
SERVER_PORT=8080
FRONTEND_ORIGINS=http://127.0.0.1,http://localhost,http://127.0.0.1:5173,http://localhost:5173
```

配置说明：

- `MONGODB_URI`：后端默认 MongoDB 连接。页面连接页传入的会话连接会优先使用。
- `MONGODB_DATABASE`：默认 database，通常可填 `admin`。
- `SERVER_PORT`：后端监听端口，默认 `8080`。
- `FRONTEND_ORIGINS`：允许访问后端的前端 Origin，多个值用英文逗号分隔。

注意：

- 如果前端通过 `http://服务器IP` 访问，`FRONTEND_ORIGINS` 要包含 `http://服务器IP`。
- 如果前端通过 `http://服务器IP:5173` 访问，`FRONTEND_ORIGINS` 要包含 `http://服务器IP:5173`。
- 这里不是“后端连接前端”，而是后端允许这些前端页面跨域请求 API。

## 5. 前端配置

前端配置文件需要自己创建，建议放在：

```text
frontend/.env.production
```

核心变量：

```env
VITE_API_BASE_URL=
```

配置说明：

- `VITE_API_BASE_URL` 是前端构建时变量。
- 修改该变量后，必须重新执行 `npm run build`。
- 如果该变量为空，前端会使用相对路径请求后端，例如 `/api/...`。
- 如果该变量不为空，前端会请求指定的后端地址。

### 推荐配置：Nginx 反向代理 `/api`

如果 Nginx 同时托管前端并代理后端，前端配置可以留空：

```env
VITE_API_BASE_URL=
```

这种方式下，浏览器请求：

```text
http://服务器IP/api/...
```

然后由 Nginx 转发到：

```text
http://127.0.0.1:8080/api/...
```

### 直接请求后端

如果前端不通过 Nginx 代理，而是直接访问后端 `8080`：

```env
VITE_API_BASE_URL=http://服务器IP:8080
```

此时后端 `FRONTEND_ORIGINS` 必须包含前端页面地址。

### k8s 中的前端配置

k8s 中可以把 `VITE_API_BASE_URL` 配成后端 Ingress 地址：

```env
VITE_API_BASE_URL=http://mongodb-visual-api.example.com
```

如果前端容器内的 Nginx 代理 `/api` 到后端 Service，则可以继续留空：

```env
VITE_API_BASE_URL=
```

不建议让浏览器直接访问集群内 Service DNS，例如：

```text
mongodb-visual-backend.default.svc.cluster.local
```

原因是浏览器通常运行在集群外，不能解析 k8s 的 CoreDNS。Service DNS 更适合写在 Nginx、Ingress 或后端服务内部配置里。

## 6. 构建后端

开发启动：

```bash
cd backend
go mod tidy
go run ./cmd/server
```

生产构建：

```bash
cd backend
go mod tidy
go build -o mongodb-visual-server ./cmd/server
```

运行二进制：

```bash
cd backend
./mongodb-visual-server
```

后端启动后默认监听：

```text
http://0.0.0.0:8080
```

健康检查：

```bash
curl http://127.0.0.1:8080/healthz
```

OpenAPI JSON：

```text
http://127.0.0.1:8080/api/openapi.json
```

## 7. 构建前端

安装依赖：

```bash
cd frontend
npm install
```

创建生产配置：

```bash
printf 'VITE_API_BASE_URL=\n' > .env.production
```

`VITE_API_BASE_URL` 表示前端在浏览器中请求后端 API 时使用的基础地址。

例如：

```env
# 前端通过 Nginx 同域反向代理 /api 到后端，推荐生产部署使用
VITE_API_BASE_URL=

# 前端直接请求本机后端
VITE_API_BASE_URL=http://127.0.0.1:8080

# 前端直接请求另一台服务器上的后端
VITE_API_BASE_URL=http://192.168.1.20:8080
```

如果使用空值，前端会请求相对路径，例如 `/api/v1/databases`。如果填写完整地址，前端会请求类似 `http://192.168.1.20:8080/api/v1/databases`。

注意：`VITE_API_BASE_URL` 是构建时变量。修改 `.env.production` 后必须重新执行 `npm run build`。

构建：

```bash
npm run build
```

构建产物：

```text
frontend/dist
```

开发模式启动：

```bash
npm run dev -- --host 0.0.0.0
```

开发模式默认端口是：

```text
5173
```

注意：`npm run dev` 只适合开发调试。生产部署或演示部署建议使用 Nginx 托管 `frontend/dist`。

## 8. Linux 服务器部署示例

假设：

- 项目路径：`/opt/mongodb-visual`
- 前端通过 Nginx 监听 `80`
- 后端监听 `8080`
- MongoDB 监听 `27017`

### 8.1 后端

```bash
cd /opt/mongodb-visual/backend
cp .env.example .env
```

编辑 `backend/.env`：

```env
MONGODB_URI=mongodb://127.0.0.1:27017/admin
MONGODB_DATABASE=admin
SERVER_PORT=8080
FRONTEND_ORIGINS=http://服务器IP,http://localhost,http://127.0.0.1
```

构建并启动：

```bash
go mod tidy
go build -o mongodb-visual-server ./cmd/server
./mongodb-visual-server
```

### 8.2 前端

```bash
cd /opt/mongodb-visual/frontend
printf 'VITE_API_BASE_URL=\n' > .env.production
npm install
npm run build
```

### 8.3 Nginx

安装 Nginx 后，新增站点配置：

```nginx
server {
    listen 80;
    server_name _;

    root /opt/mongodb-visual/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Mongo-Host $http_x_mongo_host;
        proxy_set_header X-Mongo-Port $http_x_mongo_port;
        proxy_set_header X-Mongo-Database $http_x_mongo_database;
        proxy_set_header X-Mongo-Username $http_x_mongo_username;
        proxy_set_header X-Mongo-Password $http_x_mongo_password;
        proxy_set_header X-Mongo-AuthSource $http_x_mongo_authsource;
    }

    location /healthz {
        proxy_pass http://127.0.0.1:8080/healthz;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

检查并重载 Nginx：

```bash
nginx -t
systemctl reload nginx
```

访问：

```text
http://服务器IP
http://服务器IP/connect
```

## 9. k8s 部署配置要点

k8s 部署时仍然保持前后端分离：

- 前端镜像：构建 Vue 静态文件，用 Nginx 托管。
- 后端镜像：运行 Go 服务，监听 `8080`。
- MongoDB：可以是集群内 Service，也可以是外部 MongoDB。

### 9.1 前端如何连接后端 Service

推荐两种方式。

方式一：Ingress 暴露后端 API。

```env
VITE_API_BASE_URL=http://mongodb-visual-api.example.com
```

方式二：前端 Nginx 代理 `/api` 到后端 Service。

```env
VITE_API_BASE_URL=
```

Nginx 中代理：

```nginx
location /api/ {
    proxy_pass http://mongodb-visual-backend.default.svc.cluster.local:8080/api/;
    proxy_set_header X-Mongo-Host $http_x_mongo_host;
    proxy_set_header X-Mongo-Port $http_x_mongo_port;
    proxy_set_header X-Mongo-Database $http_x_mongo_database;
    proxy_set_header X-Mongo-Username $http_x_mongo_username;
    proxy_set_header X-Mongo-Password $http_x_mongo_password;
    proxy_set_header X-Mongo-AuthSource $http_x_mongo_authsource;
}
```

推荐方式二。这样浏览器只访问前端域名，后端 Service DNS 由集群内 Nginx 解析。

### 9.2 后端环境变量

后端 Deployment 中至少配置：

```env
MONGODB_URI=mongodb://mongodb.default.svc.cluster.local:27017/admin
MONGODB_DATABASE=admin
SERVER_PORT=8080
FRONTEND_ORIGINS=http://前端域名
```

如果 MongoDB 开启认证：

```env
MONGODB_URI=mongodb://用户名:密码@mongodb.default.svc.cluster.local:27017/admin?authSource=admin
```

注意：页面连接页也支持输入 MongoDB host、port、username、password、authSource。后端环境变量只是默认兜底连接。

## 10. MongoDB 连接说明

连接页默认值：

```text
host: 127.0.0.1
port: 27017
authSource: admin
username: 空
password: 空
```

MongoDB 不像 MySQL 那样稳定自带 `root@localhost` 用户。`admin` 是认证数据库常用值，不等于默认存在 `admin` 用户。

无认证 MongoDB：

- username 留空
- password 留空
- authSource 保持 `admin`

有认证 MongoDB：

- username 填已创建的 MongoDB 用户
- password 填对应密码
- authSource 填创建用户时所在的认证 database，常见是 `admin`

k8s 中连接 MongoDB Service 时，连接页的 `host` 可以填写：

```text
mongodb.default.svc.cluster.local
```

前提是后端运行环境可以解析该地址。因为实际连接 MongoDB 的是后端，不是浏览器。

## 11. 常见问题

### 前端为什么还是 5173？

`5173` 是 Vite 开发服务器端口，只在 `npm run dev` 时使用。

如果要让前端开放 `80`，请执行：

```bash
cd frontend
npm run build
```

然后用 Nginx 监听 `80` 并托管 `frontend/dist`。

### 修改 `VITE_API_BASE_URL` 后为什么不生效？

`VITE_API_BASE_URL` 是构建时变量。修改后必须重新构建：

```bash
cd frontend
npm run build
```

### 后端的 `FRONTEND_ORIGINS` 应该怎么写？

写浏览器实际访问前端时的 Origin。

示例：

```env
FRONTEND_ORIGINS=http://192.168.1.20,http://localhost,http://127.0.0.1
```

不要写成后端地址，也不要理解成“后端连接前端”。

### 页面连接 MongoDB 失败怎么办？

按顺序检查：

1. 后端是否能访问 MongoDB 的 host 和 port
2. MongoDB 是否开启认证
3. username / password 是否正确
4. authSource 是否正确
5. 当前 MongoDB 用户是否有查看、创建、删除 database / collection 的权限

### 通过页面连接其他主机 MongoDB 时为什么失败？

页面连接页中填写的 MongoDB host 不是由浏览器直接连接，而是由后端服务连接。

例如浏览器访问：

```text
http://10.0.0.10
```

连接页填写：

```text
host: 10.0.0.11
port: 27017
```

实际链路是：

```text
Browser -> Nginx(10.0.0.10:80) -> backend(10.0.0.10:8080) -> MongoDB(10.0.0.11:27017)
```

因此需要确认：

1. `10.0.0.10` 后端能访问 `10.0.0.11:27017`
2. `10.0.0.11` 的 MongoDB 没有只绑定 `127.0.0.1`
3. `10.0.0.11` 防火墙允许 `10.0.0.10` 访问 `27017`
4. Nginx 代理 `/api` 时透传了 `X-Mongo-*` 请求头

在后端所在机器上测试 MongoDB 端口：

```bash
nc -vz 10.0.0.11 27017
```

如果 MongoDB 只监听本机，需要修改 MongoDB 配置中的 `bindIp`，例如：

```yaml
net:
  bindIp: 0.0.0.0
  port: 27017
```

然后重启 MongoDB。

如果 Nginx 没有透传 `X-Mongo-*` 请求头，后端会退回默认连接，或者拿不到页面上填写的目标 MongoDB 地址。`/api` 代理中应包含：

```nginx
proxy_set_header X-Mongo-Host $http_x_mongo_host;
proxy_set_header X-Mongo-Port $http_x_mongo_port;
proxy_set_header X-Mongo-Database $http_x_mongo_database;
proxy_set_header X-Mongo-Username $http_x_mongo_username;
proxy_set_header X-Mongo-Password $http_x_mongo_password;
proxy_set_header X-Mongo-AuthSource $http_x_mongo_authsource;
```

### 空 MongoDB 实例没有 database 怎么办？

页面支持创建 `database + first collection`。MongoDB 不适合创建真正的空 database，因此必须同时指定首个 collection。
