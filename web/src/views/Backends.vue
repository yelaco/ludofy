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
        <div class="flex justify-between items-center">
          <div class="flex items-center gap-2">
            <h2 class="text-lg font-semibold text-blue-800">
              {{ backend.stackName }}
            </h2>
          </div>
          <div class="flex items-center gap-4">
            <p class="text-xs text-gray-500">
              🕒 {{ formatDate(backend.updatedAt) }}
            </p>
            <span
              class="px-2 py-1 text-sm rounded"
              :class="{
                'bg-green-100 text-green-700': backend.status === 'active',
                'bg-yellow-100 text-yellow-700':
                  backend.status === 'delete-in-progress',
                'bg-red-100 text-red-700': backend.status === 'delete-failed',
              }"
            >
              {{ backend.status }}
            </span>
          </div>
        </div>

        <div class="flex gap-3 mt-4">
          <RouterLink
            :to="`/backend/${backend.id}`"
            class="text-blue-600 hover:underline text-sm"
          >
            🔧 View / Manage
          </RouterLink>

          <button
            class="text-red-500 text-sm hover:underline"
            @click="openDeleteDialog(backend.id)"
          >
            🗑 Delete
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div
      v-if="showDeleteDialog"
      class="fixed inset-0 flex items-center justify-center z-50"
    >
      <div class="bg-white border shadow-lg rounded-lg p-6 max-w-md w-full">
        <h2 class="text-xl font-bold text-red-600 mb-4">⚠️ Confirm Deletion</h2>
        <p class="text-gray-700 mb-2">
          This action is <strong>irreversible</strong>. Deleting this backend
          will permanently remove all associated resources.
        </p>
        <p class="text-gray-700 mb-4">
          Please type <strong>"permanently delete"</strong> below to confirm:
        </p>

        <input
          v-model="confirmationInput"
          type="text"
          placeholder="Type 'permanently delete'"
          class="border p-2 w-full mb-4 text-sm rounded"
        />

        <div class="flex justify-end gap-2">
          <button
            class="px-4 py-2 rounded text-sm text-gray-600 hover:underline"
            @click="closeDeleteDialog"
          >
            Cancel
          </button>
          <button
            :class="[
              'px-4 py-2 rounded text-sm',
              confirmationInput === 'permanently delete'
                ? 'text-white bg-red-500 hover:bg-red-600'
                : 'text-gray-400 bg-gray-200 cursor-not-allowed',
            ]"
            :disabled="confirmationInput !== 'permanently delete'"
            @click="confirmDeleteBackend"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>

    <!-- Notification Toast -->
    <div
      v-if="toastMessage"
      class="fixed top-4 left-1/2 transform -translate-x-1/2 bg-green-100 text-green-800 px-6 py-3 rounded-md shadow-lg z-50 text-xm animate-fade-in-out"
    >
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import api from "../api.js";

const backends = ref([]);

const showDeleteDialog = ref(false);
const backendIdToDelete = ref(null);
const confirmationInput = ref("");

const toastMessage = ref("");

function showToast(message, duration = 3000) {
  toastMessage.value = message;
  setTimeout(() => {
    toastMessage.value = "";
  }, duration);
}

function formatDate(isoString) {
  if (!isoString) return "Unknown date";
  const date = new Date(isoString);
  return date.toLocaleString();
}

function openDeleteDialog(id) {
  backendIdToDelete.value = id;
  confirmationInput.value = "";
  showDeleteDialog.value = true;
}

function closeDeleteDialog() {
  backendIdToDelete.value = null;
  confirmationInput.value = "";
  showDeleteDialog.value = false;
}

async function confirmDeleteBackend() {
  if (!backendIdToDelete.value) return;

  try {
    await api.removeBackend(backendIdToDelete.value);
    closeDeleteDialog();
    showToast("Backend deletion initiated.");
  } catch (error) {
    console.error("Failed to request backend deletion", error);
    showToast("Failed to request deletion. Please try again.");
  }
}

onMounted(async () => {
  try {
    const response = await api.getBackends();
    backends.value = response.data.items;
  } catch (error) {
    console.error("Failed to fetch backends", error);
  }
});
</script>
