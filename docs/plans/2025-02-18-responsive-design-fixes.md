# Responsive Design Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix responsive design issues on small devices including header search duplication, file-selection positioning, and other mobile layout problems.

**Architecture:** Create reusable responsive utilities and components, fix existing CSS and Vue components to properly handle mobile breakpoints, ensure consistent mobile detection between CSS and JavaScript.

**Tech Stack:** Vue.js 3 (Composition API), CSS with custom media queries, TypeScript

---

### Task 1: Create Responsive Utilities

**Files:**
- Create: `frontend/src/utils/responsive.ts`
- Create: `frontend/src/utils/responsive.test.ts`

**Step 1: Write the failing test**

```typescript
// frontend/src/utils/responsive.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { isMobile, isSmallMobile, isVerySmallMobile, responsiveClass } from './responsive'

describe('responsive utilities', () => {
  beforeEach(() => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(800)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('isMobile returns false for desktop width', () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(800)
    expect(isMobile.value).toBe(false)
  })

  it('isMobile returns true for mobile width', () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(600)
    expect(isMobile.value).toBe(true)
  })

  it('isSmallMobile returns true for small mobile width', () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(400)
    expect(isSmallMobile.value).toBe(true)
  })

  it('isVerySmallMobile returns true for very small mobile width', () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(350)
    expect(isVerySmallMobile.value).toBe(true)
  })

  it('responsiveClass returns correct classes', () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(600)
    expect(responsiveClass('desktop', 'mobile')).toBe('mobile')
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm test responsive.test.ts -v`
Expected: FAIL with "Cannot find module './responsive'"

**Step 3: Write minimal implementation**

```typescript
// frontend/src/utils/responsive.ts
import { computed, ref, onMounted, onUnmounted } from 'vue'

const MOBILE_BREAKPOINT = 736
const SMALL_MOBILE_BREAKPOINT = 480
const VERY_SMALL_MOBILE_BREAKPOINT = 400

const windowWidth = ref(window.innerWidth)

const updateWindowWidth = () => {
  windowWidth.value = window.innerWidth
}

export const isMobile = computed(() => windowWidth.value <= MOBILE_BREAKPOINT)
export const isSmallMobile = computed(() => windowWidth.value <= SMALL_MOBILE_BREAKPOINT)
export const isVerySmallMobile = computed(() => windowWidth.value <= VERY_SMALL_MOBILE_BREAKPOINT)

export const responsiveClass = (desktopClass: string, mobileClass: string) => {
  return isMobile.value ? mobileClass : desktopClass
}

export const useResponsive = () => {
  onMounted(() => {
    window.addEventListener('resize', updateWindowWidth)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', updateWindowWidth)
  })

  return {
    isMobile,
    isSmallMobile,
    isVerySmallMobile,
    responsiveClass
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd frontend && pnpm test responsive.test.ts -v`
Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/utils/responsive.ts frontend/src/utils/responsive.test.ts
git commit -m "feat: add responsive utilities"
```

---

### Task 2: Fix Header Search Duplication

**Files:**
- Modify: `frontend/src/components/header/HeaderBar.vue:44-58`
- Modify: `frontend/src/components/Search.vue`
- Create: `frontend/src/components/header/MobileSearch.vue`

**Step 1: Write the failing test for MobileSearch component**

```typescript
// frontend/src/components/__tests__/MobileSearch.test.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import MobileSearch from '../header/MobileSearch.vue'

describe('MobileSearch', () => {
  it('renders search input when active', () => {
    const wrapper = mount(MobileSearch, {
      props: {
        active: true
      }
    })
    expect(wrapper.find('input[type="search"]').exists()).toBe(true)
  })

  it('emits search event when input changes', async () => {
    const wrapper = mount(MobileSearch)
    await wrapper.find('input').setValue('test query')
    expect(wrapper.emitted('search')).toBeTruthy()
    expect(wrapper.emitted('search')[0]).toEqual(['test query'])
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm test MobileSearch.test.ts -v`
Expected: FAIL with "Cannot find module '../header/MobileSearch.vue'"

**Step 3: Create MobileSearch component**

```vue
<!-- frontend/src/components/header/MobileSearch.vue -->
<template>
  <div v-if="active" class="mobile-search-overlay">
    <div class="mobile-search-container">
      <input
        ref="searchInput"
        type="search"
        :placeholder="placeholder"
        :value="modelValue"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        @keyup.enter="$emit('search', modelValue)"
        class="mobile-search-input"
      />
      <button @click="$emit('close')" class="mobile-search-close">
        ✕
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  active: boolean
  modelValue: string
  placeholder?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
  search: [query: string]
  close: []
}>()
</script>

<style scoped>
.mobile-search-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: var(--z-modal, 500);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 4rem;
}

.mobile-search-container {
  background: white;
  border-radius: 8px;
  padding: 1rem;
  width: 90%;
  max-width: 400px;
  display: flex;
  gap: 0.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.mobile-search-input {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.mobile-search-close {
  padding: 0.75rem 1rem;
  background: #f0f0f0;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1.2rem;
}

.mobile-search-close:hover {
  background: #e0e0e0;
}
</style>
```

**Step 4: Update HeaderBar.vue to use responsive utilities**

```vue
<!-- frontend/src/components/header/HeaderBar.vue (lines 44-58) -->
<template>
  <header class="header">
    <!-- Existing code before search... -->
    
    <!-- Replace lines 44-58 with: -->
    <div v-if="!isMobile" class="search-container">
      <search v-model="searchQuery" @search="handleSearch" />
    </div>
    
    <action
      v-if="isMobile"
      icon="search"
      :title="$t('search')"
      @click="showMobileSearch = true"
    />
    
    <!-- Mobile search overlay -->
    <mobile-search
      v-model="searchQuery"
      :active="showMobileSearch"
      @search="handleSearch"
      @close="showMobileSearch = false"
    />
    
    <!-- Rest of existing code... -->
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useResponsive } from '@/utils/responsive'
import Search from '@/components/Search.vue'
import MobileSearch from './MobileSearch.vue'
import Action from './Action.vue'

const { isMobile } = useResponsive()
const { t } = useI18n()

const searchQuery = ref('')
const showMobileSearch = ref(false)

const handleSearch = (query: string) => {
  // Existing search logic
  showMobileSearch.value = false
}
</script>
```

**Step 5: Update CSS to remove duplicate search hiding**

```css
/* frontend/src/css/header-mobile.css - Remove or update line 5-7 */
/* Remove: header .search-container { display: none; } */

/* Add responsive behavior to existing search container */
@media (max-width: 736px) {
  header .search-container {
    display: none;
  }
  
  /* Ensure mobile search button is visible */
  header .mobile-search-button {
    display: block;
  }
}
```

**Step 6: Run tests to verify they pass**

Run: `cd frontend && pnpm test MobileSearch.test.ts -v`
Expected: PASS

**Step 7: Commit**

```bash
git add frontend/src/components/header/MobileSearch.vue frontend/src/components/header/HeaderBar.vue frontend/src/css/header-mobile.css frontend/src/components/__tests__/MobileSearch.test.ts
git commit -m "feat: fix header search duplication with mobile search overlay"
```

---

### Task 3: Fix File-Selection Component Positioning

**Files:**
- Modify: `frontend/src/views/files/FileListing.vue:3-37`
- Modify: `frontend/src/css/mobile.css:59-75`
- Create: `frontend/src/components/files/FileSelectionBar.vue`

**Step 1: Write the failing test for FileSelectionBar**

```typescript
// frontend/src/components/__tests__/FileSelectionBar.test.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import FileSelectionBar from '../files/FileSelectionBar.vue'

describe('FileSelectionBar', () => {
  it('shows selected count', () => {
    const wrapper = mount(FileSelectionBar, {
      props: {
        selectedCount: 5
      }
    })
    expect(wrapper.text()).toContain('5 selected')
  })

  it('emits clear event when clear button clicked', async () => {
    const wrapper = mount(FileSelectionBar, {
      props: {
        selectedCount: 3
      }
    })
    await wrapper.find('.clear-selection').trigger('click')
    expect(wrapper.emitted('clear')).toBeTruthy()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm test FileSelectionBar.test.ts -v`
Expected: FAIL with "Cannot find module '../files/FileSelectionBar.vue'"

**Step 3: Create reusable FileSelectionBar component**

```vue
<!-- frontend/src/components/files/FileSelectionBar.vue -->
<template>
  <div :class="['file-selection-bar', { 'is-visible': selectedCount > 0 }]" :style="styles">
    <div class="file-selection-content">
      <span class="selected-count">{{ selectedCount }} selected</span>
      <div class="selection-actions">
        <button @click="$emit('clear')" class="clear-selection">
          {{ $t('clear') }}
        </button>
        <button @click="$emit('action', 'download')" class="action-button" v-if="showDownload">
          {{ $t('download') }}
        </button>
        <button @click="$emit('action', 'share')" class="action-button" v-if="showShare">
          {{ $t('share') }}
        </button>
        <button @click="$emit('action', 'delete')" class="action-button delete" v-if="showDelete">
          {{ $t('delete') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useResponsive } from '@/utils/responsive'

const { isMobile } = useResponsive()
const { t } = useI18n()

const props = withDefaults(defineProps<{
  selectedCount: number
  position?: 'fixed' | 'sticky' | 'absolute'
  zIndex?: number
  showDownload?: boolean
  showShare?: boolean
  showDelete?: boolean
}>(), {
  position: 'fixed',
  zIndex: 300,
  showDownload: true,
  showShare: true,
  showDelete: true
})

defineEmits<{
  clear: []
  action: [action: string]
}>()

const styles = computed(() => ({
  zIndex: props.zIndex,
  position: props.position
}))
</script>

<style scoped>
.file-selection-bar {
  background: var(--bg-primary, white);
  border-top: 1px solid var(--border-color, #e0e0e0);
  padding: 1rem;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease;
  transform: translateY(100%);
}

.file-selection-bar.is-visible {
  transform: translateY(0);
}

.file-selection-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1200px;
  margin: 0 auto;
}

.selected-count {
  font-weight: 500;
  color: var(--text-primary, #333);
}

.selection-actions {
  display: flex;
  gap: 0.5rem;
}

.clear-selection,
.action-button {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--bg-secondary, #f5f5f5);
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.clear-selection:hover,
.action-button:hover {
  background: var(--bg-hover, #e8e8e8);
}

.action-button.delete {
  background: var(--danger-light, #fee);
  border-color: var(--danger-border, #fcc);
  color: var(--danger, #d00);
}

.action-button.delete:hover {
  background: var(--danger-hover, #fdd);
}

@media (max-width: 736px) {
  .file-selection-bar {
    padding: 0.75rem;
    z-index: var(--z-fixed, 300);
  }
  
  .selection-actions {
    gap: 0.25rem;
  }
  
  .clear-selection,
  .action-button {
    padding: 0.4rem 0.75rem;
    font-size: 0.85rem;
  }
}
</style>
```

**Step 4: Update FileListing.vue to use new component**

```vue
<!-- frontend/src/views/files/FileListing.vue (lines 3-37) -->
<template>
  <div class="file-listing">
    <!-- Replace existing file-selection div with: -->
    <file-selection-bar
      :selected-count="fileStore.selectedCount"
      :position="isMobile ? 'fixed' : 'sticky'"
      :z-index="isMobile ? 300 : 200"
      @clear="fileStore.clearSelection()"
      @action="handleFileAction"
    />
    
    <!-- Rest of existing code... -->
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from '@/stores/file'
import { useResponsive } from '@/utils/responsive'
import FileSelectionBar from '@/components/files/FileSelectionBar.vue'

const fileStore = useFileStore()
const { isMobile } = useResponsive()

const handleFileAction = (action: string) => {
  // Handle file actions (download, share, delete)
  switch (action) {
    case 'download':
      fileStore.downloadSelected()
      break
    case 'share':
      fileStore.shareSelected()
      break
    case 'delete':
      fileStore.deleteSelected()
      break
  }
}
</script>
```

**Step 5: Update mobile.css for better positioning**

```css
/* frontend/src/css/mobile.css - Update lines 59-75 */
#file-selection {
  /* Remove old styles and use component styles instead */
  display: none; /* Will be handled by component */
}

/* Add responsive adjustments for file listing */
@media (max-width: 736px) {
  .file-listing {
    padding-bottom: 4rem; /* Space for fixed selection bar */
  }
}
```

**Step 6: Run tests to verify they pass**

Run: `cd frontend && pnpm test FileSelectionBar.test.ts -v`
Expected: PASS

**Step 7: Commit**

```bash
git add frontend/src/components/files/FileSelectionBar.vue frontend/src/views/files/FileListing.vue frontend/src/css/mobile.css frontend/src/components/__tests__/FileSelectionBar.test.ts
git commit -m "feat: fix file-selection positioning with reusable component"
```

---

### Task 4: Audit and Fix Other Responsive Issues

**Files:**
- Audit: `frontend/src/css/mobile.css`
- Audit: `frontend/src/css/header-mobile.css`
- Audit: `frontend/src/views/Layout.vue`
- Create: `frontend/src/utils/responsive-audit.ts`

**Step 1: Create responsive audit utility**

```typescript
// frontend/src/utils/responsive-audit.ts
export const RESPONSIVE_ISSUES = [
  {
    file: 'frontend/src/css/mobile.css',
    issues: [
      'Check all @media queries for consistency with breakpoints (736px, 480px, 400px)',
      'Ensure z-index values use CSS custom properties',
      'Verify fixed positioning elements don\'t overlap on small screens'
    ]
  },
  {
    file: 'frontend/src/css/header-mobile.css',
    issues: [
      'Remove duplicate search hiding rules now handled by components',
      'Check header height adjustments on very small screens (<400px)',
      'Ensure action buttons have proper touch targets (min 44x44px)'
    ]
  },
  {
    file: 'frontend/src/views/Layout.vue',
    issues: [
      'Verify sidebar overlay doesn\'t conflict with other overlays',
      'Check transition animations work on mobile',
      'Ensure mobile layout has proper viewport handling'
    ]
  },
  {
    file: 'frontend/src/components/Sidebar.vue',
    issues: [
      'Check mobile sidebar collapse behavior',
      'Verify menu items are accessible on touch devices'
    ]
  },
  {
    file: 'frontend/src/components/Breadcrumbs.vue',
    issues: [
      'Check truncation on mobile screens',
      'Verify touch targets for breadcrumb items'
    ]
  }
]

export const auditResponsiveIssues = () => {
  const issues: string[] = []
  
  RESPONSIVE_ISSUES.forEach(({ file, issues: fileIssues }) => {
    console.log(`\n📋 ${file}:`)
    fileIssues.forEach(issue => {
      console.log(`  • ${issue}`)
    })
  })
  
  return issues
}
```

**Step 2: Run audit and identify issues**

Run: `cd frontend && node -e "import('./src/utils/responsive-audit.ts').then(m => m.auditResponsiveIssues())"`
Expected: Output listing all responsive issues to check

**Step 3: Fix mobile.css issues**

```css
/* frontend/src/css/mobile.css - Add/update */
/* Ensure consistent breakpoints */
@media (max-width: 736px) {
  /* Base mobile adjustments */
  body {
    -webkit-text-size-adjust: 100%;
  }
  
  /* Improve touch targets */
  button, 
  .action-button,
  [role="button"] {
    min-height: 44px;
    min-width: 44px;
  }
  
  /* Fix potential overflow issues */
  .container {
    overflow-x: hidden;
  }
}

@media (max-width: 480px) {
  /* Smaller mobile adjustments */
  .header {
    padding: 0.5rem;
  }
  
  /* Adjust font sizes for readability */
  body {
    font-size: 14px;
  }
}

@media (max-width: 400px) {
  /* Very small screen adjustments */
  .header .actions {
    gap: 0.25rem;
  }
  
  /* Ensure content fits */
  .content {
    padding: 0.5rem;
  }
}
```

**Step 4: Fix header-mobile.css issues**

```css
/* frontend/src/css/header-mobile.css - Update */
@media (max-width: 736px) {
  header {
    height: auto;
    min-height: 60px;
    padding: 0.5rem;
  }
  
  /* Remove duplicate search rules */
  /* header .search-container { display: none; } */
  
  /* Ensure action buttons are touch-friendly */
  header .action-button {
    min-width: 44px;
    min-height: 44px;
    padding: 0.75rem;
  }
}

@media (max-width: 480px) {
  header {
    padding: 0.25rem;
  }
  
  .header-title {
    font-size: 1rem;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 400px) {
  header {
    flex-wrap: wrap;
  }
  
  .header-actions {
    order: 3;
    width: 100%;
    justify-content: center;
    margin-top: 0.5rem;
  }
}
```

**Step 5: Fix Layout.vue mobile issues**

```vue
<!-- frontend/src/views/Layout.vue - Update mobile sidebar section -->
<template>
  <!-- Existing code... -->
  
  <!-- Update mobile sidebar overlay z-index -->
  <div 
    v-if="isMobile && sidebarOpen" 
    class="sidebar-overlay"
    :style="{ zIndex: isMobile ? 400 : 'auto' }"
    @click="sidebarOpen = false"
  />
  
  <!-- Rest of existing code... -->
</template>

<script setup lang="ts">
import { useResponsive } from '@/utils/responsive'

const { isMobile } = useResponsive()
// Rest of existing code...
</script>

<style scoped>
/* Ensure mobile sidebar has proper z-index */
@media (max-width: 736px) {
  .sidebar-overlay {
    z-index: var(--z-overlay, 900);
  }
  
  .sidebar {
    z-index: var(--z-modal, 500);
  }
}
</style>
```

**Step 6: Run lint and typecheck**

Run: `cd frontend && pnpm run lint:fix && pnpm run typecheck`
Expected: No errors

**Step 7: Commit**

```bash
git add frontend/src/utils/responsive-audit.ts frontend/src/css/mobile.css frontend/src/css/header-mobile.css frontend/src/views/Layout.vue
git commit -m "feat: audit and fix remaining responsive issues"
```

---

### Task 5: Create Responsive Design Documentation

**Files:**
- Create: `docs/RESPONSIVE-DESIGN.md`
- Update: `CONTRIBUTING.md`

**Step 1: Create responsive design guide**

```markdown
# Responsive Design Guide

## Breakpoints
- **Mobile:** ≤ 736px
- **Small Mobile:** ≤ 480px  
- **Very Small Mobile:** ≤ 400px

## Responsive Utilities
Use the `useResponsive()` composable:

```typescript
import { useResponsive } from '@/utils/responsive'

const { isMobile, isSmallMobile, responsiveClass } = useResponsive()

// Conditional rendering
<template v-if="isMobile">Mobile content</template>

// Conditional classes
<div :class="responsiveClass('desktop-class', 'mobile-class')">
```

## Component Patterns

### 1. Mobile Search
```vue
<mobile-search
  v-model="query"
  :active="showMobileSearch"
  @search="handleSearch"
  @close="showMobileSearch = false"
/>
```

### 2. File Selection Bar
```vue
<file-selection-bar
  :selected-count="selectedCount"
  :position="isMobile ? 'fixed' : 'sticky'"
  @clear="clearSelection"
  @action="handleAction"
/>
```

### 3. Responsive Layout
- Use `position: fixed` for mobile overlays
- Set appropriate `z-index` using CSS custom properties
- Add bottom padding for fixed elements

## CSS Guidelines
1. Use media queries with consistent breakpoints
2. Prefer CSS custom properties for theming
3. Ensure touch targets ≥ 44×44px
4. Test on actual mobile devices

## Testing
- Unit test responsive utilities
- Visual regression testing for breakpoints
- Test touch interactions
```

**Step 2: Update contributing guide**

```markdown
<!-- CONTRIBUTING.md - Add responsive design section -->
## Responsive Design

All UI components must be responsive. Follow these guidelines:

1. **Test on multiple screen sizes** (736px, 480px, 400px breakpoints)
2. **Use responsive utilities** from `@/utils/responsive`
3. **Follow component patterns** in `docs/RESPONSIVE-DESIGN.md`
4. **Ensure accessibility** on touch devices
```

**Step 3: Commit**

```bash
git add docs/RESPONSIVE-DESIGN.md CONTRIBUTING.md
git commit -m "docs: add responsive design guide and update contributing"
```

---

### Task 6: Final Testing and Verification

**Files:**
- All modified files
- Run: `make test` and `make lint`

**Step 1: Run backend tests**

Run: `make test`
Expected: All tests pass

**Step 2: Run frontend tests**

Run: `cd frontend && pnpm test`
Expected: All tests pass

**Step 3: Run linting**

Run: `make lint`
Expected: No linting errors

**Step 4: Build verification**

Run: `make build`
Expected: Successful build

**Step 5: Final commit**

```bash
git add .
git commit -m "chore: final verification of responsive design fixes"
```

---

Plan complete and saved to `docs/plans/2025-02-18-responsive-design-fixes.md`.

Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**