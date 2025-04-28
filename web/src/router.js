import { createRouter, createWebHistory } from "vue-router";
import { userManager } from "@/auth";

import Home from "./views/Home.vue";
import DeployBackend from "./views/DeployBackend.vue";
import UpdateBackend from "./views/UpdateBackend.vue";
import BackendDetails from "./views/BackendDetail.vue";
import Deployments from "./views/Deployments.vue";
import Backends from "./views/Backends.vue";
import Settings from "./views/Settings.vue";
import Help from "./views/Help.vue";
import HelpCustomization from "./views/help/Customization.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/deploy",
    name: "DeployBackend",
    component: DeployBackend,
    meta: { title: "Deploy backend - Ludofy" },
  },
  {
    path: "/backend/:id/update",
    name: "UpdateBackend",
    component: UpdateBackend,
    meta: { title: "Update backend - Ludofy" },
  },
  {
    path: "/backend/:id",
    name: "BackendDetails",
    component: BackendDetails,
    meta: { title: "Backend details - Ludofy" },
  },
  {
    path: "/backends",
    name: "Backends",
    component: Backends,
    meta: { title: "Backends - Ludofy" },
  },
  {
    path: "/deployments",
    name: "Deployments",
    component: Deployments,
    meta: { title: "Deployments - Ludofy" },
  },
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
    meta: { title: "Settings - Ludofy" },
  },
  {
    path: "/help/customization",
    name: "HelpCustomization",
    component: HelpCustomization,
    meta: { title: "Customization Help - Ludofy" },
  },
  {
    path: "/help",
    name: "Help",
    component: Help,
    meta: { title: "Help center - Ludofy" },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to, from, next) => {
  const user = await userManager.getUser();
  if (!user || user.expired) {
    console.warn("No user session, redirecting to Cognito login...");
    return userManager.signinRedirect();
  }
  next();
});

router.afterEach((to) => {
  if (to.meta && to.meta.title) {
    document.title = to.meta.title;
  } else {
    document.title = "Ludofy"; // fallback
  }
});

export default router;
