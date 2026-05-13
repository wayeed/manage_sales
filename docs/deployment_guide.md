# 部署文档

> 家具销售提成管理系统 - 部署与运维指南

## 目录

- [1. 环境要求](#1-环境要求)
- [2. 数据库配置](#2-数据库配置)
- [3. 后端部署](#3-后端部署)
- [4. 管理后台部署](#4-管理后台部署)
- [5. 移动端部署](#5-移动端部署)
- [6. 环境变量说明](#6-环境变量说明)
- [7. 常见问题排查](#7-常见问题排查)

---

## 1. 环境要求

### 1.1 服务器要求

| 项目 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 2 核 | 4 核及以上 |
| 内存 | 4 GB | 8 GB 及以上 |
| 磁盘 | 50 GB SSD | 100 GB SSD |
| 操作系统 | CentOS 7+ / Ubuntu 18.04+ | Ubuntu 22.04 LTS |

### 1.2 软件依赖

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.25+ | 后端编译运行 |
| Node.js | 18+ | 前端构建（推荐 20 LTS） |
| MySQL | 8.0+ | 数据库 |
| Nginx | 1.20+ | 反向代理与静态资源服务 |
| Redis | 7.0+ | 缓存（可选，推荐） |
| Git | 2.30+ | 代码拉取 |

### 1.3 网络要求

| 端口 | 用途 | 说明 |
|------|------|------|
| 80 | HTTP | Nginx 入口 |
| 443 | HTTPS | Nginx SSL 入口 |
| 3306 | MySQL | 数据库端口（仅内网访问） |
| 8080 | 后端服务 | Gin 默认端口（仅内网访问） |

---

## 2. 数据库配置

### 2.1 安装 MySQL

```bash
# Ubuntu
sudo apt update
sudo apt install mysql-server-8.0 -y

# CentOS
sudo yum install mysql-server-8.0 -y
```

### 2.2 配置 MySQL

编辑 MySQL 配置文件 `/etc/mysql/mysql.conf.d/mysqld.cnf`（Ubuntu）或 `/etc/my.cnf`（CentOS）：

```ini
[mysqld]
# 基本配置
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# 连接配置
max_connections = 500
max_allowed_packet = 64M

# InnoDB 配置
innodb_buffer_pool_size = 2G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2

# 时区
default-time-zone = '+08:00'

# 慢查询日志
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

重启 MySQL：

```bash
sudo systemctl restart mysql
sudo systemctl enable mysql
```

### 2.3 创建数据库

```bash
# 登录 MySQL
mysql -u root -p

# 执行以下 SQL
CREATE DATABASE furniture_commission DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 创建专用数据库用户
CREATE USER 'furniture_app'@'%' IDENTIFIED BY 'YourStrongPassword123!';
GRANT ALL PRIVILEGES ON furniture_commission.* TO 'furniture_app'@'%';
FLUSH PRIVILEGES;
```

### 2.4 执行数据库初始化

```bash
# 进入后端项目目录
cd backend/

# 执行表结构初始化脚本
mysql -u furniture_app -p furniture_commission < deploy/init.sql

# 执行初始数据（种子数据）
mysql -u furniture_app -p furniture_commission < deploy/seed.sql
```

> **说明：**
> - `init.sql` 包含 38 张表的结构定义
> - `seed.sql` 包含系统初始数据（管理员账号、角色权限、系统配置等）
> - 默认管理员账号：`admin` / `123456`，首次登录后请立即修改密码

### 2.5 数据库备份

建议配置定时备份：

```bash
# 创建备份脚本
cat > /opt/scripts/backup_mysql.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

mysqldump -u furniture_app -p'YourStrongPassword123!' \
  --single-transaction \
  --routines \
  --triggers \
  furniture_commission | gzip > $BACKUP_DIR/furniture_commission_$DATE.sql.gz

# 保留最近 30 天的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete
EOF

chmod +x /opt/scripts/backup_mysql.sh

# 添加到 crontab，每天凌晨 2 点备份
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/scripts/backup_mysql.sh") | crontab -
```

---

## 3. 后端部署

### 3.1 编译

```bash
# 进入后端项目目录
cd backend/

# 拉取依赖
go mod download

# 编译（Linux）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o furniture-api ./cmd/server

# 编译（如需交叉编译 Windows 版本）
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o furniture-api.exe ./cmd/server
```

### 3.2 目录结构

建议的部署目录结构：

```
/opt/furniture-commission/
├── backend/
│   ├── furniture-api          # 编译后的二进制文件
│   ├── config.yaml            # 配置文件
│   ├── deploy/
│   │   ├── init.sql
│   │   └── seed.sql
│   └── logs/                  # 日志目录
├── admin/                     # 管理后台静态文件
└── mobile/                    # 移动端 H5 静态文件
```

### 3.3 配置文件

创建配置文件 `config.yaml`：

```yaml
# 服务配置
server:
  port: 8080
  mode: release                # debug / release / test
  read_timeout: 30             # 读超时（秒）
  write_timeout: 30            # 写超时（秒）

# 数据库配置
database:
  host: 127.0.0.1
  port: 3306
  user: furniture_app
  password: YourStrongPassword123!
  dbname: furniture_commission
  charset: utf8mb4
  max_idle_conns: 10           # 最大空闲连接数
  max_open_conns: 100          # 最大打开连接数
  max_lifetime: 3600           # 连接最大存活时间（秒）

# Redis 配置（可选）
redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: "your-jwt-secret-key-change-this-in-production"
  expire_hours: 24             # Token 过期时间（小时）
  refresh_expire_hours: 168    # 刷新 Token 过期时间（小时）

# 日志配置
log:
  level: info                  # debug / info / warn / error
  file: logs/app.log
  max_size: 100                # 单个日志文件最大大小（MB）
  max_backups: 10              # 保留的旧日志文件数量
  max_age: 30                  # 保留天数
  compress: true               # 是否压缩

# 文件上传配置
upload:
  max_size: 10                 # 最大文件大小（MB）
  path: ./uploads              # 上传文件存储路径
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
    - application/pdf

# 提成配置
commission:
  order_rate: 0.05             # 订单提成默认比例
  referral_rate: 0.02          # 老带新提成默认比例
  calculate_cron: "0 2 1 * *"  # 每月1日凌晨2点自动计算

# 工资配置
salary:
  base_amount: 5000            # 默认基本工资
  generate_day: 1              # 每月几号生成工资单
```

### 3.4 启动服务

#### 直接启动

```bash
cd /opt/furniture-commission/backend/
./furniture-api -c config.yaml
```

#### 使用 systemd 管理（推荐）

创建 systemd 服务文件：

```bash
sudo cat > /etc/systemd/system/furniture-api.service << 'EOF'
[Unit]
Description=Furniture Commission API Server
After=network.target mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/furniture-commission/backend
ExecStart=/opt/furniture-commission/backend/furniture-api -c config.yaml
Restart=always
RestartSec=5
StandardOutput=append:/opt/furniture-commission/backend/logs/stdout.log
StandardError=append:/opt/furniture-commission/backend/logs/stderr.log

# 安全配置
NoNewPrivileges=true
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl start furniture-api
sudo systemctl enable furniture-api

# 查看状态
sudo systemctl status furniture-api

# 查看日志
sudo journalctl -u furniture-api -f
```

### 3.5 验证服务

```bash
# 检查服务是否启动
curl http://127.0.0.1:8080/api/health

# 预期返回
# {"code":0,"message":"success","data":{"status":"ok","version":"1.0.0"}}
```

---

## 4. 管理后台部署

### 4.1 安装依赖

```bash
# 进入管理后台项目目录
cd admin/

# 安装 Node.js 依赖
npm install
```

### 4.2 构建生产版本

```bash
# 构建生产版本
npm run build
```

构建产物将输出到 `dist/` 目录。

### 4.3 环境配置

在构建前，可通过 `.env.production` 文件配置环境变量：

```bash
# .env.production
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_APP_TITLE=家具销售提成管理系统
```

### 4.4 Nginx 配置

将构建产物部署到 Nginx：

```bash
# 复制构建产物
cp -r dist/* /opt/furniture-commission/admin/
```

创建 Nginx 配置文件：

```bash
sudo cat > /etc/nginx/sites-available/furniture-commission << 'EOF'
# 管理后台
server {
    listen 80;
    server_name admin.yourdomain.com;

    # 管理后台静态文件
    location / {
        root /opt/furniture-commission/admin;
        index index.html;
        try_files $uri $uri/ /index.html;

        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 文件上传大小限制
        client_max_body_size 10M;

        # WebSocket 支持（如需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 上传文件访问
    location /uploads/ {
        alias /opt/furniture-commission/backend/uploads/;
        expires 7d;
    }

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;
    gzip_min_length 1024;
    gzip_vary on;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/furniture-commission /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

### 4.5 HTTPS 配置（推荐）

使用 Let's Encrypt 免费证书：

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx -y

# 获取证书并自动配置
sudo certbot --nginx -d admin.yourdomain.com

# 证书自动续期（certbot 会自动添加 crontab）
sudo certbot renew --dry-run
```

---

## 5. 移动端部署

### 5.1 H5 部署

#### 构建生产版本

```bash
# 进入移动端项目目录
cd mobile/

# 安装依赖
npm install

# 构建 H5 版本
npm run build:h5
```

构建产物将输出到 `dist/build/h5/` 目录。

#### 部署到 Nginx

```bash
# 复制构建产物
cp -r dist/build/h5/* /opt/furniture-commission/mobile/
```

在 Nginx 配置中添加移动端站点：

```bash
sudo cat > /etc/nginx/sites-available/furniture-mobile << 'EOF'
# 移动端 H5
server {
    listen 80;
    server_name m.yourdomain.com;

    location / {
        root /opt/furniture-commission/mobile;
        index index.html;
        try_files $uri $uri/ /index.html;

        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 10M;
    }

    location /uploads/ {
        alias /opt/furniture-commission/backend/uploads/;
        expires 7d;
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;
    gzip_min_length 1024;
    gzip_vary on;
}
EOF

sudo ln -s /etc/nginx/sites-available/furniture-mobile /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 5.2 微信小程序打包

#### 前置准备

1. 注册微信小程序账号：https://mp.weixin.qq.com
2. 在小程序管理后台配置服务器域名
3. 下载并安装 [HBuilderX](https://www.dcloud.io/hbuilderx.html)

#### 打包步骤

```bash
# 构建小程序版本
cd mobile/
npm run build:mp-weixin
```

构建产物将输出到 `dist/build/mp-weixin/` 目录。

#### 上传发布

1. 使用 HBuilderX 打开项目
2. 点击「发行」 -> 「小程序-微信」
3. 填写小程序 AppID
4. 等待编译完成，自动打开微信开发者工具
5. 在微信开发者工具中点击「上传」
6. 登录微信小程序管理后台提交审核

#### 小程序服务器域名配置

在微信小程序管理后台 -> 开发 -> 开发管理 -> 服务器域名中添加：

- request 合法域名：`https://api.yourdomain.com`
- uploadFile 合法域名：`https://api.yourdomain.com`
- downloadFile 合法域名：`https://api.yourdomain.com`

> **注意：** 微信小程序要求所有请求域名必须为 HTTPS，且已通过 ICP 备案。

### 5.3 App 打包（可选）

```bash
# 构建 App 版本
cd mobile/
npm run build:app
```

使用 HBuilderX 进行云端打包或本地打包：

1. HBuilderX 打开项目
2. 点击「发行」 -> 「原生App-云打包」
3. 选择 Android / iOS 平台
4. 填写应用信息，提交打包
5. 下载安装包（.apk / .ipa）

---

## 6. 环境变量说明

### 6.1 后端环境变量

后端支持通过环境变量覆盖配置文件中的值，环境变量优先级高于配置文件。

| 环境变量 | 对应配置项 | 说明 | 示例 |
|----------|-----------|------|------|
| `APP_PORT` | server.port | 服务端口 | `8080` |
| `APP_MODE` | server.mode | 运行模式 | `release` |
| `DB_HOST` | database.host | 数据库主机 | `127.0.0.1` |
| `DB_PORT` | database.port | 数据库端口 | `3306` |
| `DB_USER` | database.user | 数据库用户名 | `furniture_app` |
| `DB_PASSWORD` | database.password | 数据库密码 | `YourPassword` |
| `DB_NAME` | database.dbname | 数据库名称 | `furniture_commission` |
| `REDIS_HOST` | redis.host | Redis 主机 | `127.0.0.1` |
| `REDIS_PORT` | redis.port | Redis 端口 | `6379` |
| `REDIS_PASSWORD` | redis.password | Redis 密码 | `` |
| `JWT_SECRET` | jwt.secret | JWT 密钥 | `your-secret-key` |
| `JWT_EXPIRE_HOURS` | jwt.expire_hours | Token 过期时间 | `24` |
| `LOG_LEVEL` | log.level | 日志级别 | `info` |

### 6.2 管理后台环境变量

| 环境变量 | 说明 | 示例 |
|----------|------|------|
| `VITE_API_BASE_URL` | API 基础地址 | `https://api.yourdomain.com` |
| `VITE_APP_TITLE` | 应用标题 | `家具销售提成管理系统` |

### 6.3 移动端环境变量

| 环境变量 | 说明 | 示例 |
|----------|------|------|
| `VITE_API_BASE_URL` | API 基础地址 | `https://api.yourdomain.com` |
| `VITE_APP_TITLE` | 应用标题 | `家具销售助手` |

---

## 7. 常见问题排查

### 7.1 后端服务启动失败

**问题：** 服务启动报错 `bind: address already in use`

**排查步骤：**

```bash
# 检查端口占用
lsof -i :8080
# 或
netstat -tlnp | grep 8080

# 杀死占用进程
kill -9 <PID>

# 或修改配置文件中的端口
```

**问题：** 数据库连接失败

**排查步骤：**

```bash
# 1. 检查 MySQL 是否运行
sudo systemctl status mysql

# 2. 检查数据库连接
mysql -u furniture_app -p -h 127.0.0.1 furniture_commission

# 3. 检查配置文件中的数据库连接信息
cat config.yaml | grep -A 5 "database:"

# 4. 检查防火墙
sudo ufw status
sudo ufw allow 3306/tcp
```

**问题：** `panic: runtime error: invalid memory address`

**排查步骤：**

```bash
# 查看详细错误日志
tail -100 logs/app.log

# 检查配置文件格式是否正确
# YAML 格式对缩进敏感，确保使用空格而非 Tab
```

### 7.2 前端页面空白

**问题：** 管理后台页面打开后白屏

**排查步骤：**

```bash
# 1. 检查 Nginx 配置
sudo nginx -t

# 2. 检查静态文件是否存在
ls -la /opt/furniture-commission/admin/

# 3. 检查 Nginx 错误日志
sudo tail -50 /var/log/nginx/error.log

# 4. 检查浏览器控制台是否有报错
# F12 -> Console

# 5. 确认 API 地址配置正确
# 检查 .env.production 中的 VITE_API_BASE_URL
```

**问题：** API 请求 404

**排查步骤：**

```bash
# 1. 检查 Nginx 反向代理配置
# 确认 location /api/ 配置正确

# 2. 检查后端服务是否运行
curl http://127.0.0.1:8080/api/health

# 3. 检查 Nginx 访问日志
sudo tail -50 /var/log/nginx/access.log
```

### 7.3 接口返回 401 未授权

**问题：** 登录后请求接口返回 401

**排查步骤：**

```bash
# 1. 检查 Token 是否正确传递
# 浏览器 Network 面板查看请求头是否包含 Authorization

# 2. 检查 Token 是否过期
# Token 默认 24 小时过期，需要重新登录

# 3. 检查 JWT 密钥是否一致
# 确保前后端使用的 JWT_SECRET 一致

# 4. 检查服务器时间是否同步
date
# 如果时间不同步，可能导致 Token 验证失败
sudo ntpdate ntp.aliyun.com
```

### 7.4 文件上传失败

**问题：** 上传图片或文件时报错

**排查步骤：**

```bash
# 1. 检查上传目录是否存在且有写权限
ls -la /opt/furniture-commission/backend/uploads/
mkdir -p /opt/furniture-commission/backend/uploads/
chmod 755 /opt/furniture-commission/backend/uploads/

# 2. 检查 Nginx 文件大小限制
# client_max_body_size 是否足够大

# 3. 检查后端上传大小限制
# config.yaml 中 upload.max_size 配置
```

### 7.5 数据库性能问题

**问题：** 系统响应缓慢

**排查步骤：**

```bash
# 1. 检查 MySQL 慢查询
cat /var/log/mysql/slow.log

# 2. 检查 MySQL 连接数
mysql -u furniture_app -p -e "SHOW PROCESSLIST;"

# 3. 检查数据库表索引
mysql -u furniture_app -p furniture_commission -e "SHOW INDEX FROM orders;"

# 4. 检查数据库大小
mysql -u furniture_app -p -e "
  SELECT table_name,
         ROUND(data_length/1024/1024, 2) AS 'Data (MB)',
         ROUND(index_length/1024/1024, 2) AS 'Index (MB)'
  FROM information_schema.tables
  WHERE table_schema = 'furniture_commission'
  ORDER BY data_length DESC;
"

# 5. 优化建议
# - 增大 innodb_buffer_pool_size
# - 为常用查询字段添加索引
# - 定期清理操作日志表
```

### 7.6 小程序相关

**问题：** 小程序请求接口失败

**排查步骤：**

```
1. 确认域名已配置为 HTTPS
2. 确认域名已在微信小程序后台添加为合法域名
3. 确认域名已通过 ICP 备案
4. 在微信开发者工具中勾选「不校验合法域名」进行调试
5. 检查小程序 AppID 是否正确
```

**问题：** 小程序审核被拒

**排查步骤：**

```
1. 检查是否有未配置的隐私权限
2. 检查用户隐私保护指引是否完善
3. 确认类目选择正确
4. 确认页面功能与描述一致
```

### 7.7 日志查看

```bash
# 后端应用日志
tail -f /opt/furniture-commission/backend/logs/app.log

# Nginx 访问日志
sudo tail -f /var/log/nginx/access.log

# Nginx 错误日志
sudo tail -f /var/log/nginx/error.log

# systemd 服务日志
sudo journalctl -u furniture-api -f --since "1 hour ago"
```

### 7.8 服务更新

```bash
# 1. 拉取最新代码
cd /opt/furniture-commission/backend/
git pull origin main

# 2. 重新编译
go build -o furniture-api ./cmd/server

# 3. 重启服务
sudo systemctl restart furniture-api

# 4. 检查服务状态
sudo systemctl status furniture-api

# 5. 验证服务
curl http://127.0.0.1:8080/api/health
```

### 7.9 健康检查脚本

```bash
#!/bin/bash
# /opt/scripts/health_check.sh

API_URL="http://127.0.0.1:8080/api/health"
ALERT_EMAIL="admin@yourdomain.com"

response=$(curl -s -o /dev/null -w "%{http_code}" $API_URL)

if [ $response -ne 200 ]; then
    echo "[$(date)] API health check failed, HTTP status: $response" >> /opt/furniture-commission/backend/logs/health_check.log
    # 发送告警（可集成邮件、钉钉等通知）
    # echo "Furniture API is down!" | mail -s "API Alert" $ALERT_EMAIL
    systemctl restart furniture-api
fi
```

添加到 crontab，每 5 分钟检查一次：

```bash
(crontab -l 2>/dev/null; echo "*/5 * * * * /opt/scripts/health_check.sh") | crontab -
```

---

> 文档版本：v1.0.0 | 更新日期：2026-05-05
