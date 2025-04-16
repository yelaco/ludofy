<template>
  <div class="mb-6">
    <label class="block text-sm font-semibold mb-1">Platform Name</label>
    <input
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      placeholder="e.g. ChessWorld"
      class="w-full px-4 py-2 border rounded-lg shadow-sm focus:outline-none focus:ring focus:border-blue-300"
    />
  </div>
</template>

<script setup>
defineProps({
  modelValue: String,
});
defineEmits(["update:modelValue"]);
</script>
