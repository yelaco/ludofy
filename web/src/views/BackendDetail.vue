<template>
  <div class="p-6 max-w-6xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">🔎 Backend Detail</h1>

    <div v-if="loading" class="text-gray-500">Loading backend details...</div>

    <div v-else-if="error" class="text-red-500">{{ error }}</div>

    <div v-else class="space-y-8">
      <!-- Basic backend info -->
      <div class="bg-white p-6 rounded shadow space-y-4">
        <div><strong>ID:</strong> {{ backend.id }}</div>
        <div><strong>Stack Name:</strong> {{ backend.stackName }}</div>
        <div
          v-if="
            backend.createdAt && backend.createdAt !== '0001-01-01T00:00:00Z'
          "
        >
          <strong>Created At:</strong> {{ formatDate(backend.createdAt) }}
        </div>
        <div v-else><strong>Created At:</strong> N/A</div>
      </div>

      <!-- Outputs Section -->
      <div class="bg-white p-6 rounded shadow">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-semibold">🔗 API Endpoints</h2>
          <button
            @click="showOutputs = !showOutputs"
            class="text-blue-500 hover:underline focus:outline-none text-sm"
          >
            <span v-if="showOutputs">🔽 Collapse</span>
            <span v-else>▶️ Expand</span>
          </button>
        </div>

        <div v-if="showOutputs" class="space-y-4 transition-all duration-300">
          <div
            v-for="(value, key) in backend.outputs"
            :key="key"
            class="border p-4 rounded hover:bg-gray-50"
          >
            <div class="flex justify-between items-center">
              <div class="font-semibold text-gray-700 break-all">{{ key }}</div>
              <button
                @click="copyToClipboard(value)"
                class="text-blue-500 hover:underline text-sm"
              >
                📋 Copy
              </button>
            </div>

            <div class="mt-2">
              <template v-if="parseMethodAndUrl(value).method">
                <span
                  class="inline-block bg-gray-200 text-gray-800 text-xs px-2 py-1 rounded mr-2"
                >
                  {{ parseMethodAndUrl(value).method }}
                </span>
              </template>
              <span class="text-sm text-blue-600 break-all">
                {{ parseMethodAndUrl(value).url }}
              </span>
            </div>
          </div>
        </div>

        <div v-else class="text-gray-400 text-sm">
          (Click Expand to view all API endpoints)
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import api from "@/api";

const route = useRoute();
const id = route.params.id;

const backend = ref(null);
const loading = ref(true);
const error = ref(null);

const showOutputs = ref(false); // ✨ Control expand/collapse

function formatDate(isoString) {
  const date = new Date(isoString);
  return date.toLocaleString();
}

function parseMethodAndUrl(value) {
  if (!value) return { method: "", url: "" };
  const parts = value.split(" ");
  if (parts.length === 2) {
    return { method: parts[0], url: parts[1] };
  }
  return { method: "", url: value };
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text).then(() => {
    console.log("Copied to clipboard:", text);
  });
}

onMounted(async () => {
  try {
    const response = await api.getBackend(id);
    backend.value = response.data;
  } catch (err) {
    console.error("Failed to fetch backend", err);
    error.value = "Failed to load backend details.";
  } finally {
    loading.value = false;
  }
});
</script>
