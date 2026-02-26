<template>
  <div class="loading-spinner">
    <div class="spinner">
      <div class="bounce1"></div>
      <div class="bounce2"></div>
      <div class="bounce3"></div>
    </div>
    <p v-if="message">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message?: string;
}>();
</script>

<style scoped>
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  gap: 1rem;
}

.loading-spinner p {
  margin: 0;
  color: var(--text);
}

.spinner {
  display: block;
  text-align: center;
  line-height: 0;
}

.spinner > div {
  width: 0.8em;
  height: 0.8em;
  margin: 0 0.1em;
  font-size: 1.5em;
  background: var(--iconPrimary);
  border-radius: 100%;
  display: inline-block;
  animation: sk-bouncedelay 1.4s infinite ease-in-out both;
}

.spinner .bounce1 {
  animation-delay: -0.32s;
}

.spinner .bounce2 {
  animation-delay: -0.16s;
}

.spinner .bounce3 {
  animation-delay: 0s;
}

@keyframes sk-bouncedelay {
  0%,
  80%,
  100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}
</style>
