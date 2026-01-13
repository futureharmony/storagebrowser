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
  S3Bucket: string;
  StorageType: string;
}

export function getConfig() {
  return fetchJSON<Config>(`/api/config`, {});
}
