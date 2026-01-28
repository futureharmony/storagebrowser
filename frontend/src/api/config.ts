import { fetchJSON } from "./utils";

export interface Config {
  Name: string;
  DisableExternal: boolean;
  DisableUsedPercentage: boolean;
  Color: string;
  BaseURL: string;
  Version: string;
  StaticURL: string;
  Signup: boolean;
  NoAuth: boolean;
  AuthMethod: string;
  LoginPage: boolean;
  CSS: boolean;
  ReCaptcha: boolean;
  Theme: string;
  EnableThumbs: boolean;
  ResizePreview: boolean;
  EnableExec: boolean;
  TusSettings: any;
  StorageType: string;
}

export function getConfig() {
  return fetchJSON<Config>(`/api/config`, {});
}

// Load config function (will be called after login)
export const loadConfig = async () => {
  try {
    const appConfig = await getConfig();

    // Set config on window for backward compatibility
    (window as any).FileBrowser = appConfig;

    // Update static URL function
    (window as any).__prependStaticUrl = (url: string) => {
      return `${appConfig.StaticURL}/${url.replace(/^\/+/, "")}`;
    };

    // Update manifest with loaded config
    if ((window as any).generateManifest) {
      (window as any).generateManifest(appConfig);
    }

    console.log("App config loaded:", appConfig);
  } catch (error) {
    console.error("Failed to load app config:", error);
    // Fallback to default config
    (window as any).FileBrowser = {
      StaticURL: "/static",
      BaseURL: "",
      StorageType: "local",
    };
  }
};
