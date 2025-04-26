<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">🗂️ Deployed Backends</h1>

    <div v-if="loading" class="flex justify-center items-center py-12">
      <svg
        class="animate-spin h-8 w-8 text-blue-500"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        ></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8v8z"
        ></path>
      </svg>
    </div>

    <div v-else>
      <div
        v-if="backends.length === 0 && !hasPreviousPage"
        class="text-center text-gray-500 py-12"
      >
        <div class="text-5xl mb-4">📦</div>
        <p>No backends deployed yet.</p>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="backend in backends"
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
              @click="openDeleteDialog(backend.id, backend.stackName)"
            >
              🗑 Delete
            </button>
          </div>
        </div>

        <!-- Pagination Toolbar -->
        <div
          class="flex flex-col md:flex-row justify-between items-center gap-4 mt-6 text-sm"
        >
          <!-- Page Size -->
          <div class="flex items-center gap-2">
            <span class="text-gray-600">Per page:</span>
            <select
              v-model.number="pageSize"
              class="border rounded p-1 text-sm"
            >
              <option :value="5">5</option>
              <option :value="10">10</option>
              <option :value="20">20</option>
            </select>
          </div>

          <!-- Prev / Next / Refresh -->
          <div class="flex items-center gap-3">
            <button
              class="px-3 py-1 bg-gray-200 rounded hover:bg-gray-300"
              :disabled="!hasPreviousPage"
              @click="prevPage"
            >
              ◀️ Prev
            </button>

            <button
              class="px-3 py-1 bg-gray-200 rounded hover:bg-gray-300"
              :disabled="!nextPageToken"
              @click="nextPage"
            >
              Next ▶️
            </button>

            <button
              class="px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 flex items-center gap-2"
              @click="refreshBackends"
              :disabled="loading"
            >
              <svg
                v-if="loading"
                class="animate-spin h-4 w-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10"></circle>
                <path class="opacity-75" d="M4 12a8 8 0 018-8v8z"></path>
              </svg>
              <span v-else>🔄</span>
              Refresh
            </button>
          </div>

          <!-- Total Results -->
          <div class="text-gray-500 text-sm">
            {{ backends.length }} result{{ backends.length !== 1 ? "s" : "" }}
            loaded
          </div>
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
        <p class="text-gray-700 mb-2">
          Stack name: <strong>{{ stackNameToDelete }}</strong>
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
import { ref, onMounted, watch, computed } from "vue";
import api from "../api.js";

const backends = ref([]);
const cachedPages = ref([]);
const nextPageToken = ref(null);
const pageSize = ref(5);
const loading = ref(false);

const hasPreviousPage = computed(() => cachedPages.value.length > 0);

const showDeleteDialog = ref(false);
const backendIdToDelete = ref(null);
const stackNameToDelete = ref(null);
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

async function loadBackends(startKey = null, useCache = false) {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("limit", pageSize.value);
    if (startKey) {
      params.set("startKey", JSON.stringify(startKey));
    }

    const response = await api.getBackends(`?${params.toString()}`);

    if (!useCache && backends.value.length > 0) {
      cachedPages.value.push({
        items: backends.value,
        nextPageToken: nextPageToken.value,
      });
    }

    backends.value = response.data.items;
    nextPageToken.value = response.data.nextPageToken || null;
  } finally {
    loading.value = false;
  }
}

async function nextPage() {
  if (!nextPageToken.value) return;
  await loadBackends(nextPageToken.value);
}

async function prevPage() {
  if (cachedPages.value.length === 0) return;

  const previousPage = cachedPages.value.pop();
  backends.value = previousPage.items;
  nextPageToken.value = previousPage.nextPageToken;
}

async function refreshBackends() {
  cachedPages.value = [];
  await loadBackends();
}

function openDeleteDialog(id, stackName) {
  backendIdToDelete.value = id;
  stackNameToDelete.value = stackName;
  confirmationInput.value = "";
  showDeleteDialog.value = true;
}

function closeDeleteDialog() {
  backendIdToDelete.value = null;
  stackNameToDelete.value = null;
  confirmationInput.value = "";
  showDeleteDialog.value = false;
}

async function confirmDeleteBackend() {
  if (!backendIdToDelete.value) return;

  try {
    await api.removeBackend(backendIdToDelete.value);
    closeDeleteDialog();
    showToast("Backend deletion initiated.");
    await refreshBackends();
  } catch (error) {
    console.error("Failed to request backend deletion", error);
    showToast("Failed to request deletion. Please try again.");
  }
}

watch(pageSize, async () => {
  cachedPages.value = [];
  await loadBackends();
});

onMounted(async () => {
  try {
    await loadBackends();
  } catch (error) {
    console.error("Failed to fetch backends", error);
  }
});
</script>

<style scoped>
/* Spinner already built-in, no extra CSS needed */
</style>
