package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/disintegration/imaging"
	"golang.org/x/image/draw"
)

var (
	wg sync.WaitGroup
	mu sync.Mutex
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("请提供需要生成预览图的 UR L地址")
		fmt.Println("用法: program_name <url>")
		fmt.Println("示例: program_name http://localhost:8080/")
		os.Exit(1)
	}

	// 从命令行参数获取URL
	url := os.Args[1]

	err := execute(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// execute 执行预览图生成
func execute(url string) error {
	browserPath, err := detectBrowserPath()
	if err != nil {
		return fmt.Errorf("❌ 无法获取浏览器路径: " + err.Error())
	}
	fmt.Println("🔍 使用浏览器:", browserPath)

	// 初始化浏览器分配器上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建一个 map 用于保存每个设备截图
	deviceScreenshots := make(map[string]*image.RGBA)

	// Step 1: 遍历设备并截图
	for _, dev := range Devices {
		wg.Add(1)
		go func(device DeviceParams) {
			defer wg.Done()

			// 为每个协程创建独立的浏览器上下文
			ctx, cancel := chromedp.NewContext(allocCtx)
			defer cancel()

			img, err := takeScreenshotForDevice(ctx, url, device.ScreenW, device.ScreenH, device.Name)
			if err != nil {
				fmt.Printf("❌ 截图失败 (%s): %v\n", device.Name, err)
				return
			}

			mu.Lock()
			deviceScreenshots[device.Name] = img
			mu.Unlock()

			fmt.Println("🖼️ 截图成功 (" + device.Name + ")")
		}(dev)
	}

	wg.Wait()

	// Step 2: 创建透明画布
	canvas := imaging.New(2560, 1600, color.White)

	// Step 3: 所有截图贴入到画布
	fmt.Println("🎨 正在生成预览图...")
	for _, dev := range Devices {
		screenshot := deviceScreenshots[dev.Name]
		resized := imaging.Resize(screenshot, dev.ScreenW, dev.ScreenH, imaging.Lanczos)
		draw.Draw(canvas, image.Rect(dev.LayoutX, dev.LayoutY,
			dev.LayoutX+dev.ScreenW, dev.LayoutY+dev.ScreenH),
			resized, image.Point{}, draw.Over)

		// 读取设备图片
		data, err := deviceFiles.ReadFile(dev.DevicePath)
		if err != nil {
			return fmt.Errorf("❌ 读取设备图片失败 (%s): %v", dev.DevicePath, err)
		}

		// 解码图片数据
		deviceImg, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("❌ 解码设备图片失败 (%s): %v", dev.DevicePath, err)
		}

		// 转换为 RGBA 格式以便绘制
		deviceBounds := deviceImg.Bounds()
		devicePath := image.NewRGBA(deviceBounds)
		draw.Draw(devicePath, deviceBounds, deviceImg, deviceBounds.Min, draw.Src)

		// 将外壳覆盖到画布的对应位置（LayoutX/Y）
		targetRect := image.Rect(
			dev.LayoutX-dev.PointX,
			dev.LayoutY-dev.PointY,
			dev.LayoutX-dev.PointX+deviceBounds.Dx(),
			dev.LayoutY-dev.PointY+deviceBounds.Dy(),
		)

		draw.Draw(canvas, targetRect, devicePath, image.Point{}, draw.Over)
	}

	// Step 4: 保存
	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("❌ 获取可执行文件路径失败: " + err.Error())
	}

	// 获取可执行文件所在目录
	execDir := filepath.Dir(execPath)

	// 构造输出文件路径（与可执行文件同级目录）
	outFile := filepath.Join(execDir, "preview.png")
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("❌ 截图保存失败: " + err.Error())
	}
	defer f.Close()

	if err := png.Encode(f, canvas); err != nil {
		return fmt.Errorf("❌ 截图保存失败: " + err.Error())
	}
	fmt.Println("✅ 预览图生成成功:", outFile)

	return nil
}

// detectBrowserPath 自动探测浏览器路径（支持 Windows Edge / macOS Chrome 等）
func detectBrowserPath() (string, error) {
	var paths []string
	switch runtime.GOOS {
	case "windows":
		paths = []string{
			// edge 32位程序文件夹
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			// edge 64位程序文件夹
			filepath.Join(os.Getenv("PROGRAMFILES"), "Microsoft", "Edge", "Application", "msedge.exe"),
			// Chrome
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Google", "Chrome", "Application", "chrome.exe"),
		}
	case "darwin":
		paths = []string{
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	case "linux":
		paths = []string{
			"/usr/bin/microsoft-edge",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("未找到可用的 Chromium 内核浏览器（Chrome / Edge）请安装后重试。")
}

// takeScreenshotForMacBook 截图
func takeScreenshotForDevice(ctx context.Context, url string, width, height int, deviceName string) (*image.RGBA, error) {
	var buf []byte

	switch deviceName {
	case "MacBook 16 Pro":
		err := chromedp.Run(ctx,
			chromedp.EmulateViewport(int64(width), int64(height)),
			chromedp.Navigate(url),
			chromedp.Sleep(3*time.Second),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.CaptureScreenshot(&buf),
		)
		if err != nil {
			return nil, err
		}
	case "iPad Pro 13":
		err := chromedp.Run(ctx,
			chromedp.Emulate(device.IPadPro),
			chromedp.Navigate(url),
			chromedp.Sleep(3*time.Second),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.CaptureScreenshot(&buf),
		)
		if err != nil {
			return nil, err
		}
	case "iPhone 15 Pro":
		err := chromedp.Run(ctx,
			// todo 这里直接使用 15pro 图像不对，暂时用 12pro
			chromedp.Emulate(device.IPhone12Pro),
			chromedp.Navigate(url),
			chromedp.Sleep(3*time.Second),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.CaptureScreenshot(&buf),
		)
		if err != nil {
			return nil, err
		}
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 如果是 iPhone 15 Pro，应用圆角效果
	if deviceName == "iPhone 15 Pro" {
		rgba = applyCornerTransparency(rgba, 120.0) // 120.0 是圆角半径
	}

	return rgba, nil
}

// applyCornerTransparency 圆角透明
func applyCornerTransparency(src *image.RGBA, cornerRadius float64) *image.RGBA {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	r := cornerRadius

	// 直接操作原图的像素数据
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 判断当前像素是否在某个需要变为透明的圆角内
			if isInCorner(x, y, width, height, r) {
				// 计算该像素在像素数组中的索引位置
				idx := (y-bounds.Min.Y)*src.Stride + (x-bounds.Min.X)*4
				// 将RGBA中的A (Alpha) 通道设置为0 (完全透明)
				src.Pix[idx] = 0   // R
				src.Pix[idx+1] = 0 // G
				src.Pix[idx+2] = 0 // B
				src.Pix[idx+3] = 0 // A
			}
		}
	}

	return src
}

// isInCorner 判断点(x, y)是否位于四个圆角之一的区域内（应被透明化）
func isInCorner(x, y, width, height int, radius float64) bool {
	// 将当前坐标转换为相对于四个角圆心的坐标
	// 左上角圆心: (radius, radius)
	if x < int(radius) && y < int(radius) {
		dx := float64(x) - radius
		dy := float64(y) - radius
		return dx*dx+dy*dy > radius*radius
	}
	// 右上角圆心: (float64(width)-radius, radius)
	if x > width-int(radius)-1 && y < int(radius) {
		dx := float64(x) - (float64(width) - radius)
		dy := float64(y) - radius
		return dx*dx+dy*dy > radius*radius
	}
	// 左下角圆心: (radius, float64(height)-radius)
	if x < int(radius) && y > height-int(radius)-1 {
		dx := float64(x) - radius
		dy := float64(y) - (float64(height) - radius)
		return dx*dx+dy*dy > radius*radius
	}
	// 右下角圆心: (float64(width)-radius, float64(height)-radius)
	if x > width-int(radius)-1 && y > height-int(radius)-1 {
		dx := float64(x) - (float64(width) - radius)
		dy := float64(y) - (float64(height) - radius)
		return dx*dx+dy*dy > radius*radius
	}

	// 不在任何一个角的处理区域内
	return false
}
