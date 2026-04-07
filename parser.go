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
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

// DaprMDNSParser Dapr mDNS 解析器
type DaprMDNSParser struct {
	resolver *zeroconf.Resolver
}

// NewDaprMDNSParser 创建新的 Dapr mDNS 解析器
func NewDaprMDNSParser() (*DaprMDNSParser, error) {
	// 解析器现在在每次 LookupDaprService 调用中创建
	return &DaprMDNSParser{resolver: nil}, nil
}

// ExtractAppIDFromServiceName 从 Dapr 服务名称提取 AppID
func (d *DaprMDNSParser) ExtractAppIDFromServiceName(serviceName string) string {
	return ExtractAppIDFromServiceName(serviceName)
}

// ExtractAppIDFromServiceName 从 Dapr 服务名称提取 AppID（包级函数）
func ExtractAppIDFromServiceName(serviceName string) string {
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
	// 每次查找创建新的解析器，避免复用导致的问题
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Printf("创建解析器失败: %v", err)
		return nil
	}

	entries := make(chan *zeroconf.ServiceEntry)
	resultChan := make(chan *DaprServiceInfo, 10) // 增加缓冲区大小以容纳多个实例

	// 使用 WaitGroup 等待处理 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(1)

	// 在后台处理服务条目
	go func() {
		defer wg.Done()
		for entry := range entries {
			if entry == nil {
				continue
			}
			log.Printf("发现服务: %s", entry.ServiceInstanceName())
			log.Printf("发现服务: %s", entry.Service)
			log.Printf("发现服务: %s", entry.ServiceName())
			if entry.Service == appID {
				info := d.ParseDaprServiceInfo(entry)
				select {
				case resultChan <- info:
				default:
				}
			}
		}
	}()

	// 创建带超时的上下文
	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 浏览 mDNS 服务
	browseDone := make(chan struct{})
	go func() {
		_ = resolver.Browse(lookupCtx, appID, "local.", entries)
		close(browseDone)
	}()

	var results []*DaprServiceInfo

	// 等待结果或超时
	for {
		select {
		case info := <-resultChan:
			results = append(results, info) // 收集每一个
		case <-lookupCtx.Done():
			// 等待 Browse goroutine 完成
			<-browseDone

			// 等待处理 goroutine 完成
			wg.Wait()
			// 收集剩余的结果
			close(resultChan)
			for info := range resultChan {
				results = append(results, info)
			}
			return results
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
