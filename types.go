// Copyright 2026 pyschemDa <guangdashao1990@163.com>
//
// Licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package main

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
