import { WebStorageStateStore, UserManager } from "oidc-client-ts";

const cognitoAuthConfig = {
  authority: `https://cognito-idp.ap-southeast-2.amazonaws.com/${import.meta.env.VITE_COGNITO_USER_POOL_ID}`,
  client_id: import.meta.env.VITE_COGNITO_APP_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_REDIRECT_SIGN_IN,
  response_type: "code",
  scope: "openid email profile",
  post_logout_redirect_uri: "http://localhost:5173/",
  userStore: new WebStorageStateStore({ store: window.localStorage }),
};

export const userManager = new UserManager({
  ...cognitoAuthConfig,
});

export async function signOutRedirect() {
  const clientId = import.meta.env.VITE_COGNITO_APP_CLIENT_ID;
  const logoutUri = import.meta.env.VITE_REDIRECT_SIGN_OUT;
  const cognitoDomain = import.meta.env.VITE_COGNITO_DOMAIN;
  window.location.href = `${cognitoDomain}/logout?client_id=${clientId}&logout_uri=${encodeURIComponent(logoutUri)}`;
}

export async function handleAuthRedirect() {
  const url = window.location.href;
  if (url.includes("?code=") && url.includes("state=")) {
    console.log("Detected Cognito callback, processing signin...");
    try {
      await userManager.signinCallback();
      console.log("Signin callback success!");
      window.history.replaceState({}, document.title, "/");
    } catch (error) {
      console.error("Error during signin callback:", error);
    }
  }
}
