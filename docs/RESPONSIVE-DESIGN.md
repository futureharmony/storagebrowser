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