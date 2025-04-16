import { createRouter, createWebHistory } from "vue-router";
import Home from "./views/Home.vue";
import CreatePlatform from "./views/CreatePlatform.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/create",
    name: "CreatePlatform",
    component: CreatePlatform,
  },
  {
    path: "/platform/:id",
    name: "PlatformDetail",
    component: () => import("./views/PlatformDetail.vue"),
  },
  {
    path: "/deployments",
    name: "Deployments",
    component: () => import("./views/Deployments.vue"),
  },
  {
    path: "/settings",
    name: "Settings",
    component: () => import("./views/Settings.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
