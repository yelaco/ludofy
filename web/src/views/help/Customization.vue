<template>
  <div class="p-6 max-w-3xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Backend Customization</h1>

    <p class="mb-4 text-gray-600">
      Here's an example of a valid <strong>customization.yaml</strong> you can
      upload:
    </p>

    <div
      class="relative w-full max-w-2xl mx-auto rounded-lg overflow-hidden shadow-sm border border-gray-700 bg-white-900"
    >
      <div
        class="flex items-center justify-between bg-gray-200 px-4 py-2 text-xs text-gray-600"
      >
        <span class="bg-white-700 text-gray-600 px-2 py-0 rounded text-xs"
          >yaml</span
        >
        <button
          @click="copyCode"
          class="hover:text-black flex items-center space-x-1"
        >
          <span v-if="!copied">📋 Copy</span>
          <span v-else>✅ Copied!</span>
        </button>
      </div>

      <pre class="p-0 m-0 overflow-x-auto bg-transparent"><code
          ref="codeBlock"
          class="yaml text-sm leading-relaxed p-5 bg-transparent"
        ></code></pre>
    </div>

    <p class="mt-6 text-gray-500">
      Customize the fields according to your game requirements. Make sure the
      file extension is <strong>.yaml</strong> or <strong>.yml</strong>.
    </p>

    <RouterLink
      to="/backends"
      class="inline-block mt-8 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
    >
      ← Back to backends
    </RouterLink>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import hljs from "highlight.js/lib/core";
import yaml from "highlight.js/lib/languages/yaml";
import "highlight.js/styles/github.css"; // Catppuccin Mocha theme

const codeBlock = ref(null);
const copied = ref(false);

hljs.registerLanguage("yaml", yaml);

async function loadExampleFile() {
  try {
    const response = await fetch("/help/customization-example.yaml");
    const text = await response.text();
    if (codeBlock.value) {
      codeBlock.value.textContent = text;
      hljs.highlightElement(codeBlock.value);
    }
  } catch (error) {
    console.error("Error loading example file:", error);
  }
}

async function copyCode() {
  if (codeBlock.value) {
    try {
      await navigator.clipboard.writeText(codeBlock.value.textContent);
      copied.value = true;
      setTimeout(() => (copied.value = false), 1500);
    } catch (error) {
      console.error("Failed to copy text:", error);
    }
  }
}

onMounted(loadExampleFile);
</script>

<style scoped>
pre,
code {
  background: transparent !important;
}
</style>
