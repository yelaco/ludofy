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
      chartInstance.data = props.chartData;
      chartInstance.update();
    }
  },
  { deep: true },
);

function createChart() {
  if (!canvas.value) return;

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
          max: 100,
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
