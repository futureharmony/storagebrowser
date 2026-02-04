<template>
  <div class="layout-container">
    <!-- Header -->
    <header class="layout-header">
      <slot name="header">
        <div class="default-header">Header</div>
      </slot>
    </header>

    <div class="layout-main">
      <!-- Sidebar -->
      <div class="layout-sidebar">
        <slot name="sidebar">
          <div class="default-sidebar">Sidebar</div>
        </slot>
      </div>

      <!-- Content -->
      <main class="layout-content">
        <slot name="content">
          <div class="default-content">Content</div>
        </slot>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
// This Layout component provides a consistent structure
// with header, sidebar, and content areas
</script>

<style scoped>
.layout-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  overflow: hidden;
  background: var(--background);
}

/* Header */
.layout-header {
  flex: 0 0 auto;
  height: 4em;
  background: var(--surfacePrimary);
  border-bottom: 1px solid var(--divider);
  z-index: 1000;
  position: relative;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* Main area (sidebar + content) */
.layout-main {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

/* Sidebar */
.layout-sidebar {
  flex: 0 0 auto;
  width: 200px;
  background: var(--surfacePrimary);
  border-right: 1px solid var(--divider);
  overflow-y: auto;
  position: relative;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Collapsed sidebar */
.layout-sidebar.collapsed {
  width: 60px;
}

/* Content */
.layout-content {
  flex: 1;
  overflow-y: auto;
  padding: 1em;
  background: var(--background);
}

/* RTL Support */
html[dir="rtl"] .layout-sidebar {
  border-right: none;
  border-left: 1px solid var(--divider);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .layout-sidebar {
    width: 160px;
  }
  
  .layout-sidebar.collapsed {
    width: 60px;
  }
}

/* Responsive - Mobile */
@media (max-width: 736px) {
  .layout-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    width: 300px !important;
    transform: translateX(-100%);
    z-index: 1001;
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  
  .layout-sidebar.collapsed {
    width: 300px !important;
  }
  
  .layout-sidebar.active {
    transform: translateX(0);
  }
  
  html[dir="rtl"] .layout-sidebar {
    left: auto;
    right: 0;
    transform: translateX(100%);
  }
  
  html[dir="rtl"] .layout-sidebar.active {
    transform: translateX(0);
  }
}
</style>