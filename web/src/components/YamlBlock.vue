<template>
  <div
    :class="[
      'relative w-full max-w-2xl mx-auto rounded-lg overflow-hidden shadow-sm border border-gray-700 bg-white-900 transition-colors',
      copied ? 'bg-green-100' : 'bg-white-900',
    ]"
  >
    <div
      class="flex items-center justify-between bg-gray-200 px-4 py-2 text-xs text-gray-600"
    >
      <span class="bg-white-700 text-gray-600 px-2 py-0 rounded text-xs">
        {{ label }}
      </span>
      <button
        @click="copyContent"
        class="hover:text-black flex items-center space-x-1"
      >
        <span v-if="!copied">📋 Copy</span>
        <span v-else>✅ Copied!</span>
      </button>
    </div>

    <!-- Only horizontal scroll inside pre -->
    <pre
      class="p-0 m-0 overflow-x-auto bg-transparent"
    ><code ref="codeRef" class="yaml text-sm leading-relaxed p-5 bg-transparent block min-w-max"></code></pre>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";
import hljs from "highlight.js/lib/core";
import yaml from "highlight.js/lib/languages/yaml";
import "highlight.js/styles/github.css"; // Use GitHub style or any you like

const props = defineProps({
  label: String,
  content: String,
});

const emits = defineEmits(["copy"]);

const codeRef = ref(null);
const copied = ref(false);

hljs.registerLanguage("yaml", yaml);

watch(
  () => props.content,
  (newContent) => {
    if (codeRef.value) {
      codeRef.value.textContent = newContent || "";
      hljs.highlightElement(codeRef.value);
    }
  },
  { immediate: true },
);

async function copyContent() {
  try {
    await navigator.clipboard.writeText(props.content);
    copied.value = true;
    emits("copy");

    setTimeout(() => {
      copied.value = false;
    }, 1500);
  } catch (error) {
    console.error("Failed to copy text:", error);
  }
}
</script>

<style scoped>
pre,
code {
  background: transparent !important;
  white-space: pre; /* Keep formatting */
}
</style>
