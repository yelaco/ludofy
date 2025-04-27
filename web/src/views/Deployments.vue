<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">📦 Deployment History</h1>

    <!-- Loading Spinner -->
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
        />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
      </svg>
    </div>

    <!-- No Data -->
    <div
      v-else-if="deployments.length === 0 && !hasPreviousPage"
      class="text-center text-gray-500 py-12"
    >
      <div class="text-5xl mb-4">📦</div>
      <p>No deployments yet.</p>
    </div>

    <!-- Deployment List -->
    <div v-else class="space-y-4">
      <div
        v-for="(deployment, idx) in deployments"
        :key="deployment.id"
        class="border rounded-md p-4 shadow-sm bg-white cursor-pointer"
        @click="toggleExpand(idx)"
      >
        <div class="flex justify-between items-center">
          <div class="flex items-center gap-2">
            <span>
              <span v-if="expanded[idx]">🔽</span>
              <span v-else>▶️</span>
            </span>
            <h2 class="text-lg font-semibold text-blue-800">
              {{ deployment.input.stackName }}
            </h2>
            <p class="text-xs text-gray-500">
              🕒 {{ formatDate(deployment.createdAt) }}
            </p>
          </div>

          <div class="flex items-center gap-4">
            <span
              class="px-2 py-1 text-sm rounded"
              :class="{
                'bg-green-100 text-green-700':
                  deployment.status === 'successful',
                'bg-yellow-100 text-yellow-700':
                  deployment.status === 'pending',
                'bg-sky-100 text-sky-700': deployment.status === 'deploying',
                'bg-red-100 text-red-700': deployment.status === 'failed',
              }"
            >
              {{ deployment.status }}
            </span>

            <RouterLink
              :to="`/backend/${deployment.backendId}`"
              v-if="deployment.status === 'successful'"
              class="text-blue-600 hover:underline text-sm"
            >
              🔧 View / Manage
            </RouterLink>
          </div>
        </div>

        <!-- Expanded Details -->
        <div v-if="expanded[idx]" class="mt-4 space-y-3 text-sm text-gray-700">
          <p>
            <strong>🪪 Deployment ID:</strong>
            <span class="font-mono">{{ deployment.id }}</span>
          </p>
          <p>
            🐳 <strong>Server Image:</strong>
            <span class="font-mono">{{
              deployment.input.serverConfiguration.containerImage.uri
            }}</span>
          </p>

          <div>
            🛠️ <strong>Services Enabled:</strong>
            <ul class="list-disc list-inside ml-4">
              <li v-if="deployment.input.includeChatService">Chat Service</li>
              <li v-if="deployment.input.includeFriendService">
                Friend Service
              </li>
              <li v-if="deployment.input.includeRankingService">
                Ranking Service
              </li>
              <li v-if="deployment.input.includeMatchSpectatingService">
                Match Spectating Service
              </li>
              <li
                v-if="
                  !deployment.input.includeChatService &&
                  !deployment.input.includeFriendService &&
                  !deployment.input.includeRankingService &&
                  !deployment.input.includeMatchSpectatingService
                "
                class="text-gray-400"
              >
                No additional services
              </li>
            </ul>
          </div>

          <p class="flex items-center gap-2">
            <span
              v-if="deployment.input.useCustomization"
              class="text-gray-600 flex items-center gap-1"
            >
              ✅ <strong>Customization: </strong>Enabled
            </span>
            <span v-else class="text-gray-600 flex items-center gap-1">
              🚫 <strong>Customization: </strong>Disabled
            </span>
          </p>

          <div>
            🎯 <strong>Matchmaking Configuration:</strong>
            <ul class="list-disc list-inside ml-4">
              <li>
                Match Size:
                {{ deployment.input.matchmakingConfiguration.matchSize }}
              </li>
              <li>
                Rating Algorithm:
                {{ deployment.input.matchmakingConfiguration.ratingAlgorithm }}
              </li>
              <li>
                Initial Rating:
                {{ deployment.input.matchmakingConfiguration.initialRating }}
              </li>
            </ul>
          </div>

          <div>
            ⚙️ <strong>Server Configuration:</strong>
            <ul class="list-disc list-inside ml-4">
              <li>
                Max Matches:
                {{ deployment.input.serverConfiguration.maxMatches }}
              </li>
              <li>
                CPU: {{ deployment.input.serverConfiguration.initialCpu }} vCPU
              </li>
              <li>
                Memory:
                {{ deployment.input.serverConfiguration.initialMemory / 1024 }}
                GB
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- Pagination Toolbar -->
      <div
        class="flex flex-col md:flex-row justify-between items-center gap-4 mt-6 text-sm"
      >
        <div class="flex items-center gap-2">
          <span class="text-gray-600">Per page:</span>
          <select v-model.number="pageSize" class="border rounded p-1 text-sm">
            <option :value="5">5</option>
            <option :value="10">10</option>
            <option :value="20">20</option>
          </select>
        </div>

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
            :disabled="!nextPageToken || loading"
            @click="nextPage"
          >
            Next ▶️
          </button>

          <button
            class="px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
            @click="refreshDeployments"
            :disabled="loading"
          >
            🔄 Refresh
          </button>
        </div>

        <div class="text-gray-500 text-sm">
          {{ deployments.length }} result{{
            deployments.length !== 1 ? "s" : ""
          }}
          loaded
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from "vue";
import api from "../api.js";

const deployments = ref([]);
const cachedPages = ref([]);
const nextPageToken = ref(null);
const pageSize = ref(5);
const loading = ref(false);

const expanded = ref({});
const hasPreviousPage = computed(() => cachedPages.value.length > 0);

function toggleExpand(idx) {
  expanded.value[idx] = !expanded.value[idx];
}

function formatDate(isoString) {
  if (!isoString) return "Unknown date";
  const date = new Date(isoString);
  return date.toLocaleString();
}

async function loadDeployments(startKey = null, useCache = false) {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("limit", pageSize.value);
    if (startKey) {
      params.set("startKey", JSON.stringify(startKey));
    }

    const response = await api.getDeployments(`?${params.toString()}`);

    if (!useCache && deployments.value.length > 0) {
      cachedPages.value.push({
        items: deployments.value,
        nextPageToken: nextPageToken.value,
      });
    }

    deployments.value = response.data.items;
    nextPageToken.value = response.data.nextPageToken || null;

    // Reset expanded on new load
    expanded.value = {};
    deployments.value.forEach((_, idx) => {
      expanded.value[idx] = false;
    });
  } finally {
    loading.value = false;
  }
}

async function nextPage() {
  if (!nextPageToken.value) return;
  await loadDeployments(nextPageToken.value);
}

async function prevPage() {
  if (cachedPages.value.length === 0) return;

  const previousPage = cachedPages.value.pop();
  deployments.value = previousPage.items;
  nextPageToken.value = previousPage.nextPageToken;

  // Rebuild expanded map for restored page
  expanded.value = {};
  deployments.value.forEach((_, idx) => {
    expanded.value[idx] = false;
  });
}

async function refreshDeployments() {
  cachedPages.value = [];
  await loadDeployments();
}

watch(pageSize, async () => {
  cachedPages.value = [];
  await loadDeployments();
});

onMounted(async () => {
  try {
    await loadDeployments();
  } catch (error) {
    console.error("Failed to fetch deployments", error);
  }
});
</script>

<style scoped>
/* Spinner already built-in */
</style>
