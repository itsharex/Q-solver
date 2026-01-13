#!/bin/bash

# =================Q-Solver macOS 构建脚本=================
MODE="${1:-dev}"

echo "🍎 Q-Solver macOS 构建脚本"
echo "📂 模式: $MODE"

# macOS 不需要额外的 DLL
export CGO_ENABLED=1

case "$MODE" in
    "dev")
        echo -e "\n🚀 启动 Wails 开发模式..."
        wails dev
        ;;

    "build")
        echo -e "\n🔨 开始构建 macOS 应用..."
        
        # 构建 Universal Binary (同时支持 Intel 和 Apple Silicon)
        wails build -platform darwin/universal -ldflags "-s -w"
        
        if [ $? -eq 0 ]; then
            echo "✅ 构建完成！"
            echo "📦 应用位于: build/bin/Q-Solver.app"
        else
            echo "❌ 构建失败"
            exit 1
        fi
        ;;

    *)
        echo "❌ 未知参数: $MODE"
        echo "用法:"
        echo "  ./build_darwin.sh dev    # 启动开发模式 (默认)"
        echo "  ./build_darwin.sh build  # 编译发布版本"
        exit 1
        ;;
esac
