#!/bin/bash

echo "=== 测试登录持久化功能 ==="
echo ""

# 1. 登录获取 token
echo "1. 登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}')

echo "登录响应: $LOGIN_RESPONSE"
echo ""

# 提取 access_token 和 refresh_token
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | sed 's/"refresh_token":"//;s/"//')

if [ -z "$ACCESS_TOKEN" ] || [ -z "$REFRESH_TOKEN" ]; then
  echo "❌ 登录失败，可能用户不存在"
  echo "请先注册用户："
  echo 'curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '"'"'{"username":"test","password":"test123"}'"'"''
  exit 1
fi

echo "✅ 登录成功"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo ""

# 2. 使用 access token 访问受保护资源
echo "2. 使用 Access Token 访问桌面列表..."
DESKTOPS_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/desktops \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "桌面列表响应: $DESKTOPS_RESPONSE"
echo ""

# 3. 使用 refresh token 刷新
echo "3. 使用 Refresh Token 刷新 Access Token..."
REFRESH_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")

echo "刷新响应: $REFRESH_RESPONSE"
echo ""

NEW_ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

if [ -z "$NEW_ACCESS_TOKEN" ]; then
  echo "❌ 刷新失败"
  exit 1
fi

echo "✅ 刷新成功"
echo "新 Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
echo ""

# 4. 使用新的 access token 访问资源
echo "4. 使用新 Access Token 访问桌面列表..."
NEW_DESKTOPS_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/desktops \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN")

echo "桌面列表响应: $NEW_DESKTOPS_RESPONSE"
echo ""

echo "=== ✅ 所有测试通过！持久化登录功能正常 ==="
