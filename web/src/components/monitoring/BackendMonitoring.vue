<template>
  <div class="grid gap-6 mt-4">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Active Matches</div>
        <div class="text-3xl mt-2">{{ metrics.activeMatches }}</div>
      </div>
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Avg CPU Usage</div>
        <div class="text-3xl mt-2">{{ metrics.avgCpu }}%</div>
      </div>
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Avg Memory Usage</div>
        <div class="text-3xl mt-2">{{ metrics.avgMemory }} MB</div>
      </div>
    </div>

    <div class="bg-white p-4 rounded shadow">
      <div class="text-xl font-semibold mb-2">📈 Resource Utilization</div>
      <LineChart :chart-data="usageChartData" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from "vue";
import LineChart from "./LineChart.vue";

const props = defineProps({
  backendId: String,
});

const metrics = ref({
  activeMatches: 0,
  avgCpu: 0,
  avgMemory: 0,
  usageHistory: [],
});

const usageChartData = computed(() => ({
  labels: metrics.value.usageHistory.map((d) => d.timestamp),
  datasets: [
    {
      label: "CPU",
      data: metrics.value.usageHistory.map((d) => d.cpu),
      fill: false,
      borderColor: "#3b82f6",
      tension: 0.4,
    },
    {
      label: "Memory",
      data: metrics.value.usageHistory.map((d) => d.memory),
      fill: false,
      borderColor: "#10b981",
      tension: 0.4,
    },
  ],
}));

let interval;

function fetchMetrics() {
  fetch(`/mock/monitoring/backend.json`)
    .then((res) => res.json())
    .then((data) => {
      metrics.value = data;
    })
    .catch(console.error);
}

onMounted(() => {
  fetchMetrics();
  interval = setInterval(fetchMetrics, 30000);
});

onUnmounted(() => {
  clearInterval(interval);
});
</script>
