import * as tus from "tus-js-client";
import { baseURL, tusEndpoint, tusSettings, origin } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
import { removePrefix } from "@/api/utils";

const RETRY_BASE_DELAY = 1000;
const RETRY_MAX_DELAY = 20000;
const CURRENT_UPLOAD_LIST: { [key: string]: tus.Upload } = {};

export async function upload(
  filePath: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any
) {
  if (!tusSettings) {
    // Shouldn't happen as we check for tus support before calling this function
    throw new Error("Tus.io settings are not defined");
  }

  filePath = removePrefix(filePath);
  
  // For S3 storage, strip the bucket name from the URL
  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3") {
    // Handle both formats: /bucket/path and /buckets/bucket/path
    const bucketMatch = filePath.match(/^\/buckets\/([^/]+)/);
    if (bucketMatch) {
      // Remove /buckets/bucket prefix
      filePath = filePath.slice(bucketMatch[0].length) || '/';
    } else {
      // Also check for /bucket format (without /buckets prefix)
      const simpleBucketMatch = filePath.match(/^\/([^/]+)/);
      if (simpleBucketMatch) {
        filePath = filePath.slice(simpleBucketMatch[0].length) || '/';
      }
    }
  }

  const authStore = useAuthStore();
  
  // Get current scope/bucket for S3 storage
  let scopeParam = "";
  if (appConfig.StorageType === "s3" && authStore.user?.currentScope?.name) {
    scopeParam = `&scope=${encodeURIComponent(authStore.user.currentScope.name)}`;
  }

  const resourcePath = `${tusEndpoint}${filePath}?override=${overwrite}${scopeParam}`;

  // Exit early because of typescript, tus content can't be a string
  if (content === "") {
    return false;
  }
  return new Promise<void | string>((resolve, reject) => {
    const upload = new tus.Upload(content, {
      endpoint: `${origin}${baseURL}${resourcePath}`,
      chunkSize: tusSettings.chunkSize,
      retryDelays: computeRetryDelays(tusSettings),
      parallelUploads: 1,
      storeFingerprintForResuming: false,
      headers: {
        "X-Auth": authStore.jwt,
      },
      onShouldRetry: function (err) {
        const status = err.originalResponse
          ? err.originalResponse.getStatus()
          : 0;

        // Do not retry for file conflict.
        if (status === 409) {
          return false;
        }

        return true;
      },
      onError: function (error: Error | tus.DetailedError) {
        delete CURRENT_UPLOAD_LIST[filePath];

        if (error.message === "Upload aborted") {
          return reject(error);
        }

        const message =
          error instanceof tus.DetailedError
            ? error.originalResponse === null
              ? "000 No connection"
              : error.originalResponse.getBody()
            : "Upload failed";

        console.error(error);

        reject(new Error(message));
      },
      onProgress: function (bytesUploaded) {
        if (typeof onupload === "function") {
          onupload({ loaded: bytesUploaded });
        }
      },
      onSuccess: function () {
        delete CURRENT_UPLOAD_LIST[filePath];
        resolve();
      },
    });
    CURRENT_UPLOAD_LIST[filePath] = upload;
    upload.start();
  });
}

function computeRetryDelays(tusSettings: TusSettings): number[] | undefined {
  if (!tusSettings.retryCount || tusSettings.retryCount < 1) {
    // Disable retries altogether
    return undefined;
  }
  // The tus client expects our retries as an array with computed backoffs
  // E.g.: [0, 3000, 5000, 10000, 20000]
  const retryDelays = [];
  let delay = 0;

  for (let i = 0; i < tusSettings.retryCount; i++) {
    retryDelays.push(Math.min(delay, RETRY_MAX_DELAY));
    delay =
      delay === 0 ? RETRY_BASE_DELAY : Math.min(delay * 2, RETRY_MAX_DELAY);
  }

  return retryDelays;
}

export async function useTus(content: ApiContent) {
  return isTusSupported() && content instanceof Blob;
}

function isTusSupported() {
  return tus.isSupported === true;
}

export function abortAllUploads() {
  for (const filePath in CURRENT_UPLOAD_LIST) {
    if (CURRENT_UPLOAD_LIST[filePath]) {
      CURRENT_UPLOAD_LIST[filePath].abort(true);
      CURRENT_UPLOAD_LIST[filePath].options!.onError!(
        new Error("Upload aborted")
      );
    }
    delete CURRENT_UPLOAD_LIST[filePath];
  }
}
