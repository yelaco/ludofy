<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold mb-4">Backend: {{ stackName }}</h1>
    <p class="text-gray-600">Backend details and management UI goes here.</p>
  </div>
</template>

<script setup>
import { useRoute } from "vue-router";

const route = useRoute();
const stackName = route.params.stackName;
</script>
