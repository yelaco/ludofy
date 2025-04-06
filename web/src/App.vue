<template>
  <div class="container">
    <h1>Platform setup</h1>

    <label>Platform Name</label>
    <input v-model="platformName" placeholder="e.g. ChessWorld" />

    <h2>Shared services</h2>
    <div>
      <label
        ><input type="checkbox" disabled checked /> Authentication
        (Required)</label
      >
      <label><input type="checkbox" v-model="services.chat" /> Chat</label>
      <label><input type="checkbox" v-model="services.friend" /> Friend</label>
    </div>

    <h2>Games</h2>
    <div v-for="(game, index) in gameStacks" :key="index" class="game-stack">
      <p>Game #{{ index + 1 }}</p>

      <label>Game Name</label>
      <input v-model="game.name" placeholder="Game Name" />

      <label>Docker Image URL</label>
      <div class="docker-input-row">
        <input
          v-model="game.dockerImage"
          placeholder="e.g. 123456.dkr.ecr.aws/...:latest"
        />
        <button @click="uploadDockerImage(index)">Upload</button>
      </div>

      <h4>Internal services</h4>
      <div>
        <label>
          <input type="checkbox" disabled checked />
          Matchmaking (Required)
        </label>
        <label>
          <input type="checkbox" v-model="game.ranking" />
          Ranking
        </label>
        <label>
          <input type="checkbox" v-model="game.spectating" />
          Match Spectating
        </label>
      </div>

      <button v-if="gameStacks.length >= 2" @click="removeGameStack(index)">
        Remove
      </button>
    </div>

    <button @click="addGameStack">Add Game</button>

    <hr />

    <button @click="deployPlatform">🚀 Deploy Platform</button>
  </div>
</template>

<script setup>
import { ref } from "vue";

const platformName = ref("");
const services = ref({
  authentication: true,
  chat: false,
  friend: false,
});

const gameStacks = ref([
  { name: "", dockerImage: "", ranking: false, spectating: false },
]);

function addGameStack() {
  gameStacks.value.push({
    name: "",
    dockerImage: "",
    ranking: false,
    spectating: false,
  });
}

function removeGameStack(index) {
  gameStacks.value.splice(index, 1);
}

function uploadDockerImage(index) {
  alert(`Upload Docker image for game #${index + 1}`);
  // In reality, open a file picker or Docker push UI
}

function deployPlatform() {
  const data = {
    platformName: platformName.value,
    services: services.value,
    gameStacks: gameStacks.value,
  };
  console.log("Deploying...", data);

  // You can send this to your backend
  // fetch('/api/deploy', { method: 'POST', body: JSON.stringify(data) })
}
</script>

<style>
.container {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  font-family: sans-serif;
}

input {
  display: block;
  margin: 0.5rem 0;
  width: 100%;
  padding: 0.5rem;
  box-sizing: border-box;
}

.docker-input-row {
  display: flex;
  gap: 0.5rem;
}

.docker-input-row input {
  flex: 1;
}

.docker-input-row button {
  white-space: nowrap;
}

button {
  padding: 0.4rem 1rem;
  margin-top: 0.5rem;
}

.game-stack {
  margin-bottom: 1rem;
  border: 1px solid #ccc;
  padding: 1rem;
}
</style>
