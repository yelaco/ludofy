<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">🗂️ Deployed Backends</h1>

    <div v-if="backends.length === 0" class="text-gray-500">
      No backends deployed yet.
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="(backend, idx) in backends"
        :key="backend.id"
        class="border rounded-md p-4 shadow-sm bg-white"
      >
        <div class="flex justify-between items-center mb-2">
          <h2 class="text-lg font-semibold text-blue-800">
            {{ backend.stackName }}
          </h2>
        </div>
        <p class="text-sm text-gray-500">Created: {{ backend.createdAt }}</p>

        <div class="flex gap-3 mt-4">
          <RouterLink
            :to="`/backend/${backend.id}`"
            class="text-blue-600 hover:underline text-sm"
            >🔧 View / Manage</RouterLink
          >
          <button
            class="text-red-500 text-sm hover:underline"
            @click="deleteBackend(backend.id)"
          >
            🗑 Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import api from "../api.js";

const backends = ref([]);

onMounted(async () => {
  try {
    const response = await api.getBackends();
    backends.value = response.data.items;
  } catch (error) {
    console.error("Failed to fetch backends", error);
  }
});
</script>
