<template>
  <div class="mb-8">
    <h2 class="text-xl font-semibold mb-2">Backend Services</h2>
    <div class="space-y-2">
      <label class="flex items-center gap-2">
        <input type="checkbox" disabled checked class="accent-blue-600" />
        Authentication <span class="text-sm text-gray-500">(Required)</span>
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" v-model="local.chat" @change="emitUpdate" />
        Chat
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" v-model="local.friend" @change="emitUpdate" />
        Friend
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" disabled checked class="accent-blue-600" />
        Matchmaking <span class="text-sm text-gray-500">(Required)</span>
      </label>
      <label class="flex items-center gap-2">
        <input type="checkbox" v-model="local.ranking" @change="emitUpdate" />
        Ranking
      </label>
      <label class="flex items-center gap-2">
        <input
          type="checkbox"
          v-model="local.matchSpectating"
          @change="emitUpdate"
        />
        Match Spectating
      </label>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch } from "vue";

const props = defineProps({ modelValue: Object });
const emit = defineEmits(["update:modelValue"]);

const local = reactive({ ...props.modelValue });

// Keep the local state in sync if parent updates modelValue
watch(
  () => props.modelValue,
  (newVal) => {
    Object.assign(local, newVal);
  },
);

function emitUpdate() {
  emit("update:modelValue", { ...local });
}
</script>
