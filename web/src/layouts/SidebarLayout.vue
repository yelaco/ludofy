<template>
  <div
    class="flex min-h-screen text-gray-800 bg-gray-50 dark:text-gray-200 dark:bg-gray-900 transition-colors duration-300"
  >
    <!-- Sidebar -->
    <aside
      :class="[
        'bg-white dark:bg-gray-800 dark:border-gray-700 min-h-full border-r shadow-sm overflow-hidden transition-all duration-300 ease-in-out',
        collapsed ? 'w-16' : 'w-64',
      ]"
    >
      <div class="h-16 flex items-center px-4">
        <template v-if="!collapsed">
          <div class="flex-1 flex justify-center">
            <RouterLink
              to="/"
              class="text-xl font-bold flex items-center gap-2 hover:text-blue-600 dark:text-white dark:hover:text-blue-300 transition"
            >
              🎮 Ludofy
            </RouterLink>
          </div>
          <button
            @click="toggleSidebar"
            class="text-gray-600 hover:text-black dark:text-gray-300 dark:hover:text-white"
          >
            <component :is="icons.ChevronLeft" class="w-5 h-5" />
          </button>
        </template>
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
          icon="Upload"
          label="Deploy Backend"
          :collapsed="collapsed"
        />
        <SidebarLink
          to="/backends"
          icon="Layers"
          label="Backends"
          :collapsed="collapsed"
        />
        <SidebarLink
          to="/deployments"
          icon="Box"
          label="Deployments"
          :collapsed="collapsed"
        />

        <SidebarLink
          to="/help"
          icon="BookOpen"
          label="Help"
          :collapsed="collapsed"
        />
      </nav>
      <SidebarLink
        to="/settings"
        icon="Settings"
        label="Settings"
        :collapsed="collapsed"
      />
    </aside>

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
          <button
            class="relative text-gray-500 hover:text-blue-600 dark:text-gray-300"
          >
            <component :is="icons.Bell" class="w-5 h-5" />
            <span
              class="absolute top-0 right-0 inline-block w-2 h-2 bg-red-500 rounded-full animate-ping"
            ></span>
          </button>

          <div class="relative">
            <button
              @click="toggleDropdown"
              class="flex items-center gap-2 text-gray-700 dark:text-white"
            >
              <component :is="icons.User" class="w-5 h-5" />
              <span class="text-sm font-medium">{{ userEmail }}</span>
            </button>
            <div
              v-if="showDropdown"
              class="absolute right-0 mt-2 w-40 bg-white dark:bg-gray-700 text-sm shadow rounded py-2 z-50"
            >
              <a
                href="#"
                class="block px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600"
              >
                Profile
              </a>
              <button
                @click="logout"
                class="w-full text-left block px-4 py-2 text-red-500 hover:bg-red-50 dark:hover:bg-gray-600"
              >
                Logout
              </button>
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
import { ref, watch, onMounted } from "vue";
import { signOutRedirect } from "@/auth";
import { userManager } from "@/auth";
import SidebarLink from "./SidebarLink.vue";
import { RouterView } from "vue-router";
import * as icons from "lucide-vue-next";

const COLLAPSE_KEY = "ludofy-sidebar-collapsed";
const collapsed = ref(localStorage.getItem(COLLAPSE_KEY) === "true");

watch(collapsed, (val) => {
  localStorage.setItem(COLLAPSE_KEY, String(val));
});

function toggleSidebar() {
  collapsed.value = !collapsed.value;
}

const showDropdown = ref(false);
const userEmail = ref("me");

function toggleDropdown() {
  showDropdown.value = !showDropdown.value;
}

async function logout() {
  try {
    await userManager.removeUser();
    await signOutRedirect();
    console.log("Signed out successfully");
  } catch (error) {
    console.error("Error during logout:", error);
  }
}

onMounted(async () => {
  try {
    const user = await userManager.getUser();
    if (user && user.profile && user.profile.email) {
      userEmail.value = user.profile.email;
    }
  } catch (error) {
    console.error("Failed to fetch user profile:", error);
  }
});
</script>
