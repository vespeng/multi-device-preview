package main

import "embed"

//go:embed devices/*.png
var deviceFiles embed.FS

// DeviceParams 定义设备参数
type DeviceParams struct {
	Name       string // 设备名称
	DevicePath string // 设备外壳路径
	ScreenW    int    // 屏幕区域宽度
	ScreenH    int    // 屏幕区域高度
	PointX     int    // 屏幕区域左上角 X 偏移
	PointY     int    // 屏幕区域左上角 Y 偏移
	LayoutX    int    // 布局X
	LayoutY    int    // 布局Y
}

// Devices 设备配置实例
var Devices = []DeviceParams{
	{
		Name:       "MacBook 16 Pro",
		DevicePath: "devices/macbook-pro-16.png",
		ScreenW:    1478,
		ScreenH:    955,
		PointX:     162,
		PointY:     22,
		LayoutX:    640,
		LayoutY:    300,
	},
	{
		Name:       "iPad Pro 13",
		DevicePath: "devices/ipad-pro-13.png",
		ScreenW:    624,
		ScreenH:    830,
		PointX:     28,
		PointY:     28,
		LayoutX:    280,
		LayoutY:    520,
	},
	{
		Name:       "iPhone 15 Pro",
		DevicePath: "devices/iphone-15-pro.png",
		ScreenW:    275,
		ScreenH:    594,
		PointX:     13,
		PointY:     11,
		LayoutX:    720,
		LayoutY:    780,
	},
}
