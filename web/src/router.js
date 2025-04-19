import { createRouter, createWebHistory } from "vue-router";
import Home from "./views/Home.vue";
import DeployBackend from "./views/DeployBackend.vue";
import BackendDetail from "./views/BackendDetail.vue";
import Deployments from "./views/Deployments.vue";
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
    path: "/backend/:stackName",
    name: "BackendDetail",
    component: BackendDetail,
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

export default router;
