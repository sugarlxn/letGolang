#!/bin/bash

# 异步任务系统集成指南

echo "================================"
echo "异步任务系统集成步骤"
echo "================================"

# 步骤 1: 更新 go.mod
echo -e "\n[步骤 1] 检查依赖..."
go get github.com/google/uuid
go mod tidy

# 步骤 2: 运行数据库迁移
echo -e "\n[步骤 2] 检查数据库迁移..."
echo "确保以下迁移已执行:"
echo "  - 0004_add_image_tasks.sql"
echo ""
echo "执行命令:"
echo "  sqlite3 webserver.db < migrations/0004_add_image_tasks.sql"

# 步骤 3: 配置环境变量
echo -e "\n[步骤 3] 配置环境变量..."
echo "在 .env 或系统环境中设置:"
echo ""
echo "# 文生图接口 URL（支持多个实例）"
echo "export IMAGE_GEN_URL_1=http://localhost:8000"
echo "export IMAGE_GEN_URL_2=http://localhost:8001  # 可选"
echo "export IMAGE_GEN_URL_3=http://localhost:8002  # 可选"
echo ""
echo "# 语音转文字接口 URL"
echo "export WHISPER_URL=http://localhost:8001"
echo ""
echo "# JWT 密钥"
echo "export JWT_SECRET=719c946d-14d8-4c9f-aac9-f807254bf447"

# 步骤 4: 编译并运行
echo -e "\n[步骤 4] 编译和运行..."
echo ""
echo "编译:"
echo "  go build -o webserver"
echo ""
echo "运行:"
echo "  ./webserver"
echo ""
echo "或者在开发环境直接运行:"
echo "  go run main.go"

# 步骤 5: 验证
echo -e "\n[步骤 5] 验证系统..."
echo ""
echo "1. 获取 token:"
echo "   curl -X POST http://localhost:8080/login \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"username\":\"user1\",\"password\":\"pass123\"}'"
echo ""
echo "2. 查看系统统计:"
echo "   curl http://localhost:8080/api/v1/system/stats \\"
echo "     -H 'Authorization: Bearer YOUR_TOKEN'"
echo ""
echo "3. 提交测试任务:"
echo "   curl -X POST http://localhost:8080/api/v1/image/async \\"
echo "     -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\":\"A beautiful sunset\"}'"

echo -e "\n================================"
echo "集成完成！"
echo "================================"
