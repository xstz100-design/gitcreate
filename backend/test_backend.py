"""测试后端JWT是否正常"""
import sys
sys.path.insert(0, 'C:/Users/Administrator/Desktop/vue/backend')

from app.core.security import create_access_token, decode_token

# 测试JWT编码解码
print("="*50)
print("测试JWT编码解码")
print("="*50)

# 创建token
token = create_access_token(data={"sub": 1})
print(f"\n✅ Token创建成功")

# 解码token
payload = decode_token(token)
print(f"\n✅ Token解码成功: {payload}")

if payload and payload.get("sub") == 1:
    print("\n🎉 JWT工作正常！")
else:
    print("\n❌ JWT解码失败！")
