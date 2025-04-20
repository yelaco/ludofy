import { createRouter, createWebHistory } from "vue-router";
import { userManager } from "@/auth";

import Home from "./views/Home.vue";
import DeployBackend from "./views/DeployBackend.vue";
import BackendDetail from "./views/BackendDetail.vue";
import Deployments from "./views/Deployments.vue";
import Backends from "./views/Backends.vue";
import Settings from "./views/Settings.vue";

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
  },
  {
    path: "/backend/:id",
    name: "BackendDetail",
    component: BackendDetail,
  },
  {
    path: "/backends",
    name: "Backends",
    component: Backends,
  },
  {
    path: "/deployments",
    name: "Deployments",
    component: Deployments,
  },
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
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
  console.log(user.id_token);
  next();
});

export default router;
