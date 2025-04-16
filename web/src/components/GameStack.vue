<template>
  <div class="border rounded-lg p-4 shadow-sm bg-white">
    <div class="flex justify-between items-center mb-4">
      <p class="font-semibold text-lg">🎮 Game #{{ index + 1 }}</p>
      <button
        v-if="showRemove"
        @click="$emit('remove')"
        class="text-red-500 hover:underline"
      >
        Remove
      </button>
    </div>

    <div class="mb-3">
      <label class="block text-sm font-medium mb-1">Game Name</label>
      <input
        v-model="localGame.name"
        @input="emitUpdate"
        class="w-full px-3 py-2 border rounded-md"
        placeholder="Game Name"
      />
    </div>

    <div class="mb-3">
      <label class="block text-sm font-medium mb-1">Docker Image URL</label>
      <div class="flex gap-2">
        <input
          v-model="localGame.dockerImage"
          @input="emitUpdate"
          class="flex-1 px-3 py-2 border rounded-md"
          placeholder="e.g. 123456.dkr.ecr.aws/...:latest"
        />
        <button
          @click="$emit('upload')"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Upload
        </button>
      </div>
    </div>

    <div class="mt-4 space-y-2">
      <h4 class="font-medium">Internal Services</h4>
      <label class="flex items-center gap-2">
        <input type="checkbox" disabled checked class="accent-blue-600" />
        Matchmaking <span class="text-sm text-gray-500">(Required)</span>
      </label>
      <label class="flex items-center gap-2">
        <input
          type="checkbox"
          v-model="localGame.ranking"
          @change="emitUpdate"
        />
        Ranking
      </label>
      <label class="flex items-center gap-2">
        <input
          type="checkbox"
          v-model="localGame.spectating"
          @change="emitUpdate"
        />
        Match Spectating
      </label>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch } from "vue";

const props = defineProps({
  game: Object,
  index: Number,
  showRemove: Boolean,
});
const emit = defineEmits(["update:game", "remove", "upload"]);

const localGame = reactive({ ...props.game });

function emitUpdate() {
  emit("update:game", localGame);
}
</script>
