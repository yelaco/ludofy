<template>
  <div class="grid gap-6 mt-4">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Active Matches</div>
        <div class="text-3xl mt-2">{{ metrics.activeMatches }}</div>
      </div>
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Avg CPU Usage</div>
        <div class="text-3xl mt-2">{{ metrics.avgCpu }} %</div>
      </div>
      <div class="bg-white p-4 rounded shadow">
        <div class="text-xl font-semibold">Avg Memory Usage</div>
        <div class="text-3xl mt-2">{{ metrics.avgMemory }} %</div>
      </div>
    </div>

    <div class="bg-white p-4 rounded shadow relative">
      <div class="flex justify-between items-start mb-2">
        <div class="text-xl font-semibold">📈 Resource Utilization</div>
        <TimeRangePicker @update="updateTimeRange" />
      </div>

      <LineChart :chart-data="usageChartData" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { userManager } from "@/auth";
import LineChart from "./LineChart.vue";
import TimeRangePicker from "@/components/TimeRangePicker.vue";

const showPicker = ref(false);

const props = defineProps({
  metricsEndpointUrl: String,
});

const metrics = ref({
  activeMatches: 0,
  avgCpu: 0,
  avgMemory: 0,
  usageHistory: [],
});

const startTime = ref(null);
const endTime = ref(null);
const autoRefresh = ref(false);
let refreshInterval = null;

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

function getAutoInterval(startISO, endISO) {
  const start = new Date(startISO);
  const end = new Date(endISO);
  const diffSec = (end - start) / 1000;
  let rawInterval = Math.ceil(diffSec / 10); // 10 data points

  // Enforce AWS CloudWatch rules
  if (diffSec <= 3 * 3600) {
    // max granularity: 60s
    return Math.max(60, Math.ceil(rawInterval / 60) * 60);
  } else if (diffSec <= 15 * 24 * 3600) {
    // 1 to 5 min intervals
    return Math.max(300, Math.ceil(rawInterval / 60) * 60);
  } else if (diffSec <= 63 * 24 * 3600) {
    return Math.max(3600, Math.ceil(rawInterval / 3600) * 3600);
  } else {
    return Math.max(86400, Math.ceil(rawInterval / 86400) * 86400);
  }
}

async function fetchMetrics() {
  if (!props.metricsEndpointUrl || !startTime.value || !endTime.value) return;

  const url = new URL(props.metricsEndpointUrl);
  url.searchParams.set("start", startTime.value);
  url.searchParams.set("end", endTime.value);

  const period = getAutoInterval(startTime.value, endTime.value);
  url.searchParams.set("interval", period);

  try {
    const user = await userManager.getUser();

    const res = await fetch(url.toString(), {
      headers: {
        Authorization: `Bearer ${user.id_token}`,
      },
    });

    const data = await res.json();

    const usageHistory = data.serviceMetricsList
      .slice()
      .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
      .map((d) => ({
        timestamp: new Date(d.timestamp).toLocaleTimeString("en-GB", {
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        }),
        cpu: d.cpuAvg,
        memory: d.memAvg,
      }));

    const avgCpu =
      usageHistory.reduce((acc, d) => acc + d.cpu, 0) / usageHistory.length;
    const avgMemory =
      usageHistory.reduce((acc, d) => acc + d.memory, 0) / usageHistory.length;

    metrics.value = {
      usageHistory,
      avgCpu: avgCpu.toFixed(2),
      avgMemory: avgMemory.toFixed(2),
      activeMatches: data.serverMetricsList?.[0]?.activeMatches ?? 0,
    };
  } catch (err) {
    console.error("Failed to fetch metrics:", err);
  }
}

function updateTimeRange({ start, end, autoRefresh: auto }) {
  startTime.value = start;
  endTime.value = end;
  autoRefresh.value = auto;

  fetchMetrics();

  if (refreshInterval) clearInterval(refreshInterval);
  if (auto) {
    refreshInterval = setInterval(() => {
      fetchMetrics();
    }, 30000);
  }
}

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval);
});
</script>
