<template>
  <div class="p-6 max-w-6xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">🔎 Backend Detail</h1>

    <div v-if="loading" class="text-gray-500">Loading backend details...</div>

    <div v-else-if="error" class="text-red-500">{{ error }}</div>

    <div v-else class="space-y-8">
      <!-- Basic backend info -->
      <div class="bg-white p-6 rounded shadow space-y-4">
        <div class="flex items-start justify-between">
          <div class="space-y-2">
            <div><strong>ID:</strong> {{ backend.id }}</div>
            <div><strong>Stack Name:</strong> {{ backend.stackName }}</div>
            <div
              v-if="
                backend.createdAt &&
                backend.createdAt !== '0001-01-01T00:00:00Z'
              "
            >
              <strong>Created At:</strong> {{ formatDate(backend.createdAt) }}
            </div>
            <div v-else><strong>Created At:</strong> N/A</div>
          </div>

          <!-- Right side buttons (Update + Delete) -->
          <div class="flex flex-col items-end gap-3 ml-4">
            <button
              @click="goToUpdatePage"
              class="w-20 px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700"
            >
              Update
            </button>

            <button
              @click="openDeleteDialog(backend.id, backend.stackName)"
              class="w-20 px-4 py-2 bg-red-500 text-white text-sm rounded-md hover:bg-red-600"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      <!-- Monitoring Dashboard -->
      <BackendMonitoring :backend-id="backend.id" />

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
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import BackendMonitoring from "@/components/monitoring/BackendMonitoring.vue";
import api from "@/api";

const route = useRoute();
const router = useRouter();
const id = route.params.id;

const backend = ref(null);
const loading = ref(true);
const error = ref(null);

const showOutputs = ref(false); // ✨ Control expand/collapse
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
  const date = new Date(isoString);
  return date.toLocaleString();
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
    setTimeout(() => router.push("/backends"), 1000);
  } catch (error) {
    console.error("Failed to request backend deletion", error);
    showToast("Failed to request deletion. Please try again.");
  }
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

function goToUpdatePage() {
  router.push(`/backend/${id}/update`);
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
