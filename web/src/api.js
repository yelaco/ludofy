import axios from "axios";
import { userManager } from "@/auth"; // Assuming you already have this

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Interceptor to add Authorization header dynamically
api.interceptors.request.use(
  async (config) => {
    const user = await userManager.getUser();
    if (user && user.id_token) {
      config.headers["Authorization"] = `${user.id_token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

export default {
  getDeployments(query) {
    return api.get(`/deployments${query}`);
  },
  deployBackend(data) {
    return api.post("/deploy", data);
  },
  updateBackend(data) {
    return api.post("/update", data);
  },
  getPresignedCustomizationUrl(id) {
    return api.post(`/backend/${id}/customize`);
  },
  getBackendDeployment(id) {
    return api.get(`/backend/${id}/deployment`);
  },
  removeBackend(id) {
    return api.delete(`/backend/${id}`);
  },
  getBackend(id) {
    return api.get(`/backend/${id}`);
  },
  getBackends(query) {
    return api.get(`/backends${query}`);
  },
  getDeployment(id) {
    return api.get(`/deployment/${id}`);
  },
};
