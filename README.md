# Multi Device Preview Generator

多设备预览图生成器是一个使用 Go 语言开发的命令行工具，它可以截取网站在不同设备上的显示效果，并将它们组合成一张预览图。

![GitHub](https://raw.githubusercontent.com/vespeng/multi-device-preview/main/assets/preview.png)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Language](https://img.shields.io/badge/language-Go-orange.svg)

## 功能特点

- 📱 支持多种设备类型（MacBook、iPad、iPhone）
- 🖼️ 自动生成设备外壳效果图
- ⚡ 并发截图提高效率
- 🖥️ 跨平台支持（Windows、macOS、Linux）

## 设备支持

目前支持以下设备：

1. MacBook 16 Pro
2. iPad Pro 13
3. iPhone 15 Pro

## 安装要求

- Go 1.25 或更高版本
- 支持的浏览器：
  - Google Chrome
  - Microsoft Edge

## 安装步骤

### 方法一：使用 Go install

```bash
go install github.com/vespeng/multi-device-preview@latest
```

### 方法二：从源码构建

```bash
git clone https://github.com/vespeng/multi-device-preview.git
cd multi-device-preview
go build -o multi-device-preview .
```

## 使用方法

在终端中运行以下命令：

```bash
./multi-device-preview <url>
```

例如：

```bash
./multi-device-preview https://github.com/
```

生成的预览图将保存在与可执行文件相同的目录下，文件名为 `preview.png`

## 工作原理

1. 程序启动时自动检测系统中安装的 Chrome 或 Edge 浏览器
2. 对每种设备类型并发启动无头浏览器实例
3. 在每个浏览器实例中访问指定 URL 并等待页面加载完成
4. 对每个设备进行截图
5. 将截图与对应的设备框架图片合成
6. 将所有设备预览图组合到一张画布上
7. 保存最终的预览图

## 配置说明

设备参数配置在 [config.go](config.go) 文件中，您可以自定义：

- 设备名称
- 设备框架图片路径
- 屏幕区域尺寸
- 屏幕区域偏移量
- 设备在画布上的布局位置
