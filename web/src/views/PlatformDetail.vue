<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold mb-4">Platform: {{ platformId }}</h1>
    <p class="text-gray-600">Platform details and management UI goes here.</p>
  </div>
</template>

<script setup>
import { useRoute } from "vue-router";

const route = useRoute();
const platformId = route.params.id;
</script>
