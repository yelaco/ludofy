<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">📦 Deployed Backends</h1>

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
          <span
            class="px-2 py-1 text-sm rounded"
            :class="{
              'bg-green-100 text-green-700': backend.status === 'deployed',
              'bg-yellow-100 text-yellow-700': backend.status === 'pending',
              'bg-red-100 text-red-700': backend.status === 'failed',
            }"
          >
            {{ backend.status }}
          </span>
        </div>
        <p class="text-sm text-gray-500">Created: {{ backend.createdAt }}</p>
        <p class="text-sm mt-2">Services: {{ backend.services.join(", ") }}</p>

        <div class="flex gap-3 mt-4">
          <RouterLink
            :to="`/backend/${backend.stackName}`"
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

const backends = ref([]);

onMounted(() => {
  // Placeholder data – replace with API fetch in real app
  backends.value = [
    {
      id: "12332",
      stackName: "chessworld",
      status: "deployed",
      createdAt: "2025-04-10",
      services: ["authentication", "chat", "friend"],
    },
    {
      id: "13224",
      stackName: "tictactoe",
      status: "pending",
      createdAt: "2025-04-15",
      services: ["authentication"],
    },
  ];
});

function deleteBackend(id) {
  backends.value = backends.value.filter((p) => p.id !== id);
}
</script>
