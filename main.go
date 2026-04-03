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
	"regexp"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// DaprServiceInfo 存储 Dapr 服务信息
type DaprServiceInfo struct {
	AppID         string   // Dapr AppID
	ServiceName   string   // mDNS 服务名称
	InstanceName  string   // 实例名称
	Port          int      // 端口号
	Host          string   // 主机名
	IPv4Addresses []string // IPv4 地址
	IPv6Addresses []string // IPv6 地址
	TextRecords   []string // 文本记录
}

// DaprMDNSParser Dapr mDNS 解析器
type DaprMDNSParser struct {
	resolver *zeroconf.Resolver
}

// NewDaprMDNSParser 创建新的 Dapr mDNS 解析器
func NewDaprMDNSParser() (*DaprMDNSParser, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}
	return &DaprMDNSParser{resolver: resolver}, nil
}

// ExtractAppIDFromServiceName 从 Dapr 服务名称提取 AppID
func (d *DaprMDNSParser) ExtractAppIDFromServiceName(serviceName string) string {
	// Dapr mDNS 服务格式: {appID}._{protocol}._tcp.local.
	// 或者: {appID}-dapr
	// 提取 AppID

	// 移除 .local. 或 .local 后缀
	service := strings.TrimSuffix(serviceName, ".local.")
	service = strings.TrimSuffix(service, ".local")

	// 如果包含 -dapr 后缀，移除它
	if strings.HasSuffix(service, "-dapr") {
		appID := strings.TrimSuffix(service, "-dapr")
		return appID
	}

	// 如果包含协议后缀 ._tcp 或 ._udp，提取 AppID
	re := regexp.MustCompile(`^([^\.]+)\._(tcp|udp)`)
	matches := re.FindStringSubmatch(service)
	if len(matches) > 1 {
		return matches[1]
	}

	// 直接返回 service（不包含后缀）
	parts := strings.Split(service, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	return serviceName
}

// ParseDaprServiceInfo 解析 Dapr 服务信息
func (d *DaprMDNSParser) ParseDaprServiceInfo(entry *zeroconf.ServiceEntry) *DaprServiceInfo {
	if entry == nil {
		return nil
	}

	info := &DaprServiceInfo{
		ServiceName:  entry.ServiceInstanceName(),
		InstanceName: entry.Instance,
		Port:         entry.Port,
		Host:         entry.HostName,
	}

	// 提取 AppID 和其他信息
	info.AppID = entry.Service

	// 收集所有 IPv4 地址
	for _, addr := range entry.AddrIPv4 {
		info.IPv4Addresses = append(info.IPv4Addresses, addr.String())
	}

	// 收集所有 IPv6 地址
	for _, addr := range entry.AddrIPv6 {
		info.IPv6Addresses = append(info.IPv6Addresses, addr.String())
	}

	// 收集文本记录
	info.TextRecords = entry.Text

	return info
}

// LookupDaprService 查找特定的 Dapr 服务
func (d *DaprMDNSParser) LookupDaprService(ctx context.Context, appID string, timeout time.Duration) []*DaprServiceInfo {
	entries := make(chan *zeroconf.ServiceEntry)
	//defer close(entries)

	//var result *DaprServiceInfo
	resultChan := make(chan *DaprServiceInfo, 1)

	// 在后台处理服务条目
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if entry == nil {
				continue
			}
			if entry.Service == appID {
				info := d.ParseDaprServiceInfo(entry)
				select {
				case resultChan <- info:
				default:
				}
			}
		}
	}(entries)

	// 创建带超时的上下文
	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 浏览 mDNS 服务
	go func() {
		_ = d.resolver.Browse(lookupCtx, appID, "local.", entries)
	}()

	// 等待结果或超时
	//select {
	//case result = <-resultChan:
	//	return result
	//case <-lookupCtx.Done():
	//	return nil
	//}
	var results []*DaprServiceInfo
	for {
		select {
		case info := <-resultChan:
			results = append(results, info) // 收集每一个
		case <-lookupCtx.Done():
			close(resultChan)
			return results // 时间到，返回全部
		}
	}
}

// PrintDaprServiceInfo 打印 Dapr 服务信息
func (d *DaprMDNSParser) PrintDaprServiceInfo(info *DaprServiceInfo) {
	if info == nil {
		fmt.Println("Service not found")
		return
	}

	fmt.Println("========== Dapr 服务信息 ==========")
	fmt.Printf("AppID:                %s\n", info.AppID)
	fmt.Printf("服务名称:              %s\n", info.ServiceName)
	fmt.Printf("实例名称:              %s\n", info.InstanceName)
	fmt.Printf("主机名:                %s\n", info.Host)
	fmt.Printf("Dapr 内部端口:         %d\n", info.Port)
	fmt.Printf("IPv4 地址:            %v\n", info.IPv4Addresses)
	fmt.Printf("IPv6 地址:            %v\n", info.IPv6Addresses)
	fmt.Printf("文本记录:              %v\n", info.TextRecords)
	fmt.Println("==================================")
}

func main() {
	// 定义命令行参数
	appID := flag.String("app-id", "", "指定要查找的 Dapr AppID")
	lookup := flag.Bool("lookup", false, "是否查找指定的 AppID")
	timeout := flag.Duration("timeout", 10*time.Second, "查询超时时间")

	flag.Parse()

	// 创建 Dapr mDNS 解析器
	parser, err := NewDaprMDNSParser()
	if err != nil {
		fmt.Printf("Error creating parser: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 查找特定的 Dapr 服务
	if *lookup {
		if *appID == "" {
			fmt.Println("Error: --lookup 模式需要指定 --app-id")
			return
		}

		fmt.Printf("正在查找 AppID: %s (超时: %v)\n", *appID, *timeout)
		services := parser.LookupDaprService(ctx, *appID, *timeout)

		if services == nil {
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
	flag.PrintDefaults()
}
