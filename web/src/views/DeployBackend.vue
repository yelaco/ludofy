<template>
  <div class="p-6 max-w-4xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Deploy New Game Backend</h1>

    <!-- Stack Name Input -->
    <div class="mb-6">
      <label class="block text-sm font-medium">Stack Name</label>
      <input
        type="text"
        v-model="stackName"
        @input="validateField('stackName')"
        :class="['input', errors.stackName ? 'border-red-500' : '']"
        placeholder="Enter a stack name"
      />
      <p v-if="errors.stackName" class="text-red-500 text-xs mt-1">
        {{ errors.stackName }}
      </p>
    </div>

    <ServiceSelector v-model="services" />

    <!-- Matchmaking Configuration -->
    <div class="mb-8">
      <h2 class="text-xl font-semibold mb-2">🎯 Matchmaking Configuration</h2>
      <div class="space-y-4">
        <div class="mb-4">
          <label class="block text-sm font-medium">Players per match</label>
          <input
            type="number"
            v-model="matchmaking.matchSize"
            @input="validateNumberFields"
            :class="['input', errors.matchSize ? 'border-red-500' : '']"
            min="2"
          />
          <p v-if="errors.matchSize" class="text-red-500 text-xs mt-1">
            {{ errors.matchSize }}
          </p>
        </div>

        <template v-if="services.ranking">
          <div>
            <label class="block text-sm font-medium">Rating Algorithm</label>
            <select v-model="matchmaking.ratingAlgorithm" class="input">
              <option value="glicko">Glicko</option>
              <option value="elo">ELO</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium">Initial Rating</label>
            <input
              type="number"
              v-model="matchmaking.initialRating"
              class="input"
              min="0"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Server Configuration -->
    <div class="mb-8">
      <h2 class="text-xl font-semibold mb-2">🖥️ Server Configuration</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium">Server Image URI</label>
          <input
            type="text"
            v-model="serverImageUri"
            @input="validateField('serverImageUri')"
            :class="['input', errors.serverImageUri ? 'border-red-500' : '']"
            placeholder="e.g., 123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/mygame-server:latest"
          />
          <p v-if="errors.serverImageUri" class="text-red-500 text-xs mt-1">
            {{ errors.serverImageUri }}
          </p>
        </div>

        <div>
          <label class="flex items-center space-x-2">
            <input type="checkbox" v-model="privateRegistry" class="w-4 h-4" />
            <span class="text-sm font-medium">Private registry</span>
          </label>
        </div>

        <template v-if="privateRegistry">
          <div class="space-y-4 mt-4">
            <div>
              <label class="block text-sm font-medium">Registry Username</label>
              <input
                type="text"
                v-model="registryCredentials.username"
                @input="validateField('registryUsername')"
                :class="[
                  'input',
                  errors.registryUsername ? 'border-red-500' : '',
                ]"
                placeholder="Username for private registry"
              />
              <p
                v-if="errors.registryUsername"
                class="text-red-500 text-xs mt-1"
              >
                {{ errors.registryUsername }}
              </p>
            </div>
            <div class="relative">
              <label class="block text-sm font-medium">Registry Password</label>
              <input
                :type="showPassword ? 'text' : 'password'"
                v-model="registryCredentials.password"
                @input="validateField('registryPassword')"
                :class="[
                  'input pr-12',
                  errors.registryPassword ? 'border-red-500' : '',
                ]"
                placeholder="Password or access token"
              />
              <button
                type="button"
                class="absolute top-9 right-3 text-gray-500 hover:text-gray-700"
                @click="showPassword = !showPassword"
              >
                <component :is="showPassword ? EyeOff : Eye" class="w-5 h-5" />
              </button>
              <p
                v-if="errors.registryPassword"
                class="text-red-500 text-xs mt-1"
              >
                {{ errors.registryPassword }}
              </p>
            </div>
          </div>
        </template>

        <div class="mb-4">
          <label class="block text-sm font-medium"
            >Max Concurrent Matches (Per Server)</label
          >
          <input
            type="number"
            v-model="server.maxMatches"
            @input="validateNumberFields"
            :class="['input', errors.maxMatches ? 'border-red-500' : '']"
            min="2"
            placeholder="100"
          />
          <p v-if="errors.maxMatches" class="text-red-500 text-xs mt-1">
            {{ errors.maxMatches }}
          </p>
        </div>
        <div>
          <label class="block text-sm font-medium">Processor (vCPU)</label>
          <select v-model="server.cpu" class="input">
            <option :value="0.25">0.25</option>
            <option :value="0.5">0.5</option>
            <option :value="1">1</option>
            <option :value="2">2</option>
            <option :value="4">4</option>
            <option :value="8">8</option>
            <option :value="16">16</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium">Memory (GB)</label>
          <select v-model="server.memory" class="input">
            <option
              v-for="memory in allowedMemoryOptions"
              :key="memory"
              :value="memory"
            >
              {{ memory / 1024 }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <hr class="my-8" />

    <button
      @click="submit"
      class="w-full px-6 py-3 bg-green-600 text-white font-semibold rounded-lg text-lg hover:bg-green-700 shadow-md"
    >
      🚀 Deploy
    </button>
  </div>

  <!-- Toast Notification -->
  <div
    v-if="toastMessage"
    class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 px-6 py-3 rounded-md shadow-lg text-sm"
    :class="
      toastType === 'success'
        ? 'bg-green-100 text-green-800'
        : 'bg-red-100 text-red-700'
    "
  >
    {{ toastMessage }}
  </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { Eye, EyeOff } from "lucide-vue-next";
import ServiceSelector from "@/components/ServiceSelector.vue";
import api from "@/api";
import { useRouter } from "vue-router";

const router = useRouter();

const stackName = ref("");
const services = ref({
  chat: false,
  friend: false,
  ranking: false,
  matchSpectating: false,
});

const matchmaking = ref({
  matchSize: 2,
  ratingAlgorithm: "glicko",
  initialRating: 400,
});

const server = ref({
  cpu: 0.5,
  memory: 1024,
  maxMatches: 100,
});
const serverImageUri = ref("");

const privateRegistry = ref(false);
const registryCredentials = ref({ username: "", password: "" });

const toastMessage = ref("");
const toastType = ref("success");
const errors = ref({
  stackName: "",
  serverImageUri: "",
  registryUsername: "",
  registryPassword: "",
  matchSize: "",
  maxMatches: "",
});
const showPassword = ref(false);

const numberValidationRules = [
  {
    field: matchmaking,
    key: "matchSize",
    min: 2,
    errorKey: "matchSize",
    message: "Players per match must be at least 2.",
  },
  {
    field: server,
    key: "maxMatches",
    min: 2,
    errorKey: "maxMatches",
    message: "Max concurrent matches must be at least 2.",
  },
];

const allowedMemoryOptions = computed(() => {
  const cpu = server.value.cpu;
  if (cpu === 0.25) return [512, 1024, 2048];
  if (cpu === 0.5) return [1024, 2048, 3072, 4096];
  if (cpu === 1) return [2048, 3072, 4096, 5120, 6144, 7168, 8192];
  if (cpu === 2) return Array.from({ length: 13 }, (_, i) => (4 + i) * 1024);
  if (cpu === 4) return Array.from({ length: 23 }, (_, i) => (8 + i) * 1024);
  if (cpu === 8)
    return Array.from({ length: 12 }, (_, i) => (16 + i * 4) * 1024);
  if (cpu === 16)
    return Array.from({ length: 12 }, (_, i) => (32 + i * 8) * 1024);
  return [];
});

watch(
  () => server.value.cpu,
  () => {
    if (!allowedMemoryOptions.value.includes(server.value.memory)) {
      server.value.memory = allowedMemoryOptions.value[0];
    }
  },
);

function validateField(field) {
  if (field === "stackName") {
    errors.value.stackName = stackName.value.trim()
      ? ""
      : "Stack name is required.";
  } else if (field === "serverImageUri") {
    errors.value.serverImageUri = serverImageUri.value.trim()
      ? ""
      : "Server Image URI is required.";
  } else if (field === "registryUsername" && privateRegistry.value) {
    errors.value.registryUsername = registryCredentials.value.username.trim()
      ? ""
      : "Registry username is required.";
  } else if (field === "registryPassword" && privateRegistry.value) {
    errors.value.registryPassword = registryCredentials.value.password.trim()
      ? ""
      : "Registry password is required.";
  }
}

function validateNumberFields() {
  let isValid = true;

  for (const rule of numberValidationRules) {
    const value = rule.field.value[rule.key];
    if (value < rule.min) {
      errors.value[rule.errorKey] = rule.message;
      isValid = false;
    } else {
      errors.value[rule.errorKey] = "";
    }
  }

  return isValid;
}

async function submit() {
  validateField("stackName");
  validateField("serverImageUri");

  if (privateRegistry.value) {
    validateField("registryUsername");
    validateField("registryPassword");
  }

  if (!validateNumberFields()) {
    return;
  }

  if (errors.value.stackName || errors.value.serverImageUri) {
    return;
  }

  try {
    const deployInput = {
      stackName: stackName.value,
      includeChatService: services.value.chat,
      includeFriendService: services.value.friend,
      includeRankingService: services.value.ranking,
      includeMatchSpectatingService: services.value.matchSpectating,
      matchmakingConfiguration: {
        matchSize: matchmaking.value.matchSize,
        ratingAlgorithm: matchmaking.value.ratingAlgorithm,
        initialRating: matchmaking.value.initialRating,
      },
      serverConfiguration: {
        containerImage: {
          uri: serverImageUri.value,
          isPrivate: privateRegistry.value,
          registryCredentials: registryCredentials.value,
        },
        maxMatches: server.maxMatches,
        initialCpu: server.value.cpu,
        initialMemory: server.value.memory,
      },
    };

    await api.deployBackend(deployInput);
    showToast("Deployment started successfully!", "success");
    setTimeout(() => router.push("/deployments"), 1000);
  } catch (error) {
    console.error("Failed to deploy backend", error);
    if (error.response?.status === 302) {
      showToast("Duplicate stack name. Please use another.", "error");
    } else if (error.response?.status === 409) {
      showToast("Pending deployment exists. Try again later.", "error");
    } else {
      showToast("Failed to deploy backend. Please try again.", "error");
    }
  }
}

function showToast(message, type = "success", duration = 3000) {
  toastMessage.value = message;
  toastType.value = type;
  setTimeout(() => {
    toastMessage.value = "";
  }, duration);
}
</script>

<style scoped>
.input {
  width: 100%;
  padding: 0.5rem 1rem;
  border-width: 1px;
  border-radius: 0.5rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  outline: none;
}
.input:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.5);
}
</style>
