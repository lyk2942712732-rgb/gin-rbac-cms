# 🚀 Gin-RBAC-CMS (企业级内容管理系统核心后端)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-v1.9-00ADD8?style=flat)
![Redis](https://img.shields.io/badge/Redis-v7.0-DC382D?style=flat&logo=redis)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## 📖 项目简介
本项目是一个基于 Go (Gin) 编写的高性能、高可用的企业级 CMS 后端接口服务。项目严格遵循 RESTful API 设计规范，内置了完善的 RBAC 动态权限管理系统，并深度集成了 Redis 缓存架构与全自动 CI/CD 部署流水线，已达到真实线上生产环境的交付标准。

## ✨ 核心亮点 (Core Features)
- **🔐 动态权限引擎 (RBAC)：** 基于 5 表关联的底层架构，结合 JWT 无状态鉴权，实现极其细粒度的接口级权限拦截，防范水平/垂直越权。
- **⚡ 高并发缓存架构：** 引入 Redis 应对热点文章的高并发访问，严格落地 **Cache Aside (旁路缓存)** 策略，完美解决高并发下的数据一致性难题。
- **📦 云原生与 DevOps：** 编写 Dockerfile 与 docker-compose.yml 实现环境一键隔离；基于 GitHub Actions 打通自动化构建与基于 Self-Hosted 的私有化服务器平滑部署。
- **📡 极致可观测性：** 弃用原生 log，接入企业级高性能日志库 **Uber Zap** + Lumberjack，实现结构化 JSON 日志记录与自动化文件切割保留。
- **📄 自动化接口文档：** 集成 Swaggo，利用代码注释实时生成并在线预览 API 文档。

## 🛠️ 技术栈
- **Web 框架：** Gin
- **ORM 与 数据库：** GORM, MySQL 8.0
- **缓存引擎：** Redis 7.0
- **安全与密码学：** JWT (JSON Web Token), Bcrypt
- **工程化组件：** Zap (日志), Swaggo (文档), Docker (容器), GitHub Actions (流水线)

## 🚀 极速启动
1. 克隆本项目到本地 / 服务器。
2. 确保本机已安装 Docker 与 Docker Compose。
3. 在项目根目录执行：
   ```bash
   docker compose up -d --build
