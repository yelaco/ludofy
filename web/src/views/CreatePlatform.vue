<template>
  <div class="p-6 max-w-4xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Create New Game Platform</h1>
    <PlatformForm v-model:platformName="platformName" />
    <ServiceSelector v-model="services" />

    <div class="space-y-6">
      <GameStack
        v-for="(game, index) in gameStacks"
        :key="index"
        :game="game"
        :index="index"
        :showRemove="gameStacks.length >= 2"
        @update:game="gameStacks[index] = $event"
        @remove="removeGameStack(index)"
        @upload="uploadDockerImage(index)"
      />
      <button
        @click="addGameStack"
        class="px-4 py-2 dark:bg-gray-800 dark:text-white rounded-md hover:bg-gray-400"
      >
        ➕ Add Game
      </button>
    </div>

    <hr class="my-8" />

    <button
      @click="deployPlatform"
      class="w-full px-6 py-3 bg-green-600 text-white font-semibold rounded-lg text-lg hover:bg-green-700 shadow-md"
    >
      🚀 Deploy Platform
    </button>
  </div>
</template>

<script setup>
import PlatformForm from "@/components/PlatformForm.vue";
import ServiceSelector from "@/components/ServiceSelector.vue";
import GameStack from "@/components/GameStack.vue";
import { ref } from "vue";

const platformName = ref("");
const services = ref({ chat: false, friend: false });
const gameStacks = ref([
  { name: "", dockerImage: "", ranking: false, spectating: false },
]);

function addGameStack() {
  gameStacks.value.push({
    name: "",
    dockerImage: "",
    ranking: false,
    spectating: false,
  });
}

function removeGameStack(index) {
  gameStacks.value.splice(index, 1);
}

function uploadDockerImage(index) {
  alert(`Upload Docker image for game #${index + 1}`);
}

function deployPlatform() {
  const data = {
    platformName: platformName.value,
    services: services.value,
    gameStacks: gameStacks.value,
  };
  console.log("Deploying...", data);
}
</script>
