<template>
  <div class="space-y-20 max-w-7xl mx-auto px-6 py-12">
    <!-- Hero Section -->
    <section class="text-center space-y-6 animate-fade-in">
      <h1 class="text-5xl font-bold text-blue-800">🎮 Ludofy Platform</h1>
      <p class="text-gray-600 text-xl">
        Deploy, manage, and scale your game backends effortlessly in the cloud.
      </p>

      <div class="flex flex-wrap justify-center gap-4 mt-8">
        <RouterLink
          to="/backends"
          class="px-6 py-3 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          🚀 View Backends
        </RouterLink>
        <RouterLink
          to="/deployments"
          class="px-6 py-3 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 transition"
        >
          📦 View Deployments
        </RouterLink>
      </div>
    </section>

    <!-- About Section -->
    <section class="text-center space-y-4 max-w-3xl mx-auto animate-fade-in-up">
      <h2 class="text-3xl font-semibold text-blue-700">About Ludofy</h2>
      <p class="text-gray-600 text-md">
        Ludofy provides a complete backend solution for multiplayer game
        developers. We help you focus on building great games by handling
        deployment, scaling, and server management for you.
      </p>
    </section>

    <!-- Features Section -->
    <section class="grid grid-cols-1 md:grid-cols-3 gap-8 animate-fade-in-up">
      <div
        class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-center"
      >
        <div class="text-4xl mb-2">⚡</div>
        <h3 class="text-xl font-semibold mb-2">Fast Deployment</h3>
        <p class="text-gray-500 text-sm">
          Spin up production-ready backends within minutes with minimal setup.
        </p>
      </div>

      <div
        class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-center"
      >
        <div class="text-4xl mb-2">📈</div>
        <h3 class="text-xl font-semibold mb-2">Auto Scaling</h3>
        <p class="text-gray-500 text-sm">
          Handle player spikes effortlessly with automatic scaling based on
          demand.
        </p>
      </div>

      <div
        class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-center"
      >
        <div class="text-4xl mb-2">🔒</div>
        <h3 class="text-xl font-semibold mb-2">Secure & Reliable</h3>
        <p class="text-gray-500 text-sm">
          Built-in security best practices and 99.9% uptime guarantee.
        </p>
      </div>
    </section>

    <!-- Developer Benefits Section -->
    <section class="text-center space-y-6 animate-fade-in-up">
      <h2 class="text-3xl font-semibold text-blue-700">
        Why Developers Choose Ludofy
      </h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8 mt-6">
        <div
          class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-left space-y-2"
        >
          <h4 class="text-lg font-semibold">🛠️ Easy Management</h4>
          <p class="text-gray-500 text-sm">
            Manage deployments, monitor server performance, and scale services
            without complex setup.
          </p>
        </div>
        <div
          class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-left space-y-2"
        >
          <h4 class="text-lg font-semibold">🚀 Speed to Market</h4>
          <p class="text-gray-500 text-sm">
            Launch your games faster by offloading infrastructure concerns to
            Ludofy.
          </p>
        </div>
        <div
          class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-left space-y-2"
        >
          <h4 class="text-lg font-semibold">💸 Cost Effective</h4>
          <p class="text-gray-500 text-sm">
            Pay only for what you use. Focus your budget on game development,
            not infrastructure.
          </p>
        </div>
        <div
          class="p-6 bg-white rounded-lg shadow hover:shadow-md transition text-left space-y-2"
        >
          <h4 class="text-lg font-semibold">📊 Real-time Insights</h4>
          <p class="text-gray-500 text-sm">
            Get real-time metrics and logs to monitor match health, player load,
            and backend status.
          </p>
        </div>
      </div>
    </section>

    <!-- Stats Section -->
    <section
      class="grid grid-cols-1 md:grid-cols-3 gap-8 text-center animate-fade-in-up"
    >
      <div class="p-6 bg-gray-100 rounded-lg shadow">
        <h3 class="text-4xl font-bold text-blue-700 mb-2">10,000+</h3>
        <p class="text-gray-600 text-sm">Matches Hosted</p>
      </div>
      <div class="p-6 bg-gray-100 rounded-lg shadow">
        <h3 class="text-4xl font-bold text-blue-700 mb-2">99.9%</h3>
        <p class="text-gray-600 text-sm">Uptime Guarantee</p>
      </div>
      <div class="p-6 bg-gray-100 rounded-lg shadow">
        <h3 class="text-4xl font-bold text-blue-700 mb-2">5 mins</h3>
        <p class="text-gray-600 text-sm">Average Deploy Time</p>
      </div>
    </section>

    <!-- CTA Banner -->
    <section
      class="bg-blue-600 text-white rounded-lg p-8 text-center space-y-4 animate-fade-in"
    >
      <h2 class="text-3xl font-semibold">Ready to Deploy Your Game?</h2>
      <p class="text-white text-md">
        Get started in just a few clicks with Ludofy Platform.
      </p>
      <RouterLink
        to="/deploy"
        class="inline-block mt-4 px-6 py-3 bg-white text-blue-600 font-semibold rounded hover:bg-gray-100 transition"
      >
        🛠️ Deploy a Backend
      </RouterLink>
    </section>

    <!-- Footer -->
    <footer class="text-center text-gray-400 text-xs mt-20 animate-fade-in">
      © 2025 Ludofy. All rights reserved. · Terms · Privacy · Contact Us
    </footer>
  </div>
</template>

<script setup>
// No script needed for now
</script>

<style scoped>
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(40px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fadeIn 0.8s ease-out both;
}

.animate-fade-in-up {
  animation: fadeInUp 1s ease-out both;
}
</style>
