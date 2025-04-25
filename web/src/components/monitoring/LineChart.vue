<template>
  <div>
    <canvas ref="canvas"></canvas>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import {
  Chart,
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  Title,
  CategoryScale,
  Tooltip,
  Legend,
} from "chart.js";

Chart.register(
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  Title,
  CategoryScale,
  Tooltip,
  Legend,
);

const props = defineProps({
  chartData: Object,
});

const canvas = ref(null);
let chartInstance = null;

onMounted(() => {
  createChart();
});

watch(
  () => props.chartData,
  () => {
    if (chartInstance) {
      const allValues = props.chartData.datasets.flatMap((ds) => ds.data);
      const maxValue = Math.max(...allValues);
      const suggestedMax = getSuggestedMax(maxValue);

      chartInstance.options.scales.y.max = suggestedMax;
      chartInstance.data = props.chartData;
      chartInstance.update();
    }
  },
  { deep: true },
);

function getSuggestedMax(value) {
  if (value <= 0) return 1;
  const exponent = Math.floor(Math.log10(value));
  const base = Math.pow(10, exponent);
  if (value <= base) return base;
  if (value <= 2 * base) return 2 * base;
  if (value <= 5 * base) return 5 * base;
  return 10 * base;
}

function createChart() {
  if (!canvas.value) return;

  const allValues = props.chartData.datasets.flatMap((ds) => ds.data);
  const maxValue = Math.max(...allValues);
  const suggestedMax = getSuggestedMax(maxValue);

  chartInstance = new Chart(canvas.value, {
    type: "line",
    data: props.chartData,
    options: {
      responsive: true,
      plugins: {
        legend: { display: true },
        tooltip: { mode: "index", intersect: false },
      },
      interaction: {
        mode: "nearest",
        axis: "x",
        intersect: false,
      },
      scales: {
        x: {
          display: true,
        },
        y: {
          display: true,
          beginAtZero: true,
          max: suggestedMax,
          title: {
            display: true,
            text: "Usage (%)",
          },
        },
      },
    },
  });
}
</script>
