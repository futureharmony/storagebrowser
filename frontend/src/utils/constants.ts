const getFileBrowser = (): any => {
  return (window as any).FileBrowser || {};
};

const name: string = getFileBrowser().Name || "File Browser";
const disableExternal: boolean = getFileBrowser().DisableExternal;
const disableUsedPercentage: boolean = getFileBrowser().DisableUsedPercentage;
const baseURL: string = getFileBrowser().BaseURL || "";
const staticURL: string = getFileBrowser().StaticURL || "/static";
const recaptcha: string = getFileBrowser().ReCaptcha;
const recaptchaKey: string = getFileBrowser().ReCaptchaKey;
const signup: boolean = getFileBrowser().Signup;
const logoURL = `${staticURL}/img/logo.svg`;
const noAuth: boolean = getFileBrowser().NoAuth;
const authMethod = getFileBrowser().AuthMethod || "";
const loginPage: boolean = getFileBrowser().LoginPage;
const theme: UserTheme = getFileBrowser().Theme || "";
const enableThumbs: boolean = getFileBrowser().EnableThumbs;
const resizePreview: boolean = getFileBrowser().ResizePreview;
const enableExec: boolean = getFileBrowser().EnableExec;
const tusSettings = getFileBrowser().TusSettings || {
  chunkSize: 10485760,
  retryCount: 5,
};
const origin = typeof window !== "undefined" ? window.location.origin : "";
const tusEndpoint = `/api/tus`;

// Debug version loading
console.log("Constants version loaded:", getFileBrowser().Version || "");
console.log("FileBrowser object:", getFileBrowser());

// Function to check version status
export const checkVersionStatus = () => {
  const currentVersion = getFileBrowser().Version || "";
  console.log("Version check - Current:", currentVersion, "Loaded:", getFileBrowser().Version || "");
  return currentVersion;
};

// Export version as a function to get current value
export const version = () => getFileBrowser().Version || "";

export {
  name,
  disableExternal,
  disableUsedPercentage,
  baseURL,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
  noAuth,
  authMethod,
  loginPage,
  theme,
  enableThumbs,
  resizePreview,
  enableExec,
  tusSettings,
  origin,
  tusEndpoint,
};
