<template>
  <div class="p-6 max-w-4xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Backend Customization</h1>

    <!-- Section: Main Customization Example -->
    <p class="mb-4 text-gray-600">
      Here's an example of a valid <strong>customization.yaml</strong> you can
      upload:
    </p>

    <YamlBlock
      label="Customization Example"
      :content="blocks.customization"
      @copy="copyText(blocks.customization)"
    />

    <p class="mt-6 text-gray-500">
      Customize the fields according to your game requirements. Make sure the
      file extension is
      <strong>.yaml</strong> or <strong>.yml</strong>.
    </p>

    <p class="mt-4 text-gray-500">
      For more information about customizing backend templates, see
      <a
        href="https://docs.aws.amazon.com/serverless-application-model/latest/developerguide"
        target="_blank"
        rel="noopener noreferrer"
        class="text-blue-600 hover:underline"
      >
        AWS SAM Documentation </a
      >.
    </p>

    <!-- Section: Exported Values -->
    <h2 class="text-xl font-bold mt-10 mb-6">Available Exported Values</h2>

    <p class="text-gray-500 mb-8">
      When defining your <strong>customization.yaml</strong> file, you may need
      to reference resources that were deployed by the platform for your backend
      — such as AppSync API, Cognito User Pool, game server cluster, storage
      buckets, and more. <br /><br />
      These exported values are made available through AWS CloudFormation
      <strong>Exports</strong>. You can reference them in your YAML by using the
      <strong>Fn::ImportValue</strong> intrinsic function. <br /><br />
      Below are the exported outputs organized by category. You can copy and
      paste them into your customization files as needed.
    </p>

    <div class="space-y-8">
      <div v-for="block in outputBlocks" :key="block.label">
        <h3 class="text-lg font-semibold text-gray-700 mb-2">
          {{ block.title }}
        </h3>
        <YamlBlock
          :label="block.label"
          :content="block.content"
          @copy="copyText(block.content)"
        />
      </div>
    </div>

    <!-- Back Button -->
    <RouterLink
      to="/backends"
      class="inline-block mt-10 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
    >
      ← Back to backends
    </RouterLink>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import YamlBlock from "@/components/YamlBlock.vue";

const blocks = ref({
  customization: "",
});

const outputBlocks = ref([
  { title: "🔐 Authentication", label: "Authentication", content: "" },
  { title: "🗂️ Storage", label: "Storage", content: "" },
  { title: "🧩 AppSync API", label: "AppSync", content: "" },
  { title: "⚙️ Compute (Game Server)", label: "Compute", content: "" },
  { title: "🌐 HTTP API", label: "HttpApi", content: "" },
  { title: "🔌 WebSocket API", label: "WebsocketApi", content: "" },
]);

// Fetch YAML content from file
async function fetchYaml(filePath) {
  try {
    const response = await fetch(filePath);
    return await response.text();
  } catch (error) {
    console.error("Error loading file:", filePath, error);
    return "";
  }
}

// Load all YAML files
async function loadAllYaml() {
  blocks.value.customization = await fetchYaml(
    "/help/customization/customization-example.yaml",
  );

  for (const block of outputBlocks.value) {
    block.content = await fetchYaml(
      `/help/customization/outputs/${block.label.toLowerCase()}.yaml`,
    );
  }
}

// Copy text to clipboard
async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
  } catch (error) {
    console.error("Failed to copy text:", error);
  }
}

onMounted(loadAllYaml);
</script>
