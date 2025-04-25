<template>
  <div class="inline-block relative">
    <!-- Quick Presets + Calendar Toggle -->
    <div
      class="flex items-center flex-wrap gap-2 border rounded px-2 py-1 w-fit text-sm bg-white shadow-sm"
    >
      <button
        v-for="preset in quickPresets"
        :key="preset.label"
        @click="applyPreset(preset)"
        :class="[
          'px-2 py-1 rounded',
          relative.amount === preset.amount && relative.unit === preset.unit
            ? 'bg-blue-600 text-white'
            : 'hover:bg-gray-100',
        ]"
      >
        {{ preset.label }}
      </button>
      <button
        @click="showFullPicker = !showFullPicker"
        class="p-1 rounded hover:bg-gray-100"
        title="Custom Time Range"
      >
        📅
      </button>
    </div>

    <!-- Floating Picker Panel -->
    <div
      v-if="showFullPicker"
      class="absolute top-full right-0 z-20 mt-2 bg-white border rounded shadow-lg p-4 w-max"
    >
      <!-- Mode toggle -->
      <div class="flex gap-2 mb-2">
        <button
          class="px-3 py-1 rounded border"
          :class="isRelative ? 'bg-blue-600 text-white' : 'bg-white'"
          @click="isRelative = true"
        >
          Relative
        </button>
        <button
          class="px-3 py-1 rounded border"
          :class="!isRelative ? 'bg-blue-600 text-white' : 'bg-white'"
          @click="isRelative = false"
        >
          Absolute
        </button>
        <label class="flex items-center gap-1 text-sm ml-auto">
          <input type="checkbox" v-model="autoRefresh" /> Auto-refresh
        </label>
      </div>

      <!-- Relative full selector -->
      <div v-if="isRelative" class="flex items-center gap-2">
        <select v-model="relative.unit" class="border p-1 rounded">
          <option value="minutes">Minutes</option>
          <option value="hours">Hours</option>
          <option value="days">Days</option>
        </select>
        <input
          v-model.number="relative.amount"
          type="number"
          class="border p-1 w-20 rounded"
          min="1"
          max="9999"
        />
        <button
          @click="emitRange()"
          class="bg-blue-600 text-white px-3 py-1 rounded"
        >
          Apply
        </button>
      </div>

      <!-- Absolute selector -->
      <div v-else class="space-y-2 mt-2">
        <div class="flex gap-2">
          <input
            type="date"
            v-model="absolute.startDate"
            class="border p-1 rounded"
          />
          <input
            type="time"
            v-model="absolute.startTime"
            class="border p-1 rounded"
          />
        </div>
        <div class="flex gap-2">
          <input
            type="date"
            v-model="absolute.endDate"
            class="border p-1 rounded"
          />
          <input
            type="time"
            v-model="absolute.endTime"
            class="border p-1 rounded"
          />
        </div>
        <button
          @click="emitRange()"
          class="bg-blue-600 text-white px-3 py-1 rounded"
        >
          Apply
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
const emit = defineEmits(["update"]);

const isRelative = ref(true);
const autoRefresh = ref(false);
const showFullPicker = ref(false);

const relative = ref({
  amount: 3,
  unit: "hours",
});

const absolute = ref({
  startDate: "",
  startTime: "00:00:00",
  endDate: "",
  endTime: "23:59:59",
});

const quickPresets = [
  { label: "5m", amount: 5, unit: "minutes" },
  { label: "15m", amount: 15, unit: "minutes" },
  { label: "30m", amount: 30, unit: "minutes" },
  { label: "1h", amount: 1, unit: "hours" },
  { label: "3h", amount: 3, unit: "hours" },
  { label: "1d", amount: 1, unit: "days" },
  { label: "1w", amount: 7, unit: "days" },
];

function applyPreset(preset) {
  isRelative.value = true;
  relative.value = { amount: preset.amount, unit: preset.unit };
  emitRange();
}

function emitRange() {
  if (isRelative.value) {
    const now = new Date();
    const start = new Date(now);
    const { amount, unit } = relative.value;
    if (unit === "minutes") start.setMinutes(now.getMinutes() - amount);
    else if (unit === "hours") start.setHours(now.getHours() - amount);
    else if (unit === "days") start.setDate(now.getDate() - amount);
    emit("update", {
      start: start.toISOString(),
      end: now.toISOString(),
      autoRefresh: autoRefresh.value,
    });
  } else {
    const start = new Date(
      `${absolute.value.startDate}T${absolute.value.startTime}Z`,
    );
    const end = new Date(
      `${absolute.value.endDate}T${absolute.value.endTime}Z`,
    );
    emit("update", {
      start: start.toISOString(),
      end: end.toISOString(),
      autoRefresh: autoRefresh.value,
    });
  }
}
</script>

<style scoped>
input[type="date"],
input[type="time"],
select {
  min-width: 140px;
}
</style>
