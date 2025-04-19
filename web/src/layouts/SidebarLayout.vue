<template>
  <div
    class="flex min-h-screen text-gray-800 bg-gray-50 dark:text-gray-200 dark:bg-gray-900 transition-colors duration-300"
  >
    <!-- Sidebar -->
    <aside
      :class="[
        'bg-white dark:bg-gray-800 dark:border-gray-700 h-screen border-r shadow-sm overflow-hidden transition-all duration-300 ease-in-out',
        collapsed ? 'w-16' : 'w-64',
      ]"
    >
      <!-- Sidebar Header -->
      <div class="h-16 flex items-center px-4">
        <!-- Expanded state -->
        <template v-if="!collapsed">
          <!-- Centered logo text in the available space -->
          <div class="flex-1 flex justify-center">
            <RouterLink
              to="/"
              class="text-xl font-bold flex items-center gap-2 hover:text-blue-600 dark:text-white dark:hover:text-blue-300 transition"
            >
              🎮 Ludofy
            </RouterLink>
          </div>

          <!-- Collapse button on the right -->
          <button
            @click="toggleSidebar"
            class="text-gray-600 hover:text-black dark:text-gray-300 dark:hover:text-white"
          >
            <component :is="icons.ChevronLeft" class="w-5 h-5" />
          </button>
        </template>

        <!-- Collapsed state -->
        <template v-else>
          <div class="w-full flex justify-center">
            <button
              @click="toggleSidebar"
              class="text-gray-600 hover:text-black dark:text-gray-300 dark:hover:text-white"
            >
              <component :is="icons.Menu" class="w-5 h-5" />
            </button>
          </div>
        </template>
      </div>

      <nav class="flex flex-col gap-1 px-2 py-4">
        <SidebarLink to="/" icon="Home" label="Home" :collapsed="collapsed" />
        <SidebarLink
          to="/deploy"
          icon="PlusSquare"
          label="Deploy Backend"
          :collapsed="collapsed"
        />
        <SidebarLink
          to="/deployments"
          icon="Server"
          label="Deployments"
          :collapsed="collapsed"
        />
        <SidebarLink
          to="/settings"
          icon="Settings"
          label="Settings"
          :collapsed="collapsed"
        />
      </nav>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col">
      <header
        class="bg-white dark:bg-gray-800 px-6 py-3 border-b shadow flex justify-between items-center dark:border-gray-700"
      >
        <RouterLink
          to="/"
          class="text-lg font-semibold hover:text-blue-600 dark:text-white dark:hover:text-blue-300 transition"
        >
          Ludofy Platform
        </RouterLink>

        <div class="flex items-center gap-4">
          <!-- Notifications Icon -->
          <button
            class="relative text-gray-500 hover:text-blue-600 dark:text-gray-300"
          >
            <component :is="icons.Bell" class="w-5 h-5" />
            <span
              class="absolute top-0 right-0 inline-block w-2 h-2 bg-red-500 rounded-full animate-ping"
            ></span>
          </button>

          <!-- User Profile -->
          <div
            class="relative"
            @mouseenter="showDropdown = true"
            @mouseleave="showDropdown = false"
          >
            <button
              class="flex items-center gap-2 text-gray-700 dark:text-white"
            >
              <component :is="icons.User" class="w-5 h-5" />
              <span v-if="!collapsed" class="text-sm font-medium">me</span>
            </button>
            <div
              v-if="showDropdown"
              class="absolute right-0 mt-2 w-40 bg-white dark:bg-gray-700 text-sm shadow rounded py-2 z-50"
            >
              <a
                href="#"
                class="block px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600"
                >Profile</a
              >
              <a
                href="#"
                class="block px-4 py-2 text-red-500 hover:bg-red-50 dark:hover:bg-gray-600"
                >Logout</a
              >
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 p-6 overflow-y-auto bg-gray-50 dark:bg-gray-900">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";
import SidebarLink from "./SidebarLink.vue";
import { RouterView } from "vue-router";
import * as icons from "lucide-vue-next";

// --- Persistent collapsed state ---
const COLLAPSE_KEY = "ludofy-sidebar-collapsed";
const collapsed = ref(localStorage.getItem(COLLAPSE_KEY) === "true");

watch(collapsed, (val) => {
  localStorage.setItem(COLLAPSE_KEY, String(val));
});

function toggleSidebar() {
  collapsed.value = !collapsed.value;
}

// Profile dropdown
const showDropdown = ref(false);
</script>
