# MoveCopyModal 组件

## 组件介绍

MoveCopyModal 是一个通用的移动/复制文件/文件夹对话框组件，用于 Storage Browser 应用中。它提供了统一的界面，支持用户选择目标文件夹，并处理文件冲突检测和解决。

## 功能特性

- 统一的移动和复制操作界面
- 支持文件夹导航和选择
- 实时冲突检测
- 提供覆盖、重命名或取消选项
- 支持键盘快捷键（Enter 确认，Esc 取消）
- 响应式设计，适配不同屏幕尺寸

## 使用方法

### 基本用法

```vue
<template>
  <MoveCopyModal mode="move" />
</template>

<script setup lang="ts">
import MoveCopyModal from "@/components/prompts/MoveCopyModal.vue";
</script>
```

### 在 Prompts.vue 中的集成

在 `Prompts.vue` 中，组件通过模态框系统动态加载：

```typescript
// Prompts.vue
const components = new Map<string, any>([
  // ...其他组件
  ["move", MoveCopyModal],
  ["copy", MoveCopyModal],
  // ...其他组件
]);

watch(currentPromptName, (newValue) => {
  const modal = components.get(newValue!);
  if (!modal) return;

  const props = currentPrompt?.props || {};
  if (modal === MoveCopyModal) {
    props.mode = newValue; // 使用 prompt 名称作为 mode
  }

  // ...
});
```

## Props 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| mode | `"move" \| "copy"` | 是 | 操作模式：移动或复制 |

## 事件

| 事件名 | 参数 | 描述 |
|--------|------|------|
| closeHovers | - | 关闭所有悬停层和模态框 |

## 组件结构

### 模板部分

```vue
<template>
  <div class="card floating" @keydown.esc="closeHovers" @keydown.enter="handleEnterKey">
    <!-- 标题栏 -->
    <div class="card-title">
      <h2>{{ mode === 'move' ? $t("prompts.move") : $t("prompts.copy") }}</h2>
    </div>

    <!-- 内容区 -->
    <div class="card-content">
      <!-- 复制模式提示信息 -->
      <p v-if="mode === 'copy'">{{ $t("prompts.copyMessage") }}</p>

      <!-- 文件列表组件 -->
      <file-list
        ref="fileList"
        @update:selected="(val: string | null) => (dest = val)"
        :exclude="excludedFolders"
        tabindex="1"
      />
    </div>

    <!-- 操作按钮区 -->
    <div class="card-action">
      <!-- 新建文件夹按钮 -->
      <button
        v-if="user?.perm.create"
        class="button button--flat"
        @click="fileList?.createDir()"
      >
        {{ $t("sidebar.newFolder") }}
      </button>

      <!-- 取消和确认按钮 -->
      <div>
        <button class="button button--flat button--grey" @click="closeHovers">
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="handleAction"
          :disabled="mode === 'move' && $route.path === dest"
        >
          {{ mode === 'move' ? $t("buttons.move") : $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>
```

### 脚本部分

核心方法：

1. **handleAction()**：处理确认操作，包括：
   - 验证选择的目标文件夹
   - 准备文件/文件夹信息
   - 检查冲突
   - 处理冲突解决
   - 调用 API 执行操作

2. **handleEnterKey()**：处理 Enter 键快捷键

3. **action()**：实际执行移动/复制操作的内部方法

## 依赖

- `vue`：核心框架
- `vue-i18n`：国际化支持
- `vue-router`：路由管理
- `pinia`：状态管理
- `@/stores/file`：文件操作状态管理
- `@/stores/layout`：布局状态管理
- `@/stores/auth`：用户认证状态管理
- `@/api/files`：文件操作 API
- `@/utils/buttons`：按钮状态管理
- `@/utils/upload`：上传工具函数
- `@/api/utils`：API 工具函数
- `@/utils/path`：路径处理工具函数

## 相关组件

- `FileList.vue`：用于文件夹导航和选择
- `BaseModal.vue`：基础模态框组件
- `ConflictDialog.vue`：冲突解决对话框

## 代码优化建议

1. **错误处理**：当前使用控制台输出和 $showError 函数，建议统一使用更健壮的错误处理机制
2. **类型定义**：可以为 items 数组创建更具体的类型定义
3. **可测试性**：可以将 API 调用和冲突检查逻辑提取到单独的 composable 中，以便于单元测试
4. **性能优化**：对于大量文件的操作，可以考虑添加进度条或分批处理

## Async Close Fix

### Summary

This fix addresses a race condition where modal close operations were not properly awaited before router navigation, causing UI inconsistencies.

### Technical Changes

The fix involved three key files:

1. `/frontend/src/types/layout.d.ts` - Updated `PopupProps` interface to allow async close functions
2. `/frontend/src/stores/layout.ts` - Made `closeHovers()` and `closeCurrentHover()` async, updated `setCloseOnPrompt()` type signature
3. `/frontend/src/components/prompts/MoveCopyModal.vue` - Updated all `closeHovers()` calls to be awaited

### The Problem

The original implementation had synchronous close operations that could lead to race conditions when combined with router navigation:

```typescript
// Before (problematic)
layoutStore.closeHovers(); // Synchronous - returns immediately
router.push({ path: dest.value! }); // Navigation starts before modal closes
```

### The Solution

```typescript
// After (fixed)
await layoutStore.closeHovers(); // Wait for modal to fully close
router.push({ path: dest.value! }); // Then navigate
```

### Key Changes in Store

```typescript
// stores/layout.ts
async closeHovers() {
  console.log("closeHovers called, current prompts:", this.prompts.length);
  if (this.prompts.length > 0) {
    const popped = this.prompts.pop();
    console.log("Popped prompt:", popped?.prompt);
    if (popped?.close) {
      await popped.close(); // Wait for prompt-specific close logic
    }
  }
  console.log("After closeHovers, prompts:", this.prompts.length);
}
```

### Test Scenarios

1. **Basic Functionality Test**: Verify modal closes before navigation
2. **Cancel Button Test**: Check modal closes properly when canceling
3. **Same Path Test**: Test copy operation to same folder (should reload instead of navigate)
4. **Conflict Resolution Test**: Verify both conflict dialog and main modal close completely
5. **Keyboard Navigation Test**: Test with Enter and Esc keys

### Handling Similar Issues in Other Components

For other modal components facing similar race conditions:

1. Update type definitions to accept async close functions
2. Make store actions async
3. Await close operations before navigation
4. Add proper error handling

## 变更历史

- 替代了旧的 Move.vue 和 Copy.vue 组件
- 合并了移动和复制功能到单一组件中
- 优化了用户界面和交互体验
- 添加了完整的类型定义
- 修复了模态框关闭与路由导航的竞争条件问题
