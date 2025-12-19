#!/bin/bash

# =================配置区域=================
# 获取 lib 文件夹的绝对路径
LIB_PATH="$(cd "$(dirname "$0")/lib" && pwd)"
# 获取脚本第一个参数，默认为 dev
MODE="${1:-dev}" 

echo "📂 依赖库路径: ${LIB_PATH}"

# =================环境变量设置=================
# 1. CGO 编译设置 (编译时用)
export CGO_ENABLED=1
export CGO_CFLAGS="-I${LIB_PATH}"
export CGO_LDFLAGS="-L${LIB_PATH} -lsherpa-onnx-c-api"

# 2. 运行时路径设置 (运行时用)
# 将 lib 加入 PATH，这样 dev 模式运行时能找到 dll，不需要复制到根目录
export PATH="${LIB_PATH}:$PATH"

echo "🔧 环境配置完成 (模式: $MODE)"

# =================模式选择=================
case "$MODE" in
    "dev")
        echo -e "\n🚀 启动 Wails 开发模式 (Dev Mode)..."
        # 启动开发模式
        wails dev
        ;;

    "build")
        echo -e "\n🔨 开始 Wails 打包 (Build Mode)..."
        
        # 执行构建
        # -clean: 清理旧的构建文件
        # -ldflags "-s -w": 去除调试符号，减小体积
        wails build  -ldflags "-s -w" -tags prod
        
        # 检查构建结果
        if [ $? -eq 0 ]; then
            echo "✅ 编译完成！"
            
            # --- 自动复制 DLL 到发布目录 ---
            if [ -d "build/bin" ]; then
                echo "📦 正在打包依赖库到 build/bin..."
                # 复制所有 dll 到 build/bin 目录下
                cp "${LIB_PATH}/"*.dll "build/bin/" 2>/dev/null || true
                echo "🎉 打包成功！可执行文件和 DLL 都在 build/bin 目录下。"
            else
                echo "⚠️ 未找到 build/bin 目录，DLL 复制跳过。"
            fi
        else
            echo "❌ 构建失败，请检查代码错误。"
            exit 1
        fi
        ;;

    *)
        echo "❌ 未知参数: $MODE"
        echo "用法:"
        echo "  ./build.sh dev    # 启动开发模式 (默认)"
        echo "  ./build.sh build  # 编译发布版本"
        exit 1
        ;;
esac