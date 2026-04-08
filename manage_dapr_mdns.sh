#!/bin/bash
# Copyright 2026 pyschemDa <guangdashao1990@163.com>
#
# Licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
# 应用管理脚本: manage_dapr_mdns.sh
# 用法: ./manage_dapr_mdns.sh {start|stop|status|restart}

APP_NAME="dapr-mdns-resolver"
APP_BIN="./dapr-mdns-resolver-linux-arm64"
APP_ARGS="-server -app-id-prefix yj3dev -web-base-path /dapr-mdns -web-port 8881"
LOG_FILE="mdns.log"
PID_FILE="dapr_mdns.pid"
SCRIPT_NAME=$(basename "$0")

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查应用二进制文件是否存在
check_binary() {
    if [ ! -f "$APP_BIN" ]; then
        print_error "应用二进制文件不存在: $APP_BIN"
        exit 1
    fi
    
    if [ ! -x "$APP_BIN" ]; then
        chmod +x "$APP_BIN"
        print_info "已添加执行权限: $APP_BIN"
    fi
}

# 启动应用
start_app() {
    print_info "正在启动 $APP_NAME..."
    
    # 检查是否已经在运行
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "$APP_NAME 已经在运行 (PID: $PID)"
            return
        fi
    fi
    
    # 检查二进制文件
    check_binary
    
    # 启动应用
    nohup "$APP_BIN" $APP_ARGS > "$LOG_FILE" 2>&1 &
    
    # 获取进程ID
    APP_PID=$!
    
    # 写入PID文件
    echo "$APP_PID" > "$PID_FILE"
    
    # 等待一下确保进程启动
    sleep 2
    
    # 验证进程是否在运行
    if ps -p "$APP_PID" > /dev/null 2>&1; then
        print_success "$APP_NAME 启动成功!"
        print_info "PID: $APP_PID"
        print_info "PID已保存到: $PID_FILE"
        print_info "日志文件: $LOG_FILE"
        print_info "应用运行中..."
    else
        print_error "$APP_NAME 启动失败，请检查日志: $LOG_FILE"
        rm -f "$PID_FILE" 2>/dev/null
        exit 1
    fi
}

# 停止应用
stop_app() {
    print_info "正在停止 $APP_NAME..."
    
    if [ ! -f "$PID_FILE" ]; then
        print_warning "PID文件不存在: $PID_FILE"
        print_info "尝试查找并停止相关进程..."
        
        # 尝试通过进程名查找
        PIDS=$(pgrep -f "$APP_BIN" 2>/dev/null)
        
        if [ -z "$PIDS" ]; then
            print_info "$APP_NAME 没有在运行"
            return
        fi
        
        for PID in $PIDS; do
            kill "$PID" 2>/dev/null
            print_info "已发送停止信号到进程: $PID"
        done
        
        sleep 2
        
        # 检查是否还有相关进程
        REMAINING_PIDS=$(pgrep -f "$APP_BIN" 2>/dev/null)
        if [ -n "$REMAINING_PIDS" ]; then
            print_warning "进程仍在运行，发送强制停止信号..."
            kill -9 $REMAINING_PIDS 2>/dev/null
        fi
        
    else
        PID=$(cat "$PID_FILE")
        
        if ps -p "$PID" > /dev/null 2>&1; then
            # 尝试正常停止
            kill "$PID" 2>/dev/null
            sleep 2
            
            # 如果进程还在，强制停止
            if ps -p "$PID" > /dev/null 2>&1; then
                print_warning "进程 $PID 仍在运行，发送强制停止信号..."
                kill -9 "$PID" 2>/dev/null
                sleep 1
            fi
            
            if ps -p "$PID" > /dev/null 2>&1; then
                print_error "无法停止进程 $PID"
                return 1
            else
                print_success "$APP_NAME 已停止 (PID: $PID)"
            fi
        else
            print_warning "进程 $PID 不存在"
        fi
        
        # 删除PID文件
        rm -f "$PID_FILE"
    fi
    
    # 清理可能残留的PID文件
    rm -f "$PID_FILE" 2>/dev/null
    
    print_info "已清理PID文件"
}

# 查看应用状态
status_app() {
    print_info "$APP_NAME 状态检查..."
    
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            print_success "$APP_NAME 正在运行"
            print_info "PID: $PID"
            print_info "启动命令: $APP_BIN $APP_ARGS"
            
            # 显示更多进程信息
            if command -v ps > /dev/null 2>&1; then
                echo ""
                print_info "进程详细信息:"
                ps -fp "$PID" 2>/dev/null || true
            fi
            
            return 0
        else
            print_warning "PID文件存在但进程未运行"
            print_info "PID: $PID"
            return 1
        fi
    else
        # 尝试通过进程名查找
        PIDS=$(pgrep -f "$APP_BIN" 2>/dev/null)
        
        if [ -n "$PIDS" ]; then
            print_warning "PID文件不存在，但找到相关进程"
            for PID in $PIDS; do
                if ps -p "$PID" > /dev/null 2>&1; then
                    print_info "找到进程: $PID"
                fi
            done
            return 1
        else
            print_info "$APP_NAME 没有在运行"
            return 1
        fi
    fi
}

# 重启应用
restart_app() {
    stop_app
    sleep 3
    start_app
}

# 查看日志
view_log() {
    if [ -f "$LOG_FILE" ]; then
        print_info "显示日志最后100行 (按Ctrl+C退出):"
        echo "========================================"
        tail -100f "$LOG_FILE"
    else
        print_error "日志文件不存在: $LOG_FILE"
    fi
}

# 显示使用说明
show_usage() {
    echo "用法: $SCRIPT_NAME {start|stop|restart|status|log|help}"
    echo ""
    echo "命令选项:"
    echo "  start    启动应用"
    echo "  stop     停止应用"
    echo "  restart  重启应用"
    echo "  status   查看应用状态"
    echo "  log      查看应用日志"
    echo "  help     显示此帮助信息"
    echo ""
    echo "应用信息:"
    echo "  应用名称: $APP_NAME"
    echo "  启动命令: $APP_BIN $APP_ARGS"
    echo "  日志文件: $LOG_FILE"
    echo "  PID文件:  $PID_FILE"
}

# 主函数
main() {
    case "$1" in
        start)
            start_app
            ;;
        stop)
            stop_app
            ;;
        restart)
            restart_app
            ;;
        status)
            status_app
            ;;
        log)
            view_log
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            if [ -z "$1" ]; then
                show_usage
            else
                print_error "未知命令: $1"
                echo ""
                show_usage
                exit 1
            fi
            ;;
    esac
}

# 运行主函数
main "$1"
