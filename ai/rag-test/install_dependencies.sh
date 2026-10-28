#!/bin/bash

# RAG 系统依赖安装脚本

echo "=== RAG 系统依赖安装 ==="
echo ""

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "检测到 macOS 系统"
    
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        echo "错误: 未找到 Homebrew"
        echo "请先安装 Homebrew: https://brew.sh"
        exit 1
    fi
    
    # Check if poppler is installed
    if command -v pdftotext &> /dev/null; then
        echo "✓ pdftotext 已安装"
    else
        echo "正在安装 poppler (包含 pdftotext)..."
        brew install poppler
        if [ $? -eq 0 ]; then
            echo "✓ poppler 安装成功"
        else
            echo "✗ poppler 安装失败"
            exit 1
        fi
    fi
    
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "检测到 Linux 系统"
    
    # Check if apt-get is available (Debian/Ubuntu)
    if command -v apt-get &> /dev/null; then
        if command -v pdftotext &> /dev/null; then
            echo "✓ pdftotext 已安装"
        else
            echo "正在安装 poppler-utils (包含 pdftotext)..."
            sudo apt-get update
            sudo apt-get install -y poppler-utils
            if [ $? -eq 0 ]; then
                echo "✓ poppler-utils 安装成功"
            else
                echo "✗ poppler-utils 安装失败"
                exit 1
            fi
        fi
    else
        echo "警告: 未检测到 apt-get，请手动安装 poppler-utils"
        echo "  RedHat/CentOS: sudo yum install poppler-utils"
        echo "  Arch Linux: sudo pacman -S poppler"
    fi
else
    echo "警告: 未识别的操作系统: $OSTYPE"
    echo "请手动安装 poppler:"
    echo "  macOS: brew install poppler"
    echo "  Ubuntu/Debian: sudo apt-get install poppler-utils"
fi

echo ""
echo "验证安装..."
if command -v pdftotext &> /dev/null; then
    echo "✓ pdftotext 已正确安装"
    pdftotext -v
else
    echo "✗ pdftotext 未找到，请检查安装"
    exit 1
fi

echo ""
echo "=== 安装完成 ==="
echo "现在可以运行 RAG 服务器了:"
echo "  go run main.go"
