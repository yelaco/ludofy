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
import { userManager } from "@/auth"; // Assuming you already have this
import LineChart from "./LineChart.vue";

const props = defineProps({
  metricsEndpointUrl: String,
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

function getTimeRange(minutesAgo = 30) {
  const end = new Date();
  const start = new Date(end.getTime() - minutesAgo * 60 * 1000);
  return {
    start: start.toISOString(),
    end: end.toISOString(),
  };
}

let interval;

async function fetchMetrics() {
  if (!props.metricsEndpointUrl) {
    console.warn("No metricsEndpointUrl provided");
    return;
  }

  const { start, end } = getTimeRange(30);
  interval = 300;

  const url = new URL(props.metricsEndpointUrl);
  url.searchParams.set("start", start);
  url.searchParams.set("end", end);
  url.searchParams.set("interval", interval);

  fetch(url.toString())
    .then((res) => res.json())
    .then((data) => {
      console.log(data);
      const usageHistory = data.serviceMetricsList
        .slice()
        .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        .map((d) => ({
          timestamp: new Date(d.timestamp).toLocaleTimeString("en-GB", {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          }),
          cpu: d.cpuAvg * 100,
          memory: d.memAvg * 1024,
        }));

      const avgCpu =
        usageHistory.reduce((acc, d) => acc + d.cpu, 0) / usageHistory.length;
      const avgMemory =
        usageHistory.reduce((acc, d) => acc + d.memory, 0) /
        usageHistory.length;

      metrics.value = {
        usageHistory,
        avgCpu: avgCpu.toFixed(2),
        avgMemory: avgMemory.toFixed(2),
        activeMatches: data.serverMetricsList?.[0]?.activeMatches ?? 0,
      };
    })
    .catch((err) => {
      console.error("Failed to fetch metrics:", err);
    });
}

onMounted(() => {
  fetchMetrics();
});

onUnmounted(() => {
  clearInterval(interval);
});
</script>
