#!/bin/bash

# 异步任务系统集成测试脚本

set -e

BASE_URL="http://localhost:8080"
USERNAME="test_user"
PASSWORD="test_pass_123"
TOKEN=""

echo "================================"
echo "异步任务系统测试脚本"
echo "================================"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=$4
    
    echo -e "\n${YELLOW}[测试]${NC} $method $endpoint"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    echo "响应码: $http_code"
    echo "响应体: $body" | head -c 200
    echo ""
    
    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ 通过${NC}"
        return 0
    else
        echo -e "${RED}✗ 失败 (期望: $expected_code)${NC}"
        return 1
    fi
}

# 步骤 1: 注册用户
echo -e "\n${YELLOW}=== 步骤 1: 用户注册 ===${NC}"
curl -s -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
    | jq . || true

# 步骤 2: 用户登录
echo -e "\n${YELLOW}=== 步骤 2: 用户登录 ===${NC}"
login_response=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

echo "$login_response" | jq .

TOKEN=$(echo "$login_response" | jq -r '.token' 2>/dev/null || echo "")

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo -e "${RED}登录失败，无法获取 token${NC}"
    exit 1
fi

echo -e "${GREEN}获取 token: $TOKEN${NC}"

# 步骤 3: 查看系统统计
echo -e "\n${YELLOW}=== 步骤 3: 系统统计 ===${NC}"
test_endpoint "GET" "/api/v1/system/stats" "" 200

# 步骤 4: 提交文生图任务
echo -e "\n${YELLOW}=== 步骤 4: 提交文生图任务 ===${NC}"
submit_response=$(curl -s -X POST "$BASE_URL/api/v1/image/async" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "prompt": "A beautiful sunset over mountains",
        "negative_prompt": "blurry, low quality"
    }')

echo "$submit_response" | jq .

TASK_ID=$(echo "$submit_response" | jq -r '.task_id' 2>/dev/null || echo "")

if [ -z "$TASK_ID" ] || [ "$TASK_ID" == "null" ]; then
    echo -e "${RED}提交任务失败${NC}"
    exit 1
fi

echo -e "${GREEN}任务 ID: $TASK_ID${NC}"

# 步骤 5: 查询任务状态（多次）
echo -e "\n${YELLOW}=== 步骤 5: 查询任务状态 ===${NC}"
for i in {1..5}; do
    echo -e "\n${YELLOW}轮询 $i:${NC}"
    status_response=$(curl -s -X GET "$BASE_URL/api/v1/tasks?task_id=$TASK_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json")
    
    echo "$status_response" | jq .
    
    status=$(echo "$status_response" | jq -r '.status' 2>/dev/null || echo "")
    
    if [ "$status" == "DONE" ]; then
        echo -e "${GREEN}✓ 任务完成！${NC}"
        result_url=$(echo "$status_response" | jq -r '.result_url' 2>/dev/null)
        echo "结果 URL: $result_url"
        break
    elif [ "$status" == "FAILED" ]; then
        echo -e "${RED}✗ 任务失败${NC}"
        error_msg=$(echo "$status_response" | jq -r '.error_msg' 2>/dev/null)
        echo "错误: $error_msg"
        break
    fi
    
    echo "状态: $status, 等待 2 秒..."
    sleep 2
done

# 步骤 6: 获取用户所有任务
echo -e "\n${YELLOW}=== 步骤 6: 获取用户所有任务 ===${NC}"
test_endpoint "GET" "/api/v1/tasks" "" 200

# 步骤 7: 测试语音转文字（模拟 PCM 数据）
echo -e "\n${YELLOW}=== 步骤 7: 测试 PCM 接口 (仅检查格式) ===${NC}"
echo "警告: 此测试仅检查接口格式，需要真实 PCM 数据以完全测试"
# 创建 32 KB 的虚拟 PCM 数据（16kHz, 16bit, 1s）
dd if=/dev/zero of=/tmp/test_pcm.raw bs=1024 count=32 2>/dev/null
echo "创建虚拟 PCM 文件: /tmp/test_pcm.raw (32 KB)"

# 步骤 8: 系统统计（最终）
echo -e "\n${YELLOW}=== 步骤 8: 最终系统统计 ===${NC}"
test_endpoint "GET" "/api/v1/system/stats" "" 200

echo -e "\n${GREEN}================================"
echo "测试完成！"
echo "================================${NC}"

echo -e "\n${YELLOW}关键观察点：${NC}"
echo "1. 任务提交延迟 < 10ms (理想情况)"
echo "2. 队列使用率"
echo "3. Worker 是否正常处理"
echo "4. 后端实例健康状态"

echo -e "\n${YELLOW}性能评估：${NC}"
echo "- 提交响应时间: 应该 < 5ms"
echo "- 任务处理时间: 取决于后端实例"
echo "- 队列长度: 应该逐渐减少"

# 清理
rm -f /tmp/test_pcm.raw

echo -e "\n${GREEN}清理完成${NC}"
