# 🚀 Gin-RBAC-CMS 企业级后台管理系统

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-1.9+-008ECF?style=for-the-badge&logo=go)
![GORM](https://img.shields.io/badge/GORM-v1.25-red?style=for-the-badge)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)

本项目是一个基于 Go 语言 (Gin + GORM) 开发的企业级内容管理系统 (CMS) 核心 API。项目主打**高安全性**与**云原生部署**，完整实现了基于角色的权限访问控制 (RBAC) 与无状态鉴权机制，适合作为中小型互联网企业后台业务的底层脚手架。

## ✨ 核心特性 (Features)

- **🔐 严密的权限控制 (RBAC)**：底层采用 5 张表实现角色与权限的精细化关联，通过全局中间件动态拦截越权访问。
- **🛡️ 业务安全加固**：
  - 采用 `Bcrypt` 算法进行密码哈希单向加密，防脱库撞库。
  - 核心业务逻辑（如更新/删除文章）内置**防水平越权 (IDOR)** 校验，确保用户只能操作自身资源。
  - 基于 `JWT` 的无状态身份认证。
- **🐳 云原生友好**：采用 Docker 多阶段构建 (Multi-stage build)，极大压缩镜像体积，并提供 `docker-compose.yml` 实现“代码+环境”一键拉起。
- **📖 优雅的代码工程**：采用标准的分层架构 (Models / Controllers / Middlewares)，代码高内聚低耦合。

## 🛠️ 技术栈 (Tech Stack)

- **Web 框架**: Gin
- **ORM 框架**: GORM
- **数据库**: MySQL 8.0
- **鉴权组件**: golang-jwt/jwt/v5, x/crypto/bcrypt
- **容器化部署**: Docker & Docker Compose

## 📂 目录结构 (Structure)

```text
├── controllers/       # 业务逻辑控制层 (注册/登录/文章增删改查)
├── middlewares/       # 全局中间件 (JWT 校验, RBAC 权限拦截)
├── models/            # 数据库实体模型与 GORM 关联映射
├── Dockerfile         # Docker 多阶段构建脚本
├── docker-compose.yml # 容器编排文件
├── main.go            # 服务入口与路由注册
├── go.mod             # 依赖管理

🚀 极速部署 (Quick Start)
得益于容器化方案，您无需在本地安装 MySQL 和 Go 环境，仅需安装 Docker 即可一键启动整个微服务。

# 1. 克隆项目
git clone [https://github.com/](https://github.com/)lyk2942712732-rgb/gin-rbac-cms.git
cd gin-rbac-cms

# 2. 一键编译并启动 (包含 MySQL 数据库与 Gin 服务)
docker-compose up -d --build

# 3. 检查服务运行状态
docker-compose ps


API 调试说明
建议使用 Apifox 或 Postman 导入并调试以下核心链路：

POST /api/register：注册账号

POST /api/login：登录获取 JWT Token

POST /api/articles：在 Headers 中携带 Token 发布文章
