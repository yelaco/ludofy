<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <!-- Top Bar -->
    <nav class="bg-white shadow px-6 py-4 flex justify-between items-center">
      <h1 class="text-xl font-bold text-blue-700">Ludofy</h1>
      <div class="space-x-4 text-sm">
        <RouterLink to="/" class="text-blue-600 hover:underline"
          >Home</RouterLink
        >
        <RouterLink to="/create" class="text-blue-600 hover:underline"
          >Create Platform</RouterLink
        >
        <RouterLink to="/deployments" class="text-blue-600 hover:underline"
          >Deployments</RouterLink
        >
        <!-- Add more links as needed -->
      </div>
    </nav>

    <!-- Main Content -->
    <main class="p-4">
      <RouterView />
    </main>

    <!-- Footer -->
    <footer class="text-center text-gray-500 mt-12 text-xs">
      © 2025 Ludofy PaaS
    </footer>
  </div>
</template>

<script setup>
import { RouterView, RouterLink } from "vue-router";
</script>
