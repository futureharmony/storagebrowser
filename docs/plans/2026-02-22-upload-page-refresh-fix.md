# 大文件上传防刷新方案设计

> **Goal:** 解决大文件上传时因 token 过期或网络不稳定导致页面刷新的问题

> **Architecture:** 在上传开始前预刷新 token，上传过程中使用独立的认证容错机制

> **Tech Stack:** Vue.js 3, Pinia, TypeScript

---

## 背景

### 问题现象
- 上传大文件（超过几分钟）时，页面突然刷新
- 上传失败，需要重新选择文件上传

### 根本原因
1. JWT token 有有效期（默认 2 小时）
2. 大文件上传时间可能超过 token 有效期
3. 上传期间 token 过期，请求返回 401
4. 401 触发 logout()
5. 在 noAuth 模式下，logout() 调用 `window.location.reload()`
6. 页面刷新导致上传失败

---

## 方案设计

### 方案 1: 上传前预刷新 Token（推荐）

**核心思路**：在开始上传前检查并刷新 token，确保上传过程有足够的 token 有效期

**实现**：
1. 在 `upload.ts` 的 `upload()` 函数中，上传开始前先刷新 token
2. 如果刷新失败，提示用户重新登录后再上传
3. 如果刷新成功，开始上传

**优点**：
- 从源头解决问题
- 用户体验好（上传前解决，不中断）

**缺点**：
- 需要修改上传流程

### 方案 2: 上传期间认证容错

**核心思路**：上传期间的 401 错误不触发 logout，而是尝试恢复

**实现**：
1. 创建专门的 upload API 函数，使用独立的错误处理
2. 上传期间的 401 错误：
   - 先尝试刷新 token
   - 刷新成功后重试上传
   - 如果仍然失败，显示错误但不 logout
3. 只有非上传请求的 401 才触发 logout

**优点**：
- 更健壮，能自动恢复
- 不影响其他功能

**缺点**：
- 实现复杂度稍高

---

## 推荐方案：方案 1 + 方案 2 组合

### 实现步骤

#### Step 1: 在 upload store 添加 token 刷新检查

**修改文件**: `frontend/src/stores/upload.ts`

```typescript
// 在 upload 函数开始时调用
const ensureTokenValid = async () => {
  const authStore = useAuthStore();
  if (!authStore.jwt) {
    throw new Error("Not authenticated");
  }
  
  // 检查 token 是否即将过期（剩余时间少于 5 分钟）
  const tokenData = jwtDecode<JwtPayload>(authStore.jwt);
  const expiresAt = new Date(tokenData.exp! * 1000);
  const fiveMinutes = 5 * 60 * 1000;
  
  if (expiresAt.getTime() - Date.now() < fiveMinutes) {
    // Token 即将过期，先刷新
    await renew(authStore.jwt);
  }
};
```

#### Step 2: 修改上传流程，在上传开始前刷新 token

**修改文件**: `frontend/src/stores/upload.ts`

```typescript
const upload = async (
  path: string,
  name: string,
  file: File | null,
  overwrite: boolean,
  type: ResourceType
) => {
  // 上传开始前确保 token 有效
  if (type === "file" && file) {
    try {
      await ensureTokenValid();
    } catch (error) {
      $showError("Please refresh the page and login again before uploading");
      return;
    }
  }
  
  // ... 现有逻辑
};
```

#### Step 3: 创建独立的 upload API 函数

**新建文件**: `frontend/src/api/upload.ts`

```typescript
import { fetchURL } from "./utils";
import { baseURL } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
import { removePrefix } from "./utils";

export async function uploadWithRetry(
  url: string,
  content: Blob,
  overwrite: boolean,
  onprogress: (loaded: number) => void,
  scope?: string
): Promise<string> {
  const authStore = useAuthStore();
  let lastError: Error | null = null;
  
  // 最多重试 2 次（1 次初始 + 1 次 token 刷新后）
  for (let attempt = 0; attempt < 2; attempt++) {
    try {
      return await uploadOnce(url, content, overwrite, onprogress, scope);
    } catch (error: any) {
      lastError = error;
      
      // 如果是 401 且是第一次尝试，尝试刷新 token
      if (error.status === 401 && attempt === 0) {
        try {
          const { renew } = await import("@/utils/auth");
          await renew(authStore.jwt);
          continue; // 刷新成功后重试
        } catch {
          // 刷新失败，跳出循环
          break;
        }
      }
      
      // 其他错误直接抛出
      throw error;
    }
  }
  
  throw lastError;
}

async function uploadOnce(
  url: string,
  content: Blob,
  overwrite: boolean,
  onprogress: (loaded: number) => void,
  scope?: string
): Promise<string> {
  const authStore = useAuthStore();
  url = removePrefix(url);
  
  // ... 使用 XMLHttpRequest 上传，监听 progress 事件
}
```

#### Step 4: 修改上传 store 使用新的 API

**修改文件**: `frontend/src/stores/upload.ts`

```typescript
// 替换现有的 api.post 调用
import { uploadWithRetry } from "@/api/upload";

// 在 processUploads 中
await uploadWithRetry(
  upload.path,
  upload.file!,
  upload.overwrite,
  onUpload,
  scope
).catch((err) => {
  // 上传失败但不触发 logout
  console.error("Upload failed:", err);
  $showError(err);
});
```

---

## 实施计划

### Task 1: 添加 token 有效性检查函数
- 创建 `frontend/src/utils/token.ts`
- 实现 `isTokenExpiringSoon()` 函数
- 实现 `ensureTokenValid()` 函数

### Task 2: 修改 upload store
- 导入 token 检查函数
- 在上传开始前调用
- 处理 token 刷新失败的情况

### Task 3: 创建独立的 upload API（可选，如果 Task 2 不够）
- 创建 `frontend/src/api/upload.ts`
- 实现带重试的上传函数

---

## 测试计划

1. **Token 即将过期时上传**
   - 设置 token 剩余时间 < 5 分钟
   - 开始上传
   - 验证：token 被刷新，上传成功

2. **上传期间 token 过期**
   - 开始上传后，等待 token 过期
   - 验证：自动刷新 token，上传继续

3. **网络不稳定**
   - 模拟网络错误
   - 验证：不会导致页面刷新

---

## 回滚方案

如果新方案有问题，可以回滚到当前的临时修复：
- 当前在 `utils.ts` 中的 401 处理逻辑保留
- 逐步添加新功能，测试稳定后移除临时修复
