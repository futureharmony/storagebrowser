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
const version: string = getFileBrowser().Version || "";
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

export {
  name,
  disableExternal,
  disableUsedPercentage,
  baseURL,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
  version,
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
