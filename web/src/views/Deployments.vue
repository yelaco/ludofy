<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">📦 Deployed Platforms</h1>

    <div v-if="platforms.length === 0" class="text-gray-500">
      No platforms deployed yet.
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="(platform, idx) in platforms"
        :key="platform.id"
        class="border rounded-md p-4 shadow-sm bg-white"
      >
        <div class="flex justify-between items-center mb-2">
          <h2 class="text-lg font-semibold text-blue-800">
            {{ platform.name }}
          </h2>
          <span
            class="px-2 py-1 text-sm rounded"
            :class="{
              'bg-green-100 text-green-700': platform.status === 'deployed',
              'bg-yellow-100 text-yellow-700': platform.status === 'pending',
              'bg-red-100 text-red-700': platform.status === 'failed',
            }"
          >
            {{ platform.status }}
          </span>
        </div>
        <p class="text-sm text-gray-500">Created: {{ platform.createdAt }}</p>
        <p class="text-sm mt-2">Services: {{ platform.services.join(", ") }}</p>
        <p class="text-sm mt-1">Games: {{ platform.games.length }} game(s)</p>

        <div class="flex gap-3 mt-4">
          <RouterLink
            :to="`/platform/${platform.id}`"
            class="text-blue-600 hover:underline text-sm"
            >🔧 View / Manage</RouterLink
          >
          <button
            class="text-red-500 text-sm hover:underline"
            @click="deletePlatform(platform.id)"
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

const platforms = ref([]);

onMounted(() => {
  // Placeholder data – replace with API fetch in real app
  platforms.value = [
    {
      id: "chessworld",
      name: "Chess World",
      status: "deployed",
      createdAt: "2025-04-10",
      services: ["authentication", "chat", "friend"],
      games: ["Chess"],
    },
    {
      id: "arcadehub",
      name: "Arcade Hub",
      status: "pending",
      createdAt: "2025-04-15",
      services: ["authentication"],
      games: ["Brick Breaker", "Tank Battle"],
    },
  ];
});

function deletePlatform(id) {
  platforms.value = platforms.value.filter((p) => p.id !== id);
}
</script>
