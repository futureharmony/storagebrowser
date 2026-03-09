<template>
  <div class="keyboard-shortcuts">
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("keyboardShortcuts.title") }}</h2>
      </div>
      
      <div class="card-content">
        <div class="shortcuts-grid">
          <div class="shortcut-category">
            <h3>{{ $t("keyboardShortcuts.navigation") }}</h3>
            <div class="shortcut-item" v-for="item in navigationShortcuts" :key="item.key">
              <span class="shortcut-key">{{ item.key }}</span>
              <span class="shortcut-description">{{ item.description }}</span>
            </div>
          </div>
          
          <div class="shortcut-category">
            <h3>{{ $t("keyboardShortcuts.fileOperations") }}</h3>
            <div class="shortcut-item" v-for="item in fileOperationsShortcuts" :key="item.key">
              <span class="shortcut-key">{{ item.key }}</span>
              <span class="shortcut-description">{{ item.description }}</span>
            </div>
          </div>
          
          <div class="shortcut-category">
            <h3>{{ $t("keyboardShortcuts.selection") }}</h3>
            <div class="shortcut-item" v-for="item in selectionShortcuts" :key="item.key">
              <span class="shortcut-key">{{ item.key }}</span>
              <span class="shortcut-description">{{ item.description }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="card-action">
        <button
          class="button button--flat"
          @click="$router.back()"
          :aria-label="$t('buttons.close')"
          :title="$t('buttons.close')"
        >
          {{ $t("buttons.close") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const navigationShortcuts = computed(() => [
  { key: "F1", description: t("help.f1") },
  { key: "F2", description: t("help.f2") },
  { key: "ESC", description: t("help.esc") },
  { key: "Click", description: t("help.click") },
  { key: "Double click", description: t("help.doubleClick") },
]);

const fileOperationsShortcuts = computed(() => [
  { key: "DEL", description: t("help.del") },
  { key: "CTRL + S", description: t("help.ctrl.s") },
  { key: "CTRL + SHIFT + F", description: t("help.ctrl.f") },
]);

const selectionShortcuts = computed(() => [
  { key: "CTRL + Click", description: t("help.ctrl.click") },
  { key: "CTRL + A", description: t("keyboardShortcuts.selectAll") },
  { key: "SHIFT + Click", description: t("keyboardShortcuts.rangeSelect") },
]);
</script>

<style scoped>
.keyboard-shortcuts {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 2rem;
}

.shortcuts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  width: 100%;
}

.shortcut-category {
  background: var(--surfacePrimary);
  border-radius: 8px;
  padding: 1.5rem;
  border: 1px solid var(--borderPrimary);
}

.shortcut-category h3 {
  margin: 0 0 1rem 0;
  color: var(--textPrimary);
  font-size: 1.1rem;
  font-weight: 600;
}

.shortcut-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--borderSecondary);
}

.shortcut-item:last-child {
  border-bottom: none;
}

.shortcut-key {
  background: var(--surfaceSecondary);
  color: var(--blue);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  font-weight: bold;
  border: 1px solid var(--borderPrimary);
  min-width: 60px;
  text-align: center;
}

.shortcut-description {
  color: var(--textPrimary);
  font-size: 0.95rem;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .keyboard-shortcuts {
    padding: 1rem;
  }
  
  .shortcuts-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .shortcut-category {
    padding: 1rem;
  }
  
  .shortcut-key {
    font-size: 0.85rem;
    min-width: 50px;
  }
}
</style>