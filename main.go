// Copyright 2026 pyschemDa <guangdashao1990@163.com>
//
// Licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"
)

//// ParseDaprServiceInfo 解析 Dapr 服务信息
//func (d *DaprMDNSParser) ParseDaprServiceInfo(entry *zeroconf.ServiceEntry) *DaprServiceInfo {
//	if entry == nil {
//		return nil
//	}
//
//	info := &DaprServiceInfo{
//		ServiceName:  entry.ServiceInstanceName(),
//		InstanceName: entry.Instance,
//		Port:         entry.Port,
//		Host:         entry.HostName,
//	}
//
//	// 提取 AppID 和其他信息
//	info.AppID = entry.Service
//
//	// 收集所有 IPv4 地址
//	for _, addr := range entry.AddrIPv4 {
//		info.IPv4Addresses = append(info.IPv4Addresses, addr.String())
//	}
//
//	// 收集所有 IPv6 地址
//	for _, addr := range entry.AddrIPv6 {
//		info.IPv6Addresses = append(info.IPv6Addresses, addr.String())
//	}
//
//	// 收集文本记录
//	info.TextRecords = entry.Text
//
//	return info
//}
//
//
//
//// PrintDaprServiceInfo 打印 Dapr 服务信息
//func (d *DaprMDNSParser) PrintDaprServiceInfo(info *DaprServiceInfo) {
//	if info == nil {
//		fmt.Println("Service not found")
//		return
//	}
//
//	fmt.Println("========== Dapr 服务信息 ==========")
//	fmt.Printf("AppID:                %s\n", info.AppID)
//	fmt.Printf("服务名称:              %s\n", info.ServiceName)
//	fmt.Printf("实例名称:              %s\n", info.InstanceName)
//	fmt.Printf("主机名:                %s\n", info.Host)
//	fmt.Printf("Dapr 内部端口:         %d\n", info.Port)
//	fmt.Printf("IPv4 地址:            %v\n", info.IPv4Addresses)
//	fmt.Printf("IPv6 地址:            %v\n", info.IPv6Addresses)
//	fmt.Printf("文本记录:              %v\n", info.TextRecords)
//	fmt.Println("==================================")
//}

func runServerMode(appIDPrefix string, webPort int, refreshInterval time.Duration) {
	// 创建服务缓存
	cache := NewServiceCache()

	// 创建 Dapr mDNS 解析器
	parser, err := NewDaprMDNSParser()
	if err != nil {
		log.Fatalf("创建解析器失败: %v", err)
	}

	// 创建 mDNS 监听器 , 监听是否有符合 appIDPrefix 的服务查询，
	listener, err := NewMDNSListener(cache, parser, appIDPrefix)
	if err != nil {
		log.Fatalf("创建 mDNS 监听器失败: %v", err)
	}
	defer listener.Stop()

	// 创建刷新器
	refresher := NewRefresher(cache, parser, refreshInterval)
	defer refresher.Stop()

	// 创建 Web 服务器
	webServer := NewWebServer(cache, webPort)

	// 启动刷新器
	go refresher.Start()

	// 启动 Web 服务器（阻塞）
	log.Printf("启动 Dapr mDNS 服务发现服务器")
	log.Printf("Web UI 地址: http://localhost:%d", webPort)
	log.Printf("服务刷新间隔: %v", refreshInterval)
	if err := webServer.Start(); err != nil {
		log.Fatalf("Web 服务器启动失败: %v", err)
	}
}

func main() {
	// 定义命令行参数
	appID := flag.String("app-id", "", "指定要查找的 Dapr AppID")
	appIDPrefix := flag.String("app-id-prefix", "", "指定要查找的 Dapr AppID 前缀")
	timeout := flag.Duration("timeout", 2*time.Second, "查询超时时间")
	server := flag.Bool("server", false, "启动完整服务模式（监听器 + 刷新器 + Web UI）")
	webPort := flag.Int("web-port", 8080, "Web 服务器端口（仅在 --server 模式下有效）")
	refreshInterval := flag.Duration("refresh-interval", 30*time.Second, "服务刷新间隔（仅在 --server 模式下有效）")

	flag.Parse()

	// 如果启用了 server 模式，启动完整服务
	if *server {
		runServerMode(*appIDPrefix, *webPort, *refreshInterval)
		return
	}

	// 查找特定的 Dapr 服务（保持向后兼容）
	if *appID != "" {

		// 创建 Dapr mDNS 解析器
		parser, err := NewDaprMDNSParser()
		if err != nil {
			fmt.Printf("Error creating parser: %v\n", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		fmt.Printf("正在查找 AppID: %s (超时: %v)\n", *appID, *timeout)
		services := parser.LookupDaprService(ctx, *appID, *timeout)

		if len(services) == 0 {
			fmt.Printf("❌ 未找到 AppID 为 '%s' 的服务\n", *appID)
			return
		}

		fmt.Println("✅ 找到服务")
		fmt.Printf("✅ 共找到 %d 个服务\n", len(services))
		for i, s := range services {
			fmt.Printf("\n===== 服务 %d =====\n", i+1)
			parser.PrintDaprServiceInfo(s)
		}
		fmt.Println("✅ 结束服务")
		return
	}

	// 如果没有指定任何模式，显示帮助
	flag.PrintDefaults()
}
