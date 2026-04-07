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
	"sync"
	"time"
)

// ServiceCache 缓存 Dapr 服务信息
type ServiceCache struct {
	mu    sync.RWMutex
	items map[string][]*CacheItem
}

// CacheItem 缓存项，包含服务信息和元数据
type CacheItem struct {
	ServiceInfo *DaprServiceInfo
	FirstSeen   time.Time
	LastUpdated time.Time
	LastSeen    time.Time
}

// NewServiceCache 创建新的服务缓存
func NewServiceCache() *ServiceCache {
	return &ServiceCache{
		items: make(map[string][]*CacheItem),
	}
}

// AddOrUpdateAll 添加或更新缓存项
func (c *ServiceCache) AddOrUpdateAll(appID string, infos []*DaprServiceInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiryDuration := 120 * time.Second

	// 获取该 AppID 的现有缓存项
	items := c.items[appID]

	// 第一步：清理过期或无效的缓存项
	validItems := make([]*CacheItem, 0, len(items))
	for _, item := range items {
		// 过滤条件：实例为空 或 最后更新时间超过60秒
		if item.ServiceInfo.InstanceName == "" || now.Sub(item.LastUpdated) > expiryDuration {
			continue // 跳过，相当于删除
		}
		validItems = append(validItems, item)
	}

	// 第二步：用 infos 批量更新/添加
	for _, info := range infos {
		found := false
		for _, item := range validItems {
			if item.ServiceInfo.InstanceName == info.InstanceName {
				// 已存在，更新信息
				item.ServiceInfo = info
				item.LastUpdated = now
				item.LastSeen = now
				found = true
				break
			}
		}
		// 不存在，作为新实例添加
		if !found {
			validItems = append(validItems, &CacheItem{
				ServiceInfo: info,
				FirstSeen:   now,
				LastUpdated: now,
				LastSeen:    now,
			})
		}
	}

	// 写回缓存
	c.items[appID] = validItems
}

// Add 添加缓存项
func (c *ServiceCache) Add(appID string, info *DaprServiceInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// 如果 AppID 不存在，添加新实例

	c.items[appID] = []*CacheItem{
		{
			ServiceInfo: info,
			FirstSeen:   now,
			LastUpdated: now,
			LastSeen:    now,
		},
	}
	return

}

// Get 获取缓存项列表
func (c *ServiceCache) Get(appID string) ([]*CacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items, exists := c.items[appID]
	return items, exists
}

// GetAll 获取所有缓存项
func (c *ServiceCache) GetAll() []*CacheItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 计算总实例数作为容量
	total := 0
	for _, instances := range c.items {
		total += len(instances)
	}
	items := make([]*CacheItem, 0, total)
	for _, instances := range c.items {
		items = append(items, instances...)
	}
	return items
}

// Remove 移除缓存项
func (c *ServiceCache) Remove(appID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, appID)
}

// Clear 清空缓存
func (c *ServiceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string][]*CacheItem)
}

// Count 返回缓存项数量
func (c *ServiceCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// GetAppIDs 返回所有 AppID 列表
func (c *ServiceCache) GetAppIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.items))
	for appID := range c.items {
		ids = append(ids, appID)
	}
	return ids
}
