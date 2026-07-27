# Rclone Studio

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

*English | [中文版](#中文版)*

**Rclone Studio** is a modern, gorgeous, and dynamic Web & Desktop UI for Rclone. Built with Vue 3, TailwindCSS, and Wails, it provides a seamless experience for managing your Rclone instances and cloud storage.

## Features ✨
- **Cross-Platform**: Run as a native Desktop App (via Wails) or access it as a Web App from your browser.
- **Instance Management**: Add, start, stop, and manage multiple local or external Rclone instances.
- **Realtime Dashboard**: Monitor active tasks, memory usage, and upload/download speeds in real-time.
- **Rich File Browser**: Visually browse, preview, upload, and download files from any connected remote.
- **Config Wizard**: A guided setup to add and edit remotes effortlessly.
- **Mount Management**: Easily mount your cloud remotes to local drives directly from the UI (Desktop App only).
- **Localization**: Full English and Simplified Chinese support.

## Development 🚀

### Prerequisites
- Go 1.20+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Run in Dev Mode
```bash
wails dev
```
This will start both the Go backend and the Vite frontend server with hot-reload.

### Build for Production
```bash
wails build
```

---

# 中文版

**Rclone Studio** 是一个为 Rclone 打造的现代、精美且充满动态反馈的桌面端/网页端管理工具。采用 Vue 3, TailwindCSS 和 Wails 构建，为您带来丝滑的云存储管理体验。

## 核心特性 ✨
- **双端支持**：既可以作为原生桌面应用运行，也可以在浏览器中作为网页端管理控制台。
- **多实例管理**：直观地新增、启动、停止和连接本地及远程的多个 Rclone 实例。
- **实时监控仪表盘**：像任务管理器一样实时监控上传/下载速率、内存占用和任务数。
- **可视化文件浏览**：支持文件树浏览、图片/视频预览、上传、下载、新建与重命名。
- **配置向导**：只需几步操作即可快速添加或编辑云存储配置。
- **挂载管理**：一键将云盘挂载为本地磁盘（仅限桌面端）。
- **多语言**：原生支持简体中文与英文。

## 开发指南 🚀

### 环境依赖
- Go 1.20 及以上版本
- Node.js 18 及以上版本
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 运行开发环境
```bash
wails dev
```
此命令将同时启动 Go 后端和 Vite 前端，并开启热更新。

### 编译打包
```bash
wails build
```

## License
MIT License
