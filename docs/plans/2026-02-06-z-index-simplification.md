# Z-Index 简化与优化实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 简化项目中所有 z-index 设置，建立统一的 CSS 变量层级系统，消除不必要的 !important 覆盖，确保 modal-overlay 等组件不会过度遮挡其他元素。

**Architecture:** 建立基于 CSS 变量 (`:root`) 的 z-index 层级系统，定义清晰的层级规范（背景、内容、浮动元素、覆盖层、模态框等），移除不必要的硬编码值，使用更合理的层级数值。

**Tech Stack:** Vue.js 3, CSS, Vite, vue-toastification, vue-final-modal (vfm)

---

## 任务 N: 建立 Z-Index 变量系统

**Files:**
- Modify: `frontend/src/css/_variables.css`

**Step 1: 添加 z-index CSS 变量定义**

```css
:root {
  /* 层级系统 (从低到高) */
  --z-dropdown: 100;
  --z-sticky: 200;
  --z-fixed: 300;
  --z-modal-backdrop: 400;
  --z-modal: 500;
  --z-popover: 600;
  --z-tooltip: 700;
  --z-notification: 800;
  --z-overlay: 900;      /* click-overlay 等覆盖层 */
  --z-loading: 1000;      /* 进度条等 */
}
```

**Step 2: 运行验证命令**

Run: `cd frontend && pnpm run typecheck`
Expected: TypeScript 无错误（CSS 更改不影响类型检查）

---

## 任务 N: 简化 base.css z-index 设置

**Files:**
- Modify: `frontend/src/css/base.css:99-100`, `:140-141`, `:185-186`

**Step 1: 读取并修改 .breadcrumbs z-index**

Old: `z-index: 999;`
New: `z-index: var(--z-sticky, 200);`

**Step 2: 修改 .progress z-index**

Old: `z-index: 9999999999;`
New: `z-index: var(--z-loading, 1000);`

**Step 3: 修改 .vfm-modal z-index**

Old: `z-index: 9999999 !important;`
New: `z-index: var(--z-modal, 500);` (移除 !important)

**Step 4: 运行验证命令**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过，无警告

---

## 任务 N: 简化 styles.css z-index 设置

**Files:**
- Modify: `frontend/src/css/styles.css:175`, `:184`

**Step 1: 修改 #previewer z-index**

Old: `z-index: 9999;`
New: `z-index: var(--z-overlay, 900);`

**Step 2: 修改 #previewer header z-index**

Old: `z-index: 19999;`
New: `z-index: var(--z-popover, 600);` (移除 !important)

**Step 3: 运行验证命令**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过

---

## 任务 N: 简化 header.css z-index 设置

**Files:**
- Modify: `frontend/src/css/header.css:2`, `:81`, `:93`, `:142`

**Step 1: 读取并分析 header.css 中的 z-index 使用**

```css
/* 预期修改 */
header { z-index: var(--z-fixed, 300); }
header .actions { z-index: var(--z-fixed, 300); }
.dropdown { z-index: var(--z-dropdown, 100); }
header .dropdown { z-index: var(--z-dropdown, 100); }
```

**Step 2: 应用修改并验证**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过

---

## 任务 N: 简化 mobile.css z-index 设置

**Files:**
- Modify: `frontend/src/css/mobile.css:31`, `:72`, `:92`

**Step 1: 移除过高的 z-index 值**

Old: `z-index: 99999;`
New: `z-index: var(--z-overlay, 900);`

Old: `z-index: 99999 !important;`
New: `z-index: var(--z-modal, 500);` (移除 !important)

**Step 2: 验证修改**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过

---

## 任务 N: 简化 listing.css z-index 设置

**Files:**
- Modify: `frontend/src/css/listing.css:220`, `:267`

**Step 1: 修改 listing 相关 z-index**

Old: `z-index: 999;`
New: `z-index: var(--z-dropdown, 100);`

Old: `z-index: 99999;`
New: `z-index: var(--z-overlay, 900);`

**Step 2: 验证修改**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过

---

## 任务 N: 简化 _shell.css z-index 设置

**Files:**
- Modify: `frontend/src/css/_shell.css:8`, `:43`

**Step 1: 修改 shell 相关 z-index**

Old: `z-index: 9999;`
New: `z-index: var(--z-modal, 500);`

Old: `z-index: 9998;`
New: `z-index: var(--z-modal-backdrop, 400);`

**Step 2: 验证修改**

Run: `cd frontend && pnpm run lint:fix`
Expected: ESLint 通过

---

## 任务 N: 更新 Layout.vue z-index 设置

**Files:**
- Modify: `frontend/src/views/Layout.vue:148`, `:235`, `:254`

**Step 1: 读取并修改 Layout.vue**

```vue
<style scoped>
.sidebar {
  /* sidebar 从 1000 改为更合理的值 */
  z-index: var(--z-fixed, 300);
}

.modal-overlay {
  /* modal-overlay 从 1000/1002 改为使用变量 */
  z-index: var(--z-overlay, 900);
}
</style>
```

**Step 2: 更新 JavaScript 中的 z-index 动态赋值**

Old: `zIndex.value = 1000;`
New: `zIndex.value = 900; /* var(--z-overlay) */`

Old: `zIndex.value = 1002;`
New: `zIndex.value = 500; /* var(--z-modal) */`

**Step 3: 验证修改**

Run: `cd frontend && pnpm run typecheck && pnpm run lint:fix`
Expected: 通过

---

## 任务 N: 更新 HeaderBar.vue z-index 设置

**Files:**
- Modify: `frontend/src/components/header/HeaderBar.vue:346`, `:434`, `:529`

**Step 1: 读取并修改 HeaderBar.vue**

```javascript
// 移动设备 modal-overlay 修改
// Old: zIndex = 999
// New: 使用 CSS 变量，不再动态设置
```

**Step 2: 移除不必要的动态 z-index 设置**

**Step 3: 验证修改**

Run: `cd frontend && pnpm run typecheck && pnpm run lint:fix`
Expected: 通过

---

## 任务 N: 更新 Sidebar.vue z-index 设置

**Files:**
- Modify: `frontend/src/components/Sidebar.vue:281`, `:355`

**Step 1: 读取并修改 Sidebar.vue**

Old: `z-index: 900;`
New: `z-index: var(--z-fixed, 300);`

**Step 2: 移除动态 z-index 设置代码**

**Step 3: 验证修改**

Run: `cd frontend && pnpm run typecheck && pnpm run lint:fix`
Expected: 通过

---

## 任务 N: 更新 Share.vue z-index 设置

**Files:**
- Modify: `frontend/src/views/Share.vue:43`, `:84`

**Step 1: 读取并修改 Share.vue**

Old: `style="z-index: 9999999"`
New: `:style="{ zIndex: 'var(--z-modal, 500)' }"`

Old: `z-index: 999;`
New: `z-index: var(--z-dropdown, 100);`

**Step 2: 验证修改**

Run: `cd frontend && pnpm run typecheck && pnpm run lint:fix`
Expected: 通过

---

## 任务 N: 简化 Prompts.vue z-index 设置

**Files:**
- Modify: `frontend/src/components/prompts/Prompts.vue:132`

**Step 1: 读取并修改 Prompts.vue**

Old: `z-index: 1000;`
New: `z-index: var(--z-modal, 500);`

**Step 2: 验证修改**

Run: `cd frontend && pnpm run typecheck && pnpm run lint:fix`
Expected: 通过

---

## 任务 N: 检查其他 Vue 组件

**Files:**
- 检查: `frontend/src/components/**/*.vue`
- 检查: `frontend/src/views/**/*.vue`

**Step 1: 搜索遗漏的 z-index 使用**

Run: `cd frontend && grep -r "z-index" src/ --include="*.vue" | grep -v "^Binary"`
Expected: 输出所有 z-index 使用，逐一检查

**Step 2: 简化剩余的硬编码 z-index**

**Step 3: 批量验证**

Run: `cd frontend && pnpm run build`
Expected: 构建成功

---

## 任务 N: 构建验证

**Files:**
- Modify: `frontend/src/css/_variables.css`

**Step 1: 运行前端构建**

Run: `cd frontend && pnpm run build`
Expected: 构建成功，无 CSS 错误

**Step 2: 运行类型检查**

Run: `cd frontend && pnpm run typecheck`
Expected: 通过

**Step 3: 运行 linter**

Run: `cd frontend && pnpm run lint`
Expected: 无错误

**Step 4: 提交更改**

```bash
git add frontend/src/css/
git commit -m "refactor: 建立 z-index 变量系统，简化层级设置"
```

---

## 验证清单

- [ ] CSS 变量系统已建立
- [ ] 所有硬编码 z-index 已简化
- [ ] 所有 !important 已移除（或保留理由充分）
- [ ] modal-overlay 不再过度遮挡其他组件
- [ ] 前端构建成功
- [ ] 类型检查通过
- [ ] Linter 通过

---

**Plan complete saved to:** `docs/plans/2026-02-06-z-index-simplification.md`

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
