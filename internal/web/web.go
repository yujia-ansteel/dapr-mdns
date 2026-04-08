// Copyright 2026 pyschemDa <guangdashao1990@163.com>
//
// Licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"dapr-mdns-resolver/internal/cache"
	"dapr-mdns-resolver/internal/parser"
	"dapr-mdns-resolver/pkg/types"
)

//go:embed templates/**
var templateFS embed.FS

// WebServer Web 服务器
type WebServer struct {
	cache    *cache.ServiceCache
	parser   *parser.DaprMDNSParser
	port     int
	basePath string
	server   *http.Server
}

// NewWebServer 创建新的 Web 服务器
func NewWebServer(cache *cache.ServiceCache, port int, basePath string) *WebServer {
	parser, err := parser.NewDaprMDNSParser()
	if err != nil {
		log.Printf("创建解析器失败: %v", err)
		// 继续运行，parser 为 nil，搜索功能将不可用
	}
	return &WebServer{
		cache:    cache,
		parser:   parser,
		port:     port,
		basePath: basePath,
	}
}

// normalizeBasePath 规范化基础路径，确保以斜杠开头且不以斜杠结尾（除非是根路径）
func (ws *WebServer) normalizeBasePath() {
	basePath := ws.basePath
	if basePath == "" {
		basePath = "/"
	}
	// 确保以斜杠开头
	if basePath[0] != '/' {
		basePath = "/" + basePath
	}
	// 确保不以斜杠结尾（除非是根路径）
	if basePath != "/" && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}
	ws.basePath = basePath
}

// cacheItemToMap 将 CacheItem 转换为 JSON 序列化的 map
func cacheItemToMap(item *types.CacheItem) map[string]interface{} {
	return map[string]interface{}{
		"AppID":         item.ServiceInfo.AppID,
		"ServiceName":   item.ServiceInfo.ServiceName,
		"InstanceName":  item.ServiceInfo.InstanceName,
		"Host":          item.ServiceInfo.Host,
		"Port":          item.ServiceInfo.Port,
		"IPv4Addresses": item.ServiceInfo.IPv4Addresses,
		"IPv6Addresses": item.ServiceInfo.IPv6Addresses,
		"TextRecords":   item.ServiceInfo.TextRecords,
		"FirstSeen":     item.FirstSeen.Format(time.RFC3339),
		"LastUpdated":   item.LastUpdated.Format(time.RFC3339),
		"LastSeen":      item.LastSeen.Format(time.RFC3339),
	}
}

// loggingMiddleware 记录所有请求的中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建自定义 ResponseWriter 来捕获状态码
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		log.Printf("请求开始: %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		log.Printf("请求完成: %s %s -> %d (%v)", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter 自定义 ResponseWriter 用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Start 启动 Web 服务器
func (ws *WebServer) Start() error {
	// 规范化基础路径
	//basePath := ws.normalizeBasePath()

	internalMux := http.NewServeMux()

	// 静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	internalMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API 路由
	internalMux.HandleFunc("/api/services", ws.handleAPIServices)
	internalMux.HandleFunc("/api/services/", ws.handleAPIServiceDetail)
	internalMux.HandleFunc("/api/search", ws.handleAPISearch)

	// 主页面
	internalMux.HandleFunc("/", ws.handleIndex)

	// 健康检查端点
	internalMux.HandleFunc("/health", ws.handleHealth)

	// 创建最终处理器
	var finalHandler http.Handler
	if ws.basePath == "/" || ws.basePath == "" {
		// 无基础路径，直接使用内部 mux
		finalHandler = internalMux
		log.Printf("路由注册完成（无基础路径）:")
	} else {
		// 自定义基础路径包装器，确保空路径被重写为 "/"
		finalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 保存原始路径用于日志记录
			originalPath := r.URL.Path
			// 移除基础路径前缀
			if strings.HasPrefix(originalPath, ws.basePath) {
				remaining := originalPath[len(ws.basePath):]
				if remaining == "" {
					remaining = "/"
				}
				// 更新请求路径
				r.URL.Path = remaining
				// 调用内部处理器
				internalMux.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})
		log.Printf("路由注册完成（基础路径: %s）:", ws.basePath)
	}
	log.Printf("  %sstatic/ -> static file server", ws.basePath)
	log.Printf("  %sapi/services -> handleAPIServices", ws.basePath)
	log.Printf("  %sapi/services/ -> handleAPIServiceDetail", ws.basePath)
	log.Printf("  %sapi/search -> handleAPISearch", ws.basePath)
	log.Printf("  %s -> handleIndex", ws.basePath)
	log.Printf("  %shealth -> handleHealth", ws.basePath)

	// 使用日志中间件包装所有处理程序
	loggingHandler := loggingMiddleware(finalHandler)

	ws.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", ws.port),
		Handler:      loggingHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Web 服务器启动，监听端口 %d，基础路径 %s", ws.port, ws.basePath)
	log.Printf("开始监听请求...")
	err := ws.server.ListenAndServe()
	if err != nil {
		log.Printf("Web 服务器停止: %v", err)
	}
	return err
}

// Stop 停止 Web 服务器
func (ws *WebServer) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}

// loadTemplate 加载 HTML 模板
func (ws *WebServer) loadTemplate() (*template.Template, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	return tmpl, err
}

// handleIndex 处理首页请求
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	// 如果路径为空或根路径，都视为首页
	if r.URL.Path != "/" && r.URL.Path != "" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := ws.loadTemplate()
	if err != nil {
		http.Error(w, fmt.Sprintf("模板加载错误: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取所有缓存的服务并按 AppID 分组
	items := ws.cache.GetAll()

	// 按 AppID 分组并计算过期状态
	itemsByAppID := make(map[string][]*types.CacheItem)
	for _, item := range items {
		appID := item.ServiceInfo.AppID
		itemsByAppID[appID] = append(itemsByAppID[appID], item)
	}

	// 构建分组列表
	type InstanceView struct {
		*types.CacheItem
		IsExpired bool
	}
	type AppGroup struct {
		AppID     string
		Instances []InstanceView
	}
	now := time.Now()
	groups := make([]AppGroup, 0, len(itemsByAppID))
	for appID, instances := range itemsByAppID {
		instanceViews := make([]InstanceView, 0, len(instances))
		for _, item := range instances {
			isExpired := now.Sub(item.LastSeen) > 60*time.Second
			instanceViews = append(instanceViews, InstanceView{
				CacheItem: item,
				IsExpired: isExpired,
			})
		}
		groups = append(groups, AppGroup{
			AppID:     appID,
			Instances: instanceViews,
		})
	}

	data := struct {
		Now      time.Time
		Groups   []AppGroup
		BasePath string
	}{
		Now:      now,
		Groups:   groups,
		BasePath: ws.basePath,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("模板执行失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleAPIServices 处理 API 服务列表请求
func (ws *WebServer) handleAPIServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "只支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	items := ws.cache.GetAll()
	services := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		services = append(services, cacheItemToMap(item))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		log.Printf("JSON 编码失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleAPIServiceDetail 处理 API 服务详情请求
func (ws *WebServer) handleAPIServiceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "只支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取 AppID
	appID := r.URL.Path[len("/api/services/"):]
	if appID == "" {
		http.Error(w, "缺少 AppID", http.StatusBadRequest)
		return
	}

	items, exists := ws.cache.Get(appID)
	if !exists {
		http.Error(w, "服务未找到", http.StatusNotFound)
		return
	}

	services := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		services = append(services, cacheItemToMap(item))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		log.Printf("JSON 编码失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleAPISearch 处理搜索指定 AppID 的服务
func (ws *WebServer) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "只支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("appID")
	if appID == "" {
		http.Error(w, "缺少 appID 参数", http.StatusBadRequest)
		return
	}

	if ws.parser == nil {
		http.Error(w, "解析器不可用", http.StatusInternalServerError)
		return
	}

	// 设置查询超时
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// 实时查询 mDNS
	services := ws.parser.LookupDaprService(ctx, appID, 2*time.Second)

	if len(services) > 0 {
		// 添加到缓存
		ws.cache.AddOrUpdateAll(appID, services)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"appID":   appID,
			"found":   len(services),
			"message": fmt.Sprintf("找到 %d 个实例，已添加到缓存", len(services)),
		})
	} else {
		// 从缓存中删除
		ws.cache.Remove(appID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"appID":   appID,
			"found":   0,
			"message": "未找到实例，已从缓存中删除",
		})
	}
}

// handleHealth 处理健康检查请求
func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取所有AppID用于诊断
	appIDs := ws.cache.GetAppIDs()

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "dapr-mdns-resolver",
		"version":   "1.0.0",
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "unknown", // 可以添加启动时间跟踪
		"cache": map[string]interface{}{
			"count":         ws.cache.Count(),
			"app_ids":       appIDs,
			"app_ids_count": len(appIDs),
		},
		"endpoints": []string{
			"/",
			"/health",
			"/api/services",
			"/api/services/{appID}",
			"/api/search",
			"/static/",
		},
		"note": "If you see different JSON format, you might be hitting a different service on port 8080",
	}); err != nil {
		log.Printf("JSON 编码失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}
