// Copyright 2026 pyschemDa <guangdashao1990@163.com>
//
// Licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package server

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"dapr-mdns-resolver/internal/cache"
	"dapr-mdns-resolver/internal/parser"
	"dapr-mdns-resolver/pkg/types"
)

// MDNSListener mDNS 监听器
type MDNSListener struct {
	parser *parser.DaprMDNSParser
	cache  *cache.ServiceCache
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMDNSListener 创建新的 mDNS 监听器
func NewMDNSListener(cache *cache.ServiceCache, parser *parser.DaprMDNSParser, appIDPrefix string) (*MDNSListener, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 启动 mDNS 查询捕获 goroutine
	go captureMDNSQueries(cache, parser, appIDPrefix, ctx)

	return &MDNSListener{
		parser: parser,
		cache:  cache,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Stop 停止监听器
func (l *MDNSListener) Stop() {
	l.cancel()
}

// Refresher 定时刷新器，定期查询缓存中的服务以获取最新信息
type Refresher struct {
	cache    *cache.ServiceCache
	parser   *parser.DaprMDNSParser
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRefresher 创建新的刷新器
func NewRefresher(cache *cache.ServiceCache, parser *parser.DaprMDNSParser, interval time.Duration) *Refresher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Refresher{
		cache:    cache,
		parser:   parser,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start 启动定时刷新任务
func (r *Refresher) Start() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.refresh()
		case <-r.ctx.Done():
			return
		}
	}
}

// refresh 刷新缓存中所有服务的详细信息
func (r *Refresher) refresh() {
	appIDs := r.cache.GetAppIDs()
	log.Printf("开始刷新 %d 个服务的详细信息", len(appIDs))
	if len(appIDs) == 0 {
		return
	}

	for _, appID := range appIDs {
		services := r.parser.LookupDaprService(r.ctx, appID, 1*time.Second)
		for i, s := range services {
			log.Printf("\n===== 服务 %d =====\n", i+1)
			r.parser.PrintDaprServiceInfo(s)
		}

		log.Println("✅ 结束服务")
		if len(services) > 0 {
			r.cache.AddOrUpdateAll(appID, services)
			log.Printf("刷新服务: AppID=%s, 找到 %d 个实例", appID, len(services))
		} else {
			log.Printf("服务未找到或已离线: AppID=%s", appID)
		}
	}
}

// Stop 停止刷新器
func (r *Refresher) Stop() {
	r.cancel()
}

// captureMDNSQueries 捕获 mDNS 查询流量，解析查询中的服务名称并存储到缓存
func captureMDNSQueries(cache *cache.ServiceCache, parser *parser.DaprMDNSParser, appIDPrefix string, ctx context.Context) {
	// 获取多播接口列表
	interfaces := listMulticastInterfaces()
	if len(interfaces) == 0 {
		log.Printf("没有可用的多播接口")
		return
	}

	// 启动 IPv4 和 IPv6 监听器
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		captureMDNSQueriesByProtocol("IPv4", cache, parser, appIDPrefix, ctx, interfaces)
	}()

	go func() {
		defer wg.Done()
		captureMDNSQueriesByProtocol("IPv6", cache, parser, appIDPrefix, ctx, interfaces)
	}()

	wg.Wait()
}

// listMulticastInterfaces 返回支持多播的接口列表
func listMulticastInterfaces() []net.Interface {
	var result []net.Interface
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, ifi := range ifaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}

		if (ifi.Flags & net.FlagMulticast) > 0 {
			result = append(result, ifi)
		}
	}

	return result
}

// captureMDNSQueriesByProtocol 捕获指定协议的 mDNS 查询
func captureMDNSQueriesByProtocol(protocol string, cache *cache.ServiceCache, parser *parser.DaprMDNSParser, appIDPrefix string, ctx context.Context, interfaces []net.Interface) {
	var addr *net.UDPAddr
	var network string
	var mdnsGroup net.IP

	if protocol == "IPv4" {
		addr = &net.UDPAddr{
			IP:   net.ParseIP("224.0.0.0"),
			Port: 5353,
		}
		network = "udp4"
		mdnsGroup = net.IPv4(224, 0, 0, 251)
	} else {
		addr = &net.UDPAddr{
			IP:   net.ParseIP("ff02::"),
			Port: 5353,
		}
		network = "udp6"
		mdnsGroup = net.ParseIP("ff02::fb")
	}

	// 创建 UDP 连接
	udpConn, err := net.ListenUDP(network, addr)
	if err != nil {
		log.Printf("无法绑定 %s mDNS 通配符地址: %v", protocol, err)
		return
	}
	defer udpConn.Close()

	// 根据协议加入多播组
	var failedJoins int
	if protocol == "IPv4" {
		pkConn := ipv4.NewPacketConn(udpConn)
		if err := pkConn.SetControlMessage(ipv4.FlagInterface, true); err != nil {
			log.Printf("设置 IPv4 控制消息失败: %v", err)
		}
		defer pkConn.Close()

		for _, iface := range interfaces {
			if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroup}); err != nil {
				log.Printf("加入 IPv4 多播组失败 (接口 %s): %v", iface.Name, err)
				failedJoins++
			}
		}
	} else {
		pkConn := ipv6.NewPacketConn(udpConn)
		if err := pkConn.SetControlMessage(ipv6.FlagInterface, true); err != nil {
			log.Printf("设置 IPv6 控制消息失败: %v", err)
		}
		defer pkConn.Close()

		for _, iface := range interfaces {
			if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroup}); err != nil {
				log.Printf("加入 IPv6 多播组失败 (接口 %s): %v", iface.Name, err)
				failedJoins++
			}
		}
	}

	if failedJoins == len(interfaces) {
		log.Printf("无法加入任何 %s 多播组", protocol)
		return
	}

	// 设置读取缓冲区大小
	if err := udpConn.SetReadBuffer(1024 * 1024); err != nil { // 1MB
		log.Printf("设置读取缓冲区大小失败: %v", err)
	}

	log.Printf("%s mDNS 查询捕获器已启动", protocol)

	// 处理数据包
	processPackets(cache, parser, appIDPrefix, ctx, udpConn, protocol)
}

// processPackets 通用数据包处理逻辑
func processPackets(cache *cache.ServiceCache, parser *parser.DaprMDNSParser, appIDPrefix string, ctx context.Context, conn net.PacketConn, protocol string) {
	buf := make([]byte, 65536)
	dnsMsg := new(dns.Msg)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s mDNS 查询捕获器停止", protocol)
			return
		default:
			// 设置读取超时，以便可以定期检查上下文
			if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
				log.Printf("设置读取超时失败: %v", err)
			}
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				// 忽略超时错误
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}

				log.Printf("读取 %s mDNS 数据包错误: %v", protocol, err)
				continue
			}

			// 解析 DNS 消息
			err = dnsMsg.Unpack(buf[:n])
			if err != nil {
				log.Printf("解析 %s DNS 消息错误: %v", protocol, err)
				continue
			}

			// 只处理查询（QR=0）
			if dnsMsg.Response {
				continue
			}

			// 处理每个查询问题
			for _, q := range dnsMsg.Question {
				serviceName := q.Name
				// 提取 AppID
				appID := parser.ExtractAppIDFromServiceName(serviceName)
				if appID == "" || !strings.HasPrefix(appID, appIDPrefix) {
					continue
				}

				// 创建基本的服务信息（只有 AppID）
				info := &types.DaprServiceInfo{
					AppID:       appID,
					ServiceName: serviceName,
				}

				// 存储到缓存
				_, exists := cache.Get(appID)
				if !exists {
					cache.Add(appID, info)
					log.Printf("捕获到新服务: %s", appID)
					services := parser.LookupDaprService(ctx, appID, 100*time.Millisecond)

					if len(services) > 0 {
						log.Printf("刷新服务: AppID=%s, 找到 %d 个实例", appID, len(services))
						cache.AddOrUpdateAll(appID, services)
					} else {
						log.Printf("服务未找到或已离线: AppID=%s", appID)
					}
				}
			}
		}
	}
}
