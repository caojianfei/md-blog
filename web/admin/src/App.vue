<script setup>
import { onMounted } from 'vue';

onMounted(() => {
  window.addEventListener('unauthorized', () => {
    // If the request util emits unauthorized event, reload page to go to login
    window.location.reload();
  });
});
</script>

<template>
  <router-view />
</template>
