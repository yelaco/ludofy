import "@/style.css";

import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import { handleAuthRedirect } from "@/auth"; // ⬅️ import the handler

async function main() {
  await handleAuthRedirect(); // ⬅️ Important! complete login before mounting

  const app = createApp(App);
  app.use(router);
  app.mount("#app");
}

main();
