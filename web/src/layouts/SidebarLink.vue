<template>
  <RouterLink
    :to="to"
    :title="collapsed ? label : ''"
    class="flex items-center px-3 py-2 rounded hover:bg-blue-50 transition gap-3 relative"
    :class="{
      'bg-blue-100 text-blue-700 font-semibold': isActive,
      'justify-center': collapsed,
    }"
  >
    <component :is="iconComponent" class="w-5 h-5" />
    <span v-if="!collapsed" class="whitespace-nowrap">{{ label }}</span>
  </RouterLink>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";
import * as icons from "lucide-vue-next";

const props = defineProps({
  to: String,
  label: String,
  icon: String,
  collapsed: Boolean,
});

const iconComponent = computed(() => icons[props.icon] || icons.Circle);
const route = useRoute();
const isActive = computed(() => route.path === props.to);
</script>
