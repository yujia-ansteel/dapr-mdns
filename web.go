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
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

// WebServer Web 服务器
type WebServer struct {
	cache  *ServiceCache
	port   int
	server *http.Server
}

// NewWebServer 创建新的 Web 服务器
func NewWebServer(cache *ServiceCache, port int) *WebServer {
	return &WebServer{
		cache: cache,
		port:  port,
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
	mux := http.NewServeMux()

	// 静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API 路由
	mux.HandleFunc("/api/services", ws.handleAPIServices)
	mux.HandleFunc("/api/services/", ws.handleAPIServiceDetail)

	// 主页面
	mux.HandleFunc("/", ws.handleIndex)

	// 健康检查端点
	mux.HandleFunc("/health", ws.handleHealth)

	log.Printf("路由注册完成:")
	log.Printf("  /static/ -> static file server")
	log.Printf("  /api/services -> handleAPIServices")
	log.Printf("  /api/services/ -> handleAPIServiceDetail")
	log.Printf("  / -> handleIndex")
	log.Printf("  /health -> handleHealth")

	// 使用日志中间件包装所有处理程序
	loggingHandler := loggingMiddleware(mux)

	ws.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", ws.port),
		Handler:      loggingHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Web 服务器启动，监听端口 %d", ws.port)
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

// handleIndex 处理首页请求
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	//log.Printf("handleIndex called with path: %s", r.URL.Path)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 简单的 HTML 模板
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dapr mDNS 服务发现</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            padding: 30px;
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        .stats {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .service-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        .service-card {
            border: 1px solid #ddd;
            border-radius: 5px;
            padding: 20px;
            background-color: white;
            transition: box-shadow 0.3s;
        }
        .service-card:hover {
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .service-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .app-id {
            font-weight: bold;
            font-size: 1.2em;
            color: #2196F3;
        }
        .status {
            padding: 3px 10px;
            border-radius: 12px;
            font-size: 0.9em;
            color: white;
            background-color: #4CAF50;
        }
        .info-row {
            margin: 8px 0;
            display: flex;
        }
        .label {
            font-weight: 600;
            color: #666;
            min-width: 100px;
        }
        .value {
            color: #333;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            color: #888;
            font-size: 0.9em;
        }
        .refresh-btn {
            background-color: #4CAF50;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .refresh-btn:hover {
            background-color: #45a049;
        }
        .service-card.expired {
            border-color: #f44336;
            background-color: #ffebee;
        }
        .status-expired {
            background-color: #f44336;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Dapr mDNS 服务发现</h1>
        
        <div class="stats">
            <div>
                <strong>发现的服务数量:</strong> <span id="service-count">0</span>
            </div>
            <div>
                <strong>最后更新时间:</strong> <span id="last-update">{{.Now.Format "2006-01-02 15:04:05"}}</span>
            </div>
            <button class="refresh-btn" onclick="refreshData()">手动刷新</button>
        </div>
        
        <div id="service-list" class="service-list">
            {{if .Groups}}
                {{range .Groups}}
                <h2 style="grid-column: 1 / -1; margin: 20px 0 10px; padding-bottom: 5px; border-bottom: 2px solid #4CAF50;">
                    {{.AppID}} ({{len .Instances}} 个实例)
                </h2>
                    {{range .Instances}}
                    <div class="service-card {{if .IsExpired}}expired{{end}}">
                        <div class="service-header">
                            <div class="app-id">{{.ServiceInfo.AppID}}</div>
                            <div class="status {{if .IsExpired}}status-expired{{else}}status-online{{end}}">{{if .IsExpired}}离线{{else}}在线{{end}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">实例名称:</div>
                            <div class="value">{{.ServiceInfo.InstanceName}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">主机名:</div>
                            <div class="value">{{.ServiceInfo.Host}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">端口:</div>
                            <div class="value">{{.ServiceInfo.Port}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">IPv4:</div>
                            <div class="value">{{range .ServiceInfo.IPv4Addresses}}{{.}} {{end}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">首次发现:</div>
                            <div class="value">{{.FirstSeen.Format "2006-01-02 15:04:05"}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">最后更新:</div>
                            <div class="value">{{.LastUpdated.Format "2006-01-02 15:04:05"}}</div>
                        </div>
                        <div class="info-row">
                            <div class="label">最后发现:</div>
                            <div class="value">{{.LastSeen.Format "2006-01-02 15:04:05"}}</div>
                        </div>
                    </div>
                    {{end}}
                {{end}}
            {{else}}
                <p style="grid-column: 1 / -1; text-align: center; color: #888;">未发现任何服务</p>
            {{end}}
        </div>
        
        <div class="footer">
            <p>数据每 5 秒自动刷新 | Dapr mDNS 发现工具</p>
        </div>
    </div>

    <script>
let refreshInterval = 5000; // 5 秒

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function fetchServices() {
            fetch('/api/services')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('service-count').textContent = data.length;
                    renderServices(data);
                    document.getElementById('last-update').textContent = new Date().toLocaleString();
                })
                .catch(error => {
                    console.error('获取服务数据失败:', error);
                });
        }
        
        function renderServices(services) {
            const container = document.getElementById('service-list');
            if (services.length === 0) {
                container.innerHTML = '<p style="grid-column: 1 / -1; text-align: center; color: #888;">未发现任何服务</p>';
                return;
            }

            // 按 AppID 分组
            const groups = {};
            const now = new Date();
            services.forEach(service => {
                const appID = service.AppID;
                if (!groups[appID]) {
                    groups[appID] = [];
                }
                // 计算是否过期：最后发现时间超过60秒
                const lastSeen = service.LastSeen ? new Date(service.LastSeen) : null;
                const isExpired = lastSeen ? (now - lastSeen) > 60000 : true; // 60秒 = 60000毫秒
                groups[appID].push({ ...service, isExpired });
            });

            // 构建 HTML
            let html = '';
            Object.keys(groups).sort().forEach(appID => {
                const instances = groups[appID];
                html += '<h2 style="grid-column: 1 / -1; margin: 20px 0 10px; padding-bottom: 5px; border-bottom: 2px solid #4CAF50;">' +
                        escapeHtml(appID) + ' (' + instances.length + ' 个实例)' +
                        '</h2>';
                instances.forEach(service => {
                    const instanceName = service.InstanceName || 'N/A';
                    const host = service.Host || 'N/A';
                    const port = service.Port || 'N/A';
                    const ipv4 = service.IPv4Addresses ? service.IPv4Addresses.join(', ') : 'N/A';
                    const firstSeen = service.FirstSeen ? new Date(service.FirstSeen).toLocaleString() : 'N/A';
                    const lastUpdated = service.LastUpdated ? new Date(service.LastUpdated).toLocaleString() : 'N/A';
                    const lastSeen = service.LastSeen ? new Date(service.LastSeen).toLocaleString() : 'N/A';
                    const statusClass = service.isExpired ? 'status-expired' : 'status-online';
                    const statusText = service.isExpired ? '离线' : '在线';
                    const expiredClass = service.isExpired ? 'expired' : '';
                    
                    html += '<div class="service-card ' + expiredClass + '">' +
                            '    <div class="service-header">' +
                            '        <div class="app-id">' + escapeHtml(appID) + '</div>' +
                            '        <div class="status ' + statusClass + '">' + statusText + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">实例名称:</div>' +
                            '        <div class="value">' + escapeHtml(instanceName) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">主机名:</div>' +
                            '        <div class="value">' + escapeHtml(host) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">端口:</div>' +
                            '        <div class="value">' + escapeHtml(port.toString()) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">IPv4:</div>' +
                            '        <div class="value">' + escapeHtml(ipv4) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">首次发现:</div>' +
                            '        <div class="value">' + escapeHtml(firstSeen) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">最后更新:</div>' +
                            '        <div class="value">' + escapeHtml(lastUpdated) + '</div>' +
                            '    </div>' +
                            '    <div class="info-row">' +
                            '        <div class="label">最后发现:</div>' +
                            '        <div class="value">' + escapeHtml(lastSeen) + '</div>' +
                            '    </div>' +
                            '</div>';
                });
            });
            container.innerHTML = html;
        }
        
        function refreshData() {
            fetchServices();
        }
        
        // 页面加载时获取数据
        document.addEventListener('DOMContentLoaded', function() {
            fetchServices();
            // 设置定时刷新
            setInterval(fetchServices, refreshInterval);
        });
    </script>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("模板解析错误: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取所有缓存的服务并按 AppID 分组
	items := ws.cache.GetAll()

	// 按 AppID 分组并计算过期状态
	itemsByAppID := make(map[string][]*CacheItem)
	for _, item := range items {
		appID := item.ServiceInfo.AppID
		itemsByAppID[appID] = append(itemsByAppID[appID], item)
	}

	// 构建分组列表
	type InstanceView struct {
		*CacheItem
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
		Now    time.Time
		Groups []AppGroup
	}{
		Now:    now,
		Groups: groups,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// handleAPIServices 处理 API 服务列表请求
func (ws *WebServer) handleAPIServices(w http.ResponseWriter, r *http.Request) {
	//log.Printf("handleAPIServices called with path: %s", r.URL.Path)
	if r.Method != "GET" {
		http.Error(w, "只支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	items := ws.cache.GetAll()
	services := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		service := map[string]interface{}{
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
		services = append(services, service)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// handleAPIServiceDetail 处理 API 服务详情请求
func (ws *WebServer) handleAPIServiceDetail(w http.ResponseWriter, r *http.Request) {
	//log.Printf("handleAPIServiceDetail called with path: %s", r.URL.Path)
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
		service := map[string]interface{}{
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
		services = append(services, service)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// handleHealth 处理健康检查请求
func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	//log.Printf("handleHealth called with path: %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	// 获取所有AppID用于诊断
	appIDs := ws.cache.GetAppIDs()

	json.NewEncoder(w).Encode(map[string]interface{}{
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
			"/static/",
		},
		"note": "If you see different JSON format, you might be hitting a different service on port 8080",
	})
}
