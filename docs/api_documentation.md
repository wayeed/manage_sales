# API 接口文档

> 家具销售提成管理系统 - 后端接口说明文档

## 目录

- [1. 概述](#1-概述)
- [2. 认证模块](#2-认证模块)
- [3. 用户管理](#3-用户管理)
- [4. 角色权限](#4-角色权限)
- [5. 商品管理](#5-商品管理)
- [6. 库存管理](#6-库存管理)
- [7. 订单管理](#7-订单管理)
- [8. 提成管理](#8-提成管理)
- [9. 工资管理](#9-工资管理)
- [10. 数据分析](#10-数据分析)
- [11. 系统配置](#11-系统配置)

---

## 1. 概述

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| 基础路径 | `/api` |
| 数据格式 | JSON |
| 字符编码 | UTF-8 |
| 认证方式 | Bearer Token (JWT) |

### 1.2 通用请求头

```
Content-Type: application/json
Authorization: Bearer <token>
```

### 1.3 通用响应格式

**成功响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

**分页响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

**错误响应：**

```json
{
  "code": 40001,
  "message": "用户名或密码错误",
  "data": null
}
```

### 1.4 通用错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 认证失败 |
| 40003 | Token 过期 |
| 40301 | 无权限 |
| 40401 | 资源不存在 |
| 42201 | 参数校验失败 |
| 50001 | 服务器内部错误 |

---

## 2. 认证模块

### 2.1 登录

用户登录获取访问令牌。

**请求**

```
POST /api/login
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例**

```json
{
  "username": "admin",
  "password": "123456"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-05-06T08:00:00Z",
    "user": {
      "id": 1,
      "username": "admin",
      "real_name": "系统管理员",
      "role": "admin",
      "avatar": ""
    }
  }
}
```

---

### 2.2 登出

退出登录，使当前 Token 失效。

**请求**

```
POST /api/logout
```

**请求头**

```
Authorization: Bearer <token>
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 2.3 获取当前用户信息

获取当前登录用户的详细信息及权限。

**请求**

```
GET /api/users/me
```

**请求头**

```
Authorization: Bearer <token>
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "real_name": "系统管理员",
    "phone": "13800138000",
    "email": "admin@example.com",
    "avatar": "",
    "role": "admin",
    "store_id": 1,
    "store_name": "总店",
    "status": 1,
    "permissions": ["*"],
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

---

## 3. 用户管理

### 3.1 用户列表

获取用户分页列表，支持关键字搜索和条件筛选。

**请求**

```
GET /api/users
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字（用户名/姓名/手机号） |
| role | string | 否 | 角色筛选 |
| status | int | 否 | 状态筛选（0-禁用，1-启用） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "username": "admin",
        "real_name": "系统管理员",
        "phone": "13800138000",
        "role": "admin",
        "store_id": 1,
        "store_name": "总店",
        "status": 1,
        "last_login_at": "2026-05-05T08:00:00Z",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 3.2 创建用户

创建新用户账号。

**请求**

```
POST /api/users
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，唯一 |
| password | string | 是 | 密码，最少 6 位 |
| real_name | string | 是 | 真实姓名 |
| phone | string | 否 | 手机号 |
| email | string | 否 | 邮箱 |
| role | string | 是 | 角色 |
| store_id | int | 否 | 所属门店 ID |

**请求示例**

```json
{
  "username": "zhangsan",
  "password": "123456",
  "real_name": "张三",
  "phone": "13900139000",
  "role": "salesman",
  "store_id": 1
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "username": "zhangsan",
    "real_name": "张三",
    "phone": "13900139000",
    "role": "salesman",
    "store_id": 1,
    "status": 1,
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 3.3 更新用户

更新指定用户的信息。

**请求**

```
PUT /api/users/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 用户 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| real_name | string | 否 | 真实姓名 |
| phone | string | 否 | 手机号 |
| email | string | 否 | 邮箱 |
| store_id | int | 否 | 所属门店 ID |
| status | int | 否 | 状态（0-禁用，1-启用） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "username": "zhangsan",
    "real_name": "张三",
    "phone": "13900139000",
    "role": "salesman",
    "store_id": 1,
    "status": 1,
    "updated_at": "2026-05-05T10:30:00Z"
  }
}
```

---

### 3.4 删除用户

删除指定用户（逻辑删除）。

**请求**

```
DELETE /api/users/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 用户 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3.5 重置密码

重置指定用户的密码。

**请求**

```
POST /api/users/:id/reset-password
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 用户 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| new_password | string | 是 | 新密码，最少 6 位 |

**请求示例**

```json
{
  "new_password": "654321"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3.6 分配角色

为指定用户分配角色。

**请求**

```
POST /api/users/:id/roles
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 用户 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role_ids | []int | 是 | 角色 ID 列表 |

**请求示例**

```json
{
  "role_ids": [2, 3]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 4. 角色权限

### 4.1 角色列表

获取系统角色列表。

**请求**

```
GET /api/roles
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "admin",
        "display_name": "系统管理员",
        "description": "拥有所有权限",
        "permission_count": 128,
        "user_count": 2,
        "created_at": "2026-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "name": "store_manager",
        "display_name": "门店经理",
        "description": "管理门店相关业务",
        "permission_count": 64,
        "user_count": 5,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 8,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 4.2 创建角色

创建新角色。

**请求**

```
POST /api/roles
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 角色标识，唯一 |
| display_name | string | 是 | 角色显示名称 |
| description | string | 否 | 角色描述 |
| permission_ids | []int | 否 | 权限 ID 列表 |

**请求示例**

```json
{
  "name": "warehouse_keeper",
  "display_name": "仓库管理员",
  "description": "管理仓库库存相关操作",
  "permission_ids": [10, 11, 12, 13]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 9,
    "name": "warehouse_keeper",
    "display_name": "仓库管理员",
    "description": "管理仓库库存相关操作",
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 4.3 更新角色

更新指定角色信息。

**请求**

```
PUT /api/roles/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 角色 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| display_name | string | 否 | 角色显示名称 |
| description | string | 否 | 角色描述 |
| permission_ids | []int | 否 | 权限 ID 列表（全量更新） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 9,
    "name": "warehouse_keeper",
    "display_name": "仓库管理员",
    "description": "管理仓库库存相关操作",
    "updated_at": "2026-05-05T10:30:00Z"
  }
}
```

---

### 4.4 删除角色

删除指定角色。

**请求**

```
DELETE /api/roles/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 角色 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 4.5 权限列表

获取系统所有权限列表。

**请求**

```
GET /api/permissions
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "user:list",
        "display_name": "用户列表",
        "module": "用户管理",
        "type": "menu"
      },
      {
        "id": 2,
        "name": "user:create",
        "display_name": "创建用户",
        "module": "用户管理",
        "type": "button"
      }
    ],
    "total": 128
  }
}
```

---

### 4.6 权限树

获取权限树形结构，用于角色权限分配。

**请求**

```
GET /api/permissions/tree
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "user",
      "display_name": "用户管理",
      "type": "menu",
      "children": [
        {
          "id": 2,
          "name": "user:list",
          "display_name": "用户列表",
          "type": "menu",
          "children": [
            {
              "id": 3,
              "name": "user:create",
              "display_name": "创建用户",
              "type": "button"
            },
            {
              "id": 4,
              "name": "user:update",
              "display_name": "编辑用户",
              "type": "button"
            }
          ]
        }
      ]
    }
  ]
}
```

---

## 5. 商品管理

### 5.1 品类列表

获取商品品类列表。

**请求**

```
GET /api/categories
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |
| parent_id | int | 否 | 父级品类 ID，不传则查所有 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "沙发",
        "parent_id": 0,
        "level": 1,
        "sort_order": 1,
        "product_count": 25,
        "status": 1,
        "children": [
          {
            "id": 2,
            "name": "布艺沙发",
            "parent_id": 1,
            "level": 2,
            "sort_order": 1,
            "product_count": 10,
            "status": 1
          }
        ]
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 5.2 创建品类

创建新的商品品类。

**请求**

```
POST /api/categories
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 品类名称 |
| parent_id | int | 否 | 父级品类 ID，默认 0（顶级） |
| sort_order | int | 否 | 排序序号 |
| icon | string | 否 | 品类图标 |

**请求示例**

```json
{
  "name": "实木家具",
  "parent_id": 0,
  "sort_order": 10
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 16,
    "name": "实木家具",
    "parent_id": 0,
    "level": 1,
    "sort_order": 10,
    "status": 1,
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 5.3 更新品类

更新指定品类信息。

**请求**

```
PUT /api/categories/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 品类 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 品类名称 |
| parent_id | int | 否 | 父级品类 ID |
| sort_order | int | 否 | 排序序号 |
| icon | string | 否 | 品类图标 |
| status | int | 否 | 状态（0-禁用，1-启用） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 16,
    "name": "实木家具",
    "parent_id": 0,
    "sort_order": 10,
    "status": 1,
    "updated_at": "2026-05-05T10:30:00Z"
  }
}
```

---

### 5.4 删除品类

删除指定品类。

**请求**

```
DELETE /api/categories/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 品类 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 5.5 商品列表

获取商品分页列表。

**请求**

```
GET /api/products
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字（名称/编码） |
| category_id | int | 否 | 品类 ID 筛选 |
| status | int | 否 | 状态筛选（0-下架，1-上架） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "北欧简约布艺沙发",
        "code": "SF-001",
        "category_id": 2,
        "category_name": "布艺沙发",
        "main_image": "https://example.com/images/sf001.jpg",
        "price": 3999.00,
        "cost_price": 2000.00,
        "status": 1,
        "sku_count": 3,
        "total_stock": 50,
        "sales_count": 120,
        "created_at": "2026-01-15T00:00:00Z"
      }
    ],
    "total": 200,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 5.6 创建商品

创建新商品。

**请求**

```
POST /api/products
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 商品名称 |
| code | string | 是 | 商品编码，唯一 |
| category_id | int | 是 | 品类 ID |
| description | string | 否 | 商品描述 |
| main_image | string | 否 | 主图 URL |
| images | []string | 否 | 图片列表 |
| price | decimal | 是 | 销售价 |
| cost_price | decimal | 否 | 成本价 |
| commission_rate | decimal | 否 | 提成比例（0-1） |
| skus | []object | 否 | SKU 列表 |

**SKU 对象结构**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | SKU 名称（如颜色规格） |
| code | string | 是 | SKU 编码 |
| price | decimal | 否 | SKU 价格（不填则使用商品价格） |
| cost_price | decimal | 否 | SKU 成本价 |
| attributes | object | 否 | 规格属性（如颜色、尺寸） |

**请求示例**

```json
{
  "name": "北欧简约布艺沙发",
  "code": "SF-002",
  "category_id": 2,
  "description": "简约北欧风格布艺沙发，舒适透气",
  "main_image": "https://example.com/images/sf002.jpg",
  "price": 4299.00,
  "cost_price": 2100.00,
  "commission_rate": 0.05,
  "skus": [
    {
      "name": "灰色 三人位",
      "code": "SF-002-G-3",
      "price": 4299.00,
      "attributes": {"color": "灰色", "size": "三人位"}
    },
    {
      "name": "米色 三人位",
      "code": "SF-002-B-3",
      "price": 4299.00,
      "attributes": {"color": "米色", "size": "三人位"}
    }
  ]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 201,
    "name": "北欧简约布艺沙发",
    "code": "SF-002",
    "category_id": 2,
    "price": 4299.00,
    "cost_price": 2100.00,
    "commission_rate": 0.05,
    "status": 1,
    "skus": [
      {
        "id": 501,
        "name": "灰色 三人位",
        "code": "SF-002-G-3",
        "price": 4299.00
      },
      {
        "id": 502,
        "name": "米色 三人位",
        "code": "SF-002-B-3",
        "price": 4299.00
      }
    ],
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 5.7 更新商品

更新指定商品信息。

**请求**

```
PUT /api/products/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品 ID |

**请求参数 (Body)**

与创建商品相同，所有字段均为可选。

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 201,
    "name": "北欧简约布艺沙发（升级款）",
    "code": "SF-002",
    "price": 4599.00,
    "updated_at": "2026-05-05T10:30:00Z"
  }
}
```

---

### 5.8 删除商品

删除指定商品。

**请求**

```
DELETE /api/products/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 5.9 商品上下架

切换商品上下架状态。

**请求**

```
PUT /api/products/:id/status
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 是 | 状态（0-下架，1-上架） |

**请求示例**

```json
{
  "status": 0
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 201,
    "status": 0,
    "updated_at": "2026-05-05T11:00:00Z"
  }
}
```

---

### 5.10 SKU 管理

管理商品 SKU。

#### 获取 SKU 列表

```
GET /api/products/:id/skus
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 501,
        "product_id": 201,
        "name": "灰色 三人位",
        "code": "SF-002-G-3",
        "price": 4299.00,
        "cost_price": 2100.00,
        "stock": 20,
        "attributes": {"color": "灰色", "size": "三人位"}
      }
    ],
    "total": 2
  }
}
```

#### 创建 SKU

```
POST /api/products/:id/skus
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | SKU 名称 |
| code | string | 是 | SKU 编码，唯一 |
| price | decimal | 否 | SKU 价格 |
| cost_price | decimal | 否 | SKU 成本价 |
| attributes | object | 否 | 规格属性 |

#### 更新 SKU

```
PUT /api/products/:id/skus
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_id | int | 是 | SKU ID |
| name | string | 否 | SKU 名称 |
| price | decimal | 否 | SKU 价格 |
| cost_price | decimal | 否 | SKU 成本价 |
| attributes | object | 否 | 规格属性 |

#### 删除 SKU

```
DELETE /api/products/:id/skus
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_id | int | 是 | SKU ID |

---

## 6. 库存管理

### 6.1 库存查询

查询各仓库商品库存。

**请求**

```
GET /api/inventory/stocks
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| warehouse_id | int | 否 | 仓库 ID 筛选 |
| product_id | int | 否 | 商品 ID 筛选 |
| sku_code | string | 否 | SKU 编码筛选 |
| low_stock | int | 否 | 低库存筛选（1-仅显示低库存） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "warehouse_id": 1,
        "warehouse_name": "总仓",
        "product_id": 1,
        "product_name": "北欧简约布艺沙发",
        "sku_id": 501,
        "sku_name": "灰色 三人位",
        "sku_code": "SF-002-G-3",
        "quantity": 20,
        "locked_quantity": 3,
        "available_quantity": 17,
        "low_stock_threshold": 5,
        "is_low_stock": false,
        "updated_at": "2026-05-05T08:00:00Z"
      }
    ],
    "total": 500,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 6.2 库存流水

查询库存变动流水记录。

**请求**

```
GET /api/inventory/transactions
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| warehouse_id | int | 否 | 仓库 ID 筛选 |
| product_id | int | 否 | 商品 ID 筛选 |
| type | string | 否 | 类型筛选（in-入库，out-出库，transfer-调拨） |
| start_date | string | 否 | 开始日期（2026-01-01） |
| end_date | string | 否 | 结束日期（2026-05-05） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "warehouse_id": 1,
        "warehouse_name": "总仓",
        "product_id": 1,
        "product_name": "北欧简约布艺沙发",
        "sku_code": "SF-002-G-3",
        "type": "in",
        "type_name": "入库",
        "quantity": 50,
        "before_quantity": 0,
        "after_quantity": 50,
        "reference_type": "purchase",
        "reference_id": 1,
        "remark": "采购入库 PO-2026-001",
        "operator_name": "李四",
        "created_at": "2026-01-15T10:00:00Z"
      }
    ],
    "total": 1200,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 6.3 库存预警

获取库存预警列表。

**请求**

```
GET /api/inventory/alerts
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| status | int | 否 | 状态筛选（0-未处理，1-已处理） |
| warehouse_id | int | 否 | 仓库 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "warehouse_id": 1,
        "warehouse_name": "总仓",
        "product_id": 5,
        "product_name": "实木餐桌",
        "sku_code": "CZ-003-O",
        "current_quantity": 2,
        "low_stock_threshold": 5,
        "status": 0,
        "status_name": "未处理",
        "created_at": "2026-05-04T08:00:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 6.4 处理预警

处理指定库存预警。

**请求**

```
POST /api/inventory/alerts/:id/handle
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 预警 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 处理方式（create_purchase-创建采购单，ignore-忽略） |
| remark | string | 否 | 处理备注 |

**请求示例**

```json
{
  "action": "create_purchase",
  "remark": "已创建采购单 PO-2026-015"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 6.5 采购管理

#### 采购单列表

```
GET /api/purchases
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |
| status | string | 否 | 状态筛选（draft-草稿，pending-待审核，approved-已审核，received-已入库，rejected-已驳回） |
| supplier_id | int | 否 | 供应商 ID 筛选 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "PO-2026-001",
        "supplier_id": 1,
        "supplier_name": "优品家具供应商",
        "warehouse_id": 1,
        "warehouse_name": "总仓",
        "total_amount": 50000.00,
        "item_count": 5,
        "status": "approved",
        "status_name": "已审核",
        "created_by": "李四",
        "created_at": "2026-01-15T10:00:00Z"
      }
    ],
    "total": 30,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建采购单

```
POST /api/purchases
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| supplier_id | int | 是 | 供应商 ID |
| warehouse_id | int | 是 | 入库仓库 ID |
| expected_date | string | 否 | 预计到货日期 |
| remark | string | 否 | 备注 |
| items | []object | 是 | 采购明细 |

**采购明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品 ID |
| sku_id | int | 是 | SKU ID |
| quantity | int | 是 | 采购数量 |
| unit_price | decimal | 是 | 采购单价 |

**请求示例**

```json
{
  "supplier_id": 1,
  "warehouse_id": 1,
  "expected_date": "2026-05-15",
  "remark": "5月常规补货",
  "items": [
    {
      "product_id": 1,
      "sku_id": 501,
      "quantity": 20,
      "unit_price": 2000.00
    }
  ]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 15,
    "order_no": "PO-2026-015",
    "status": "draft",
    "total_amount": 40000.00,
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 6.6 审核采购单

审核通过或驳回采购单。

**请求**

```
POST /api/purchases/:id/approve
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 采购单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 审核操作（approve-通过，reject-驳回） |
| remark | string | 否 | 审核备注 |

**请求示例**

```json
{
  "action": "approve",
  "remark": "审核通过"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 15,
    "status": "approved",
    "approved_by": "王经理",
    "approved_at": "2026-05-05T11:00:00Z"
  }
}
```

---

### 6.7 确认入库

确认采购单入库。

**请求**

```
POST /api/purchases/:id/receipt
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 采购单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| items | []object | 否 | 实际入库明细（不传则按采购数量入库） |

**入库明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_id | int | 是 | SKU ID |
| actual_quantity | int | 是 | 实际入库数量 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 15,
    "status": "received",
    "received_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 6.8 调拨管理

#### 调拨单列表

```
GET /api/transfers
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |
| status | string | 否 | 状态筛选（draft-草稿，pending-待审核，approved-已审核，out-已出库，completed-已完成，rejected-已驳回） |
| from_warehouse_id | int | 否 | 调出仓库 ID |
| to_warehouse_id | int | 否 | 调入仓库 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "TR-2026-001",
        "from_warehouse_id": 1,
        "from_warehouse_name": "总仓",
        "to_warehouse_id": 2,
        "to_warehouse_name": "城南店仓库",
        "item_count": 3,
        "status": "completed",
        "status_name": "已完成",
        "created_by": "李四",
        "created_at": "2026-03-01T10:00:00Z"
      }
    ],
    "total": 20,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建调拨单

```
POST /api/transfers
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| from_warehouse_id | int | 是 | 调出仓库 ID |
| to_warehouse_id | int | 是 | 调入仓库 ID |
| remark | string | 否 | 备注 |
| items | []object | 是 | 调拨明细 |

**调拨明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品 ID |
| sku_id | int | 是 | SKU ID |
| quantity | int | 是 | 调拨数量 |

**请求示例**

```json
{
  "from_warehouse_id": 1,
  "to_warehouse_id": 2,
  "remark": "城南店补货",
  "items": [
    {
      "product_id": 1,
      "sku_id": 501,
      "quantity": 5
    }
  ]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "order_no": "TR-2026-010",
    "status": "draft",
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 6.9 审核调拨单

**请求**

```
POST /api/transfers/:id/approve
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 调拨单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 审核操作（approve-通过，reject-驳回） |
| remark | string | 否 | 审核备注 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "status": "approved",
    "approved_at": "2026-05-05T11:00:00Z"
  }
}
```

---

### 6.10 确认出库

确认调拨单出库。

**请求**

```
POST /api/transfers/:id/out
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 调拨单 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "status": "out",
    "out_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 6.11 确认入库

确认调拨单入库（调入方确认）。

**请求**

```
POST /api/transfers/:id/in
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 调拨单 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "status": "completed",
    "in_at": "2026-05-06T09:00:00Z"
  }
}
```

---

### 6.12 仓库管理

#### 仓库列表

```
GET /api/warehouses
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "总仓",
        "code": "WH-001",
        "address": "市工业园区A栋",
        "manager_name": "李四",
        "phone": "13800138000",
        "status": 1,
        "product_count": 200,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 5
  }
}
```

#### 创建仓库

```
POST /api/warehouses
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 仓库名称 |
| code | string | 是 | 仓库编码，唯一 |
| address | string | 否 | 仓库地址 |
| manager_id | int | 否 | 负责人 ID |
| phone | string | 否 | 联系电话 |

#### 更新仓库

```
PUT /api/warehouses/:id
```

#### 删除仓库

```
DELETE /api/warehouses/:id
```

---

### 6.13 礼品管理

#### 礼品列表

```
GET /api/gifts
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |
| status | int | 否 | 状态筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "定制抱枕",
        "code": "GF-001",
        "cost_price": 50.00,
        "total_quantity": 500,
        "used_quantity": 120,
        "available_quantity": 380,
        "status": 1,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建礼品

```
POST /api/gifts
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 礼品名称 |
| code | string | 是 | 礼品编码，唯一 |
| cost_price | decimal | 是 | 成本价 |
| total_quantity | int | 是 | 总数量 |
| description | string | 否 | 礼品描述 |
| image | string | 否 | 礼品图片 |

#### 更新礼品

```
PUT /api/gifts/:id
```

#### 删除礼品

```
DELETE /api/gifts/:id
```

---

### 6.14 供应商管理

#### 供应商列表

```
GET /api/suppliers
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |
| status | int | 否 | 状态筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "优品家具供应商",
        "code": "SP-001",
        "contact_person": "赵六",
        "phone": "13700137000",
        "address": "省家具产业园B区",
        "product_count": 50,
        "purchase_count": 30,
        "status": 1,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 12,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建供应商

```
POST /api/suppliers
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 供应商名称 |
| code | string | 是 | 供应商编码，唯一 |
| contact_person | string | 否 | 联系人 |
| phone | string | 否 | 联系电话 |
| email | string | 否 | 邮箱 |
| address | string | 否 | 地址 |
| bank_name | string | 否 | 银行名称 |
| bank_account | string | 否 | 银行账号 |
| remark | string | 否 | 备注 |

#### 更新供应商

```
PUT /api/suppliers/:id
```

#### 删除供应商

```
DELETE /api/suppliers/:id
```

---

## 7. 订单管理

### 7.1 订单列表

获取订单分页列表。

**请求**

```
GET /api/orders
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字（订单号/客户名） |
| status | string | 否 | 状态筛选（draft-草稿，pending-待审核，approved-已审核，delivering-配送中，completed-已完成，cancelled-已取消，returned-已退货） |
| store_id | int | 否 | 门店 ID 筛选 |
| sales_id | int | 否 | 销售员 ID 筛选 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "ORD-2026-00001",
        "customer_id": 1,
        "customer_name": "张先生",
        "customer_phone": "13500135000",
        "store_id": 1,
        "store_name": "总店",
        "sales_id": 5,
        "sales_name": "王五",
        "total_amount": 12999.00,
        "paid_amount": 8000.00,
        "unpaid_amount": 4999.00,
        "item_count": 3,
        "status": "delivering",
        "status_name": "配送中",
        "created_at": "2026-04-20T10:00:00Z"
      }
    ],
    "total": 300,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 7.2 创建订单

创建新订单。

**请求**

```
POST /api/orders
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| customer_id | int | 是 | 客户 ID |
| store_id | int | 是 | 门店 ID |
| delivery_address | string | 否 | 配送地址 |
| delivery_date | string | 否 | 预计配送日期 |
| remark | string | 否 | 订单备注 |
| items | []object | 是 | 订单明细 |
| gifts | []object | 否 | 赠品明细 |

**订单明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品 ID |
| sku_id | int | 是 | SKU ID |
| quantity | int | 是 | 数量 |
| unit_price | decimal | 是 | 单价 |
| discount | decimal | 否 | 折扣（0-1） |

**赠品明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| gift_id | int | 是 | 礼品 ID |
| quantity | int | 是 | 数量 |

**请求示例**

```json
{
  "customer_id": 1,
  "store_id": 1,
  "delivery_address": "市南区XX小区3栋502",
  "delivery_date": "2026-05-10",
  "remark": "客户要求周末配送",
  "items": [
    {
      "product_id": 1,
      "sku_id": 501,
      "quantity": 1,
      "unit_price": 4299.00,
      "discount": 1.0
    },
    {
      "product_id": 10,
      "sku_id": 601,
      "quantity": 2,
      "unit_price": 1999.00,
      "discount": 0.95
    }
  ],
  "gifts": [
    {
      "gift_id": 1,
      "quantity": 2
    }
  ]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 100,
    "order_no": "ORD-2026-00100",
    "total_amount": 8197.10,
    "status": "draft",
    "items": [
      {
        "id": 301,
        "product_name": "北欧简约布艺沙发",
        "sku_name": "灰色 三人位",
        "quantity": 1,
        "unit_price": 4299.00,
        "subtotal": 4299.00
      }
    ],
    "created_at": "2026-05-05T10:00:00Z"
  }
}
```

---

### 7.3 订单详情

获取订单详细信息。

**请求**

```
GET /api/orders/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 订单 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 100,
    "order_no": "ORD-2026-00100",
    "customer_id": 1,
    "customer_name": "张先生",
    "customer_phone": "13500135000",
    "store_id": 1,
    "store_name": "总店",
    "sales_id": 5,
    "sales_name": "王五",
    "delivery_address": "市南区XX小区3栋502",
    "delivery_date": "2026-05-10",
    "total_amount": 8197.10,
    "paid_amount": 4000.00,
    "unpaid_amount": 4197.10,
    "status": "approved",
    "status_name": "已审核",
    "remark": "客户要求周末配送",
    "items": [
      {
        "id": 301,
        "product_id": 1,
        "product_name": "北欧简约布艺沙发",
        "sku_id": 501,
        "sku_name": "灰色 三人位",
        "sku_code": "SF-002-G-3",
        "quantity": 1,
        "unit_price": 4299.00,
        "discount": 1.0,
        "subtotal": 4299.00
      }
    ],
    "gifts": [
      {
        "id": 1,
        "gift_name": "定制抱枕",
        "quantity": 2
      }
    ],
    "payments": [
      {
        "id": 1,
        "amount": 4000.00,
        "payment_method": "bank_transfer",
        "payment_date": "2026-05-05",
        "status": "approved"
      }
    ],
    "created_at": "2026-05-05T10:00:00Z",
    "updated_at": "2026-05-05T11:00:00Z"
  }
}
```

---

### 7.4 审核订单

审核通过或驳回订单。

**请求**

```
POST /api/orders/:id/approve
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 订单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 审核操作（approve-通过，reject-驳回） |
| remark | string | 否 | 审核备注 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 100,
    "status": "approved",
    "approved_by": "王经理",
    "approved_at": "2026-05-05T11:00:00Z"
  }
}
```

---

### 7.5 取消订单

取消订单。

**请求**

```
POST /api/orders/:id/cancel
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 订单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reason | string | 是 | 取消原因 |

**请求示例**

```json
{
  "reason": "客户要求取消"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 100,
    "status": "cancelled",
    "cancelled_at": "2026-05-05T12:00:00Z"
  }
}
```

---

### 7.6 退货处理

处理订单退货。

**请求**

```
POST /api/orders/:id/return
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 订单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reason | string | 是 | 退货原因 |
| items | []object | 是 | 退货明细 |

**退货明细对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_item_id | int | 是 | 订单明细 ID |
| quantity | int | 是 | 退货数量 |
| refund_amount | decimal | 是 | 退款金额 |

**请求示例**

```json
{
  "reason": "商品质量问题",
  "items": [
    {
      "order_item_id": 301,
      "quantity": 1,
      "refund_amount": 4299.00
    }
  ]
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "return_no": "RT-2026-00001",
    "total_refund": 4299.00,
    "status": "pending",
    "created_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 7.7 订单动态

获取订单状态变更动态。

**请求**

```
GET /api/orders/feed
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| order_id | int | 否 | 订单 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_id": 100,
        "order_no": "ORD-2026-00100",
        "action": "status_change",
        "description": "订单状态从「草稿」变更为「已审核」",
        "operator_name": "王经理",
        "created_at": "2026-05-05T11:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 7.8 录入回款

录入订单回款记录。

**请求**

```
POST /api/payments
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int | 是 | 订单 ID |
| amount | decimal | 是 | 回款金额 |
| payment_method | string | 是 | 回款方式（cash-现金，bank_transfer-银行转账，wechat-微信，alipay-支付宝） |
| payment_date | string | 是 | 回款日期 |
| reference_no | string | 否 | 参考号/流水号 |
| remark | string | 否 | 备注 |

**请求示例**

```json
{
  "order_id": 100,
  "amount": 4000.00,
  "payment_method": "bank_transfer",
  "payment_date": "2026-05-05",
  "reference_no": "TXN20260505001",
  "remark": "首付款"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 50,
    "order_id": 100,
    "order_no": "ORD-2026-00100",
    "amount": 4000.00,
    "payment_method": "bank_transfer",
    "payment_method_name": "银行转账",
    "payment_date": "2026-05-05",
    "status": "pending",
    "created_at": "2026-05-05T10:30:00Z"
  }
}
```

---

### 7.9 回款列表

获取回款记录列表。

**请求**

```
GET /api/payments
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| order_id | int | 否 | 订单 ID 筛选 |
| status | string | 否 | 状态筛选（pending-待审核，approved-已审核，rejected-已驳回） |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 50,
        "order_id": 100,
        "order_no": "ORD-2026-00100",
        "customer_name": "张先生",
        "amount": 4000.00,
        "payment_method": "bank_transfer",
        "payment_method_name": "银行转账",
        "payment_date": "2026-05-05",
        "status": "approved",
        "status_name": "已审核",
        "created_by": "王五",
        "created_at": "2026-05-05T10:30:00Z"
      }
    ],
    "total": 200,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 7.10 审核回款

审核回款记录。

**请求**

```
POST /api/payments/:id/approve
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 回款 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 审核操作（approve-通过，reject-驳回） |
| remark | string | 否 | 审核备注 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 50,
    "status": "approved",
    "approved_at": "2026-05-05T12:00:00Z"
  }
}
```

---

### 7.11 客户管理

#### 客户列表

```
GET /api/customers
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字（姓名/手机号） |
| source | string | 否 | 来源筛选 |
| store_id | int | 否 | 门店 ID 筛选 |
| sales_id | int | 否 | 销售员 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "张先生",
        "phone": "13500135000",
        "gender": "male",
        "source": "walk_in",
        "source_name": "自然进店",
        "store_id": 1,
        "store_name": "总店",
        "sales_id": 5,
        "sales_name": "王五",
        "order_count": 3,
        "total_amount": 25000.00,
        "last_order_date": "2026-04-20",
        "status": 1,
        "created_at": "2026-02-10T10:00:00Z"
      }
    ],
    "total": 500,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建客户

```
POST /api/customers
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 客户姓名 |
| phone | string | 是 | 手机号 |
| gender | string | 否 | 性别（male/female） |
| birthday | string | 否 | 生日 |
| address | string | 否 | 地址 |
| source | string | 否 | 来源（walk_in-进店，referral-转介绍，online-线上，other-其他） |
| store_id | int | 否 | 所属门店 ID |
| sales_id | int | 否 | 跟进销售员 ID |
| remark | string | 否 | 备注 |

#### 更新客户

```
PUT /api/customers/:id
```

#### 删除客户

```
DELETE /api/customers/:id
```

---

### 7.12 跟进记录

#### 创建跟进记录

```
POST /api/customers/:id/follow-ups
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 客户 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 跟进类型（phone-电话，visit-上门，wechat-微信，other-其他） |
| content | string | 是 | 跟进内容 |
| next_follow_up_date | string | 否 | 下次跟进日期 |
| result | string | 否 | 跟进结果（interested-有意向，pending-待考虑，rejected-无意向，deal-成交） |

**请求示例**

```json
{
  "type": "phone",
  "content": "电话沟通，客户对沙发款式满意，考虑周末到店看样",
  "next_follow_up_date": "2026-05-08",
  "result": "interested"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "customer_id": 1,
    "type": "phone",
    "type_name": "电话跟进",
    "content": "电话沟通，客户对沙发款式满意，考虑周末到店看样",
    "result": "interested",
    "result_name": "有意向",
    "next_follow_up_date": "2026-05-08",
    "created_by": "王五",
    "created_at": "2026-05-05T15:00:00Z"
  }
}
```

#### 跟进记录列表

```
GET /api/customers/:id/follow-ups
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "customer_id": 1,
        "type": "phone",
        "type_name": "电话跟进",
        "content": "电话沟通，客户对沙发款式满意",
        "result": "interested",
        "result_name": "有意向",
        "next_follow_up_date": "2026-05-08",
        "created_by": "王五",
        "created_at": "2026-05-05T15:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 7.13 同行管理

#### 同行列表

```
GET /api/peers
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| keyword | string | 否 | 搜索关键字 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "XX家具城",
        "contact_person": "刘总",
        "phone": "13600136000",
        "address": "市北区XX路88号",
        "main_products": "沙发、餐桌、床垫",
        "remark": "主要竞争对手",
        "status": 1,
        "created_at": "2026-01-15T00:00:00Z"
      }
    ],
    "total": 20,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建同行

```
POST /api/peers
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 同行名称 |
| contact_person | string | 否 | 联系人 |
| phone | string | 否 | 联系电话 |
| address | string | 否 | 地址 |
| main_products | string | 否 | 主营产品 |
| remark | string | 否 | 备注 |

#### 更新同行

```
PUT /api/peers/:id
```

#### 删除同行

```
DELETE /api/peers/:id
```

---

## 8. 提成管理

### 8.1 提成列表

获取提成记录列表。

**请求**

```
GET /api/commissions
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| user_id | int | 否 | 员工 ID 筛选 |
| order_id | int | 否 | 订单 ID 筛选 |
| period | string | 否 | 薪资周期（2026-04） |
| type | string | 否 | 类型筛选（order-订单提成，fund_pool-基金池，referral-老带新） |
| status | string | 否 | 状态筛选（pending-待确认，confirmed-已确认，paid-已发放） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 5,
        "user_name": "王五",
        "order_id": 100,
        "order_no": "ORD-2026-00100",
        "type": "order",
        "type_name": "订单提成",
        "amount": 409.86,
        "rate": 0.05,
        "base_amount": 8197.10,
        "period": "2026-05",
        "status": "pending",
        "status_name": "待确认",
        "created_at": "2026-05-05T11:00:00Z"
      }
    ],
    "total": 500,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 8.2 提成汇总

获取提成汇总数据。

**请求**

```
GET /api/commissions/summary
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 薪资周期（2026-04） |
| store_id | int | 否 | 门店 ID 筛选 |
| user_id | int | 否 | 员工 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "period": "2026-05",
    "total_commission": 150000.00,
    "order_commission": 120000.00,
    "fund_pool_commission": 20000.00,
    "referral_commission": 10000.00,
    "user_summary": [
      {
        "user_id": 5,
        "user_name": "王五",
        "store_name": "总店",
        "order_commission": 8000.00,
        "fund_pool_commission": 1500.00,
        "referral_commission": 500.00,
        "total_commission": 10000.00,
        "order_count": 15
      }
    ]
  }
}
```

---

### 8.3 手工调整提成

对指定提成记录进行手工调整。

**请求**

```
POST /api/commissions/:id/adjust
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 提成记录 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | decimal | 是 | 调整后金额 |
| reason | string | 是 | 调整原因 |

**请求示例**

```json
{
  "amount": 500.00,
  "reason": "客户为大客户，额外奖励"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "original_amount": 409.86,
    "adjusted_amount": 500.00,
    "adjust_reason": "客户为大客户，额外奖励",
    "adjusted_by": "王经理",
    "adjusted_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 8.4 触发提成计算

手动触发提成计算。

**请求**

```
POST /api/commissions/calculate
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 是 | 薪资周期（2026-04） |
| user_ids | []int | 否 | 指定员工 ID 列表（不传则计算全部） |

**请求示例**

```json
{
  "period": "2026-05"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task-001",
    "period": "2026-05",
    "calculated_count": 50,
    "total_commission": 150000.00,
    "started_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 8.5 基金池管理

#### 基金池列表

```
GET /api/fund-pools
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| status | string | 否 | 状态筛选（active-进行中，settled-已结算） |
| period | string | 否 | 周期筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "2026年5月销售基金池",
        "period": "2026-05",
        "total_amount": 50000.00,
        "distributed_amount": 0,
        "participant_count": 15,
        "status": "active",
        "status_name": "进行中",
        "created_at": "2026-05-01T00:00:00Z"
      }
    ],
    "total": 6,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建基金池

```
POST /api/fund-pools
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 基金池名称 |
| period | string | 是 | 周期（2026-05） |
| total_amount | decimal | 是 | 基金池总额 |
| distribution_rule | string | 是 | 分配规则（equal-平均分配，proportional-按业绩比例） |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| remark | string | 否 | 备注 |

**请求示例**

```json
{
  "name": "2026年5月销售基金池",
  "period": "2026-05",
  "total_amount": 50000.00,
  "distribution_rule": "proportional",
  "start_date": "2026-05-01",
  "end_date": "2026-05-31"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 7,
    "name": "2026年5月销售基金池",
    "period": "2026-05",
    "total_amount": 50000.00,
    "status": "active",
    "created_at": "2026-05-05T14:00:00Z"
  }
}
```

---

### 8.6 基金池份额

获取指定基金池的参与人份额明细。

**请求**

```
GET /api/fund-pools/:id/shares
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 基金池 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "fund_pool_id": 1,
    "fund_pool_name": "2026年5月销售基金池",
    "total_amount": 50000.00,
    "shares": [
      {
        "user_id": 5,
        "user_name": "王五",
        "store_name": "总店",
        "sales_amount": 200000.00,
        "sales_proportion": 0.25,
        "share_amount": 12500.00,
        "status": "pending"
      }
    ]
  }
}
```

---

### 8.7 结算基金池

结算指定基金池，将份额分配给参与人。

**请求**

```
POST /api/fund-pools/settle
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fund_pool_id | int | 是 | 基金池 ID |
| shares | []object | 否 | 自定义份额（不传则按规则自动计算） |

**自定义份额对象**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | int | 是 | 员工 ID |
| share_amount | decimal | 是 | 分配金额 |

**请求示例**

```json
{
  "fund_pool_id": 1
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "fund_pool_id": 1,
    "status": "settled",
    "settled_at": "2026-05-05T16:00:00Z",
    "total_distributed": 50000.00,
    "participant_count": 15
  }
}
```

---

### 8.8 老带新管理

#### 老带新记录列表

```
GET /api/referrals
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| referrer_id | int | 否 | 推荐人 ID 筛选 |
| referee_id | int | 否 | 被推荐人 ID 筛选 |
| period | string | 否 | 周期筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "referrer_id": 5,
        "referrer_name": "王五",
        "referee_id": 8,
        "referee_name": "赵九",
        "order_id": 100,
        "order_no": "ORD-2026-00100",
        "order_amount": 8197.10,
        "commission_rate": 0.02,
        "commission_amount": 163.94,
        "period": "2026-05",
        "status": "pending",
        "created_at": "2026-05-05T11:00:00Z"
      }
    ],
    "total": 30,
    "page": 1,
    "page_size": 20
  }
}
```

#### 创建老带新记录

```
POST /api/referrals
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| referrer_id | int | 是 | 推荐人（老员工） ID |
| referee_id | int | 是 | 被推荐人（新员工） ID |
| order_id | int | 是 | 关联订单 ID |
| commission_rate | decimal | 否 | 提成比例（不传则使用系统配置） |
| remark | string | 否 | 备注 |

**请求示例**

```json
{
  "referrer_id": 5,
  "referee_id": 8,
  "order_id": 100,
  "remark": "王五带新员工赵九成交"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "referrer_id": 5,
    "referrer_name": "王五",
    "referee_id": 8,
    "referee_name": "赵九",
    "commission_amount": 163.94,
    "status": "pending",
    "created_at": "2026-05-05T11:00:00Z"
  }
}
```

---

## 9. 工资管理

### 9.1 工资列表

获取工资单列表。

**请求**

```
GET /api/salaries
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| period | string | 否 | 薪资周期（2026-04） |
| user_id | int | 否 | 员工 ID 筛选 |
| status | string | 否 | 状态筛选（draft-草稿，confirmed-已确认，paid-已发放） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 5,
        "user_name": "王五",
        "store_name": "总店",
        "period": "2026-05",
        "base_salary": 5000.00,
        "commission": 10000.00,
        "bonus": 0,
        "deduction": 0,
        "total_salary": 15000.00,
        "status": "draft",
        "status_name": "草稿",
        "created_at": "2026-06-01T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 9.2 工资详情

获取指定工资单详细信息。

**请求**

```
GET /api/salaries/:id
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 工资单 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "user_id": 5,
    "user_name": "王五",
    "store_name": "总店",
    "period": "2026-05",
    "base_salary": 5000.00,
    "commission_detail": {
      "order_commission": 8000.00,
      "fund_pool_commission": 1500.00,
      "referral_commission": 500.00,
      "total_commission": 10000.00
    },
    "bonus": 0,
    "deduction": 0,
    "total_salary": 15000.00,
    "status": "draft",
    "status_name": "草稿",
    "created_at": "2026-06-01T00:00:00Z"
  }
}
```

---

### 9.3 生成工资

按周期生成工资单。

**请求**

```
POST /api/salaries/generate
```

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 是 | 薪资周期（2026-04） |
| user_ids | []int | 否 | 指定员工 ID 列表（不传则生成全部） |

**请求示例**

```json
{
  "period": "2026-05"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "period": "2026-05",
    "generated_count": 50,
    "total_salary": 750000.00,
    "generated_at": "2026-06-01T00:00:00Z"
  }
}
```

---

### 9.4 确认工资

确认指定工资单。

**请求**

```
POST /api/salaries/:id/confirm
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 工资单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| base_salary | decimal | 否 | 调整基本工资 |
| bonus | decimal | 否 | 奖金 |
| deduction | decimal | 否 | 扣款 |
| remark | string | 否 | 备注 |

**请求示例**

```json
{
  "bonus": 500,
  "remark": "5月优秀员工奖励"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "status": "confirmed",
    "total_salary": 15500.00,
    "confirmed_by": "王经理",
    "confirmed_at": "2026-06-02T10:00:00Z"
  }
}
```

---

### 9.5 发放工资

确认发放工资。

**请求**

```
POST /api/salaries/:id/pay
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 工资单 ID |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| payment_method | string | 否 | 发放方式（bank_transfer-银行转账，cash-现金） |
| remark | string | 否 | 备注 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "status": "paid",
    "paid_by": "财务李姐",
    "paid_at": "2026-06-05T10:00:00Z"
  }
}
```

---

## 10. 数据分析

### 10.1 仪表盘概览

获取系统仪表盘概览数据。

**请求**

```
GET /api/dashboard/overview
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| store_id | int | 否 | 门店 ID 筛选 |
| period | string | 否 | 周期（today-今天，week-本周，month-本月，year-本年） |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "period": "month",
    "sales_amount": 2500000.00,
    "sales_amount_change": 0.15,
    "order_count": 300,
    "order_count_change": 0.08,
    "payment_amount": 2000000.00,
    "payment_rate": 0.80,
    "customer_count": 150,
    "customer_count_change": 0.12,
    "inventory_alert_count": 15,
    "pending_order_count": 8,
    "pending_payment_count": 12,
    "top_sales": [
      {
        "user_name": "王五",
        "sales_amount": 200000.00,
        "order_count": 15
      }
    ],
    "recent_orders": [
      {
        "id": 100,
        "order_no": "ORD-2026-00100",
        "customer_name": "张先生",
        "total_amount": 8197.10,
        "status": "approved",
        "created_at": "2026-05-05T10:00:00Z"
      }
    ]
  }
}
```

---

### 10.2 销售总览

获取销售数据总览。

**请求**

```
GET /api/reports/sales/summary
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| store_id | int | 否 | 门店 ID 筛选 |
| category_id | int | 否 | 品类 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_sales_amount": 2500000.00,
    "total_order_count": 300,
    "average_order_amount": 8333.33,
    "total_cost": 1250000.00,
    "gross_profit": 1250000.00,
    "gross_profit_rate": 0.50,
    "by_category": [
      {
        "category_name": "沙发",
        "sales_amount": 1000000.00,
        "order_count": 120,
        "proportion": 0.40
      }
    ],
    "by_store": [
      {
        "store_name": "总店",
        "sales_amount": 1500000.00,
        "order_count": 180,
        "proportion": 0.60
      }
    ]
  }
}
```

---

### 10.3 销售趋势

获取销售趋势数据。

**请求**

```
GET /api/reports/sales/trend
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| granularity | string | 否 | 粒度（day-日，week-周，month-月），默认 day |
| store_id | int | 否 | 门店 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "trend": [
      {
        "date": "2026-05-01",
        "sales_amount": 80000.00,
        "order_count": 10,
        "payment_amount": 60000.00
      },
      {
        "date": "2026-05-02",
        "sales_amount": 95000.00,
        "order_count": 12,
        "payment_amount": 75000.00
      }
    ]
  }
}
```

---

### 10.4 业绩排行

获取员工业绩排行榜。

**请求**

```
GET /api/reports/sales/ranking
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 周期（2026-05） |
| store_id | int | 否 | 门店 ID 筛选 |
| limit | int | 否 | 排行数量，默认 20 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ranking": [
      {
        "rank": 1,
        "user_id": 5,
        "user_name": "王五",
        "store_name": "总店",
        "sales_amount": 200000.00,
        "order_count": 15,
        "commission": 10000.00,
        "avg_order_amount": 13333.33
      },
      {
        "rank": 2,
        "user_id": 6,
        "user_name": "李六",
        "store_name": "城南店",
        "sales_amount": 180000.00,
        "order_count": 12,
        "commission": 9000.00,
        "avg_order_amount": 15000.00
      }
    ]
  }
}
```

---

### 10.5 利润分析

获取利润分析数据。

**请求**

```
GET /api/reports/profit/analysis
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| store_id | int | 否 | 门店 ID 筛选 |
| category_id | int | 否 | 品类 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_revenue": 2500000.00,
    "total_cost": 1250000.00,
    "gross_profit": 1250000.00,
    "gross_profit_rate": 0.50,
    "by_category": [
      {
        "category_name": "沙发",
        "revenue": 1000000.00,
        "cost": 500000.00,
        "gross_profit": 500000.00,
        "gross_profit_rate": 0.50
      }
    ],
    "by_store": [
      {
        "store_name": "总店",
        "revenue": 1500000.00,
        "cost": 750000.00,
        "gross_profit": 750000.00,
        "gross_profit_rate": 0.50
      }
    ],
    "by_month": [
      {
        "month": "2026-05",
        "revenue": 500000.00,
        "cost": 250000.00,
        "gross_profit": 250000.00,
        "gross_profit_rate": 0.50
      }
    ]
  }
}
```

---

### 10.6 回款分析

获取回款分析数据。

**请求**

```
GET /api/reports/payment/analysis
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| store_id | int | 否 | 门店 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_order_amount": 2500000.00,
    "total_payment_amount": 2000000.00,
    "total_unpaid_amount": 500000.00,
    "payment_rate": 0.80,
    "overdue_amount": 100000.00,
    "overdue_count": 15,
    "by_method": [
      {
        "method": "bank_transfer",
        "method_name": "银行转账",
        "amount": 1200000.00,
        "proportion": 0.60
      },
      {
        "method": "wechat",
        "method_name": "微信",
        "amount": 500000.00,
        "proportion": 0.25
      }
    ],
    "by_sales": [
      {
        "user_name": "王五",
        "order_amount": 200000.00,
        "payment_amount": 180000.00,
        "unpaid_amount": 20000.00,
        "payment_rate": 0.90
      }
    ]
  }
}
```

---

### 10.7 库存分析

获取库存分析数据。

**请求**

```
GET /api/reports/inventory/analysis
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| warehouse_id | int | 否 | 仓库 ID 筛选 |
| category_id | int | 否 | 品类 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_sku_count": 500,
    "total_stock_value": 5000000.00,
    "low_stock_count": 15,
    "out_of_stock_count": 3,
    "over_stock_count": 20,
    "by_warehouse": [
      {
        "warehouse_name": "总仓",
        "sku_count": 300,
        "stock_value": 3000000.00,
        "low_stock_count": 8
      }
    ],
    "by_category": [
      {
        "category_name": "沙发",
        "sku_count": 50,
        "stock_value": 1000000.00,
        "low_stock_count": 3
      }
    ],
    "stock_turnover": [
      {
        "category_name": "沙发",
        "turnover_rate": 2.5,
        "avg_stock_days": 45
      }
    ]
  }
}
```

---

### 10.8 提成分析

获取提成分析数据。

**请求**

```
GET /api/reports/commission/analysis
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |
| store_id | int | 否 | 门店 ID 筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_commission": 150000.00,
    "order_commission": 120000.00,
    "fund_pool_commission": 20000.00,
    "referral_commission": 10000.00,
    "commission_rate": 0.06,
    "by_store": [
      {
        "store_name": "总店",
        "total_commission": 80000.00,
        "sales_amount": 1500000.00,
        "commission_rate": 0.053
      }
    ],
    "by_type": [
      {
        "type": "order",
        "type_name": "订单提成",
        "amount": 120000.00,
        "proportion": 0.80
      }
    ],
    "top_earners": [
      {
        "user_name": "王五",
        "total_commission": 10000.00,
        "order_commission": 8000.00,
        "fund_pool_commission": 1500.00,
        "referral_commission": 500.00
      }
    ]
  }
}
```

---

## 11. 系统配置

### 11.1 配置列表

获取系统配置列表。

**请求**

```
GET /api/configs
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group | string | 否 | 配置分组筛选 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "key": "commission.order_rate",
        "value": "0.05",
        "type": "decimal",
        "group": "commission",
        "description": "订单提成默认比例"
      },
      {
        "key": "commission.referral_rate",
        "value": "0.02",
        "type": "decimal",
        "group": "commission",
        "description": "老带新提成默认比例"
      },
      {
        "key": "salary.base_amount",
        "value": "5000",
        "type": "int",
        "group": "salary",
        "description": "默认基本工资"
      }
    ],
    "total": 25
  }
}
```

---

### 11.2 获取配置

获取指定配置项的值。

**请求**

```
GET /api/configs/:key
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | 配置键名 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "key": "commission.order_rate",
    "value": "0.05",
    "type": "decimal",
    "group": "commission",
    "description": "订单提成默认比例"
  }
}
```

---

### 11.3 更新配置

更新指定配置项的值。

**请求**

```
PUT /api/configs/:key
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | 配置键名 |

**请求参数 (Body)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| value | string | 是 | 配置值 |

**请求示例**

```json
{
  "value": "0.06"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "key": "commission.order_rate",
    "value": "0.06",
    "updated_at": "2026-05-05T16:00:00Z"
  }
}
```

---

### 11.4 门店列表

获取门店列表。

**请求**

```
GET /api/stores
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "总店",
        "code": "STORE-001",
        "address": "市中心XX路100号",
        "manager_name": "王经理",
        "phone": "13800138000",
        "user_count": 20,
        "status": 1,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 5
  }
}
```

---

### 11.5 操作日志

获取系统操作日志。

**请求**

```
GET /api/operation-logs
```

**请求参数 (Query)**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |
| user_id | int | 否 | 操作人 ID 筛选 |
| module | string | 否 | 模块筛选 |
| action | string | 否 | 操作类型筛选 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "user_name": "admin",
        "module": "用户管理",
        "action": "create",
        "action_name": "创建用户",
        "target_type": "user",
        "target_id": 10,
        "description": "创建用户「张三」",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2026-05-05T10:00:00Z"
      }
    ],
    "total": 5000,
    "page": 1,
    "page_size": 20
  }
}
```

---

> 文档版本：v1.0.0 | 更新日期：2026-05-05
