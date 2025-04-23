<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">📦 Deployment History</h1>

    <div v-if="deployments.length === 0" class="text-gray-500">
      No deployments yet.
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="(deployment, idx) in deployments"
        :key="deployment.id"
        class="border rounded-md p-4 shadow-sm bg-white"
      >
        <!-- Header (always shown) -->
        <div class="flex justify-between items-center">
          <div class="flex items-center gap-2">
            <button @click="toggleExpand(idx)" class="focus:outline-none">
              <span v-if="expanded[idx]">🔽</span>
              <span v-else>▶️</span>
            </button>
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

        <!-- Details (only when expanded) -->
        <div v-if="expanded[idx]" class="mt-4 space-y-3 text-sm text-gray-700">
          <p>
            <strong>🪪 Deployment ID: </strong>
            <span class="font-mono">{{ deployment.id }}</span>
          </p>
          <p>
            🐳 <strong>Server Image: </strong>
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
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import api from "../api.js";

const deployments = ref([]);
const expanded = ref({});

function toggleExpand(idx) {
  expanded.value[idx] = !expanded.value[idx];
}

function formatDate(isoString) {
  if (!isoString) return "Unknown date";
  const date = new Date(isoString);
  return date.toLocaleString(); // "4/20/2025, 10:30:00 PM" (localized to user's timezone)
}

onMounted(async () => {
  try {
    const response = await api.getDeployments();
    deployments.value = response.data.items;

    // Initialize all deployments collapsed
    deployments.value.forEach((_, idx) => {
      expanded.value[idx] = false;
    });
  } catch (error) {
    console.error("Failed to fetch deployments", error);
  }
});
</script>
