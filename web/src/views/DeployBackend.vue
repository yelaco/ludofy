<template>
  <div class="p-6 max-w-4xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Deploy New Game Backend</h1>
    <BackendForm v-model="stackName" />
    <ServiceSelector v-model="services" />

    <hr class="my-8" />

    <button
      @click="deployBackend"
      class="w-full px-6 py-3 bg-green-600 text-white font-semibold rounded-lg text-lg hover:bg-green-700 shadow-md"
    >
      🚀 Deploy
    </button>
  </div>
</template>

<script setup>
import BackendForm from "@/components/BackendForm.vue";
import ServiceSelector from "@/components/ServiceSelector.vue";
import { ref } from "vue";

const stackName = ref("");
const services = ref({
  chat: false,
  friend: false,
  ranking: false,
  matchSpectating: false,
});

function deployBackend() {
  const data = {
    stackName: stackName.value,
    services: services.value,
  };
  console.log("Deploying...", data);
}
</script>
