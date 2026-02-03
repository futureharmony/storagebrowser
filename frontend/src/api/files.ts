import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { baseURL } from "@/utils/constants";
import { upload as postTus, useTus } from "./tus";
import { createURL, fetchURL, removePrefix, StatusError } from "./utils";

export async function fetch(url: string, signal?: AbortSignal, scope?: string) {
  console.log('[API FETCH] original url:', url);

  const appConfig = (window as any).FileBrowser || {};
  let path = url;

  if (appConfig.StorageType === "s3") {
    if (scope) {
      // For S3 storage with scope, strip the bucket prefix from the URL
      // The scope is already known, so we need the path within the bucket
      const bucketMatch = path.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        path = path.slice(bucketMatch[0].length) || '/';
      }
    } else {
      path = removePrefix(path);
      const bucketMatch = path.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        path = path.slice(bucketMatch[0].length) || '/';
      }
    }
  } else {
    path = removePrefix(path);
  }

  const urlParams = new URLSearchParams();
  // Decode first to avoid double encoding when path contains encoded characters
  const decodedPath = decodeURIComponent(path);
  urlParams.set('path', decodedPath);
  const res = await fetchURL(`/api/resources?${urlParams.toString()}`, { signal }, true, scope);

  let data: Resource;
  try {
    data = (await res.json()) as Resource;
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }

  // Determine the correct base URL for item links
  // If scope is provided, use S3 bucket path regardless of config (for compatibility)
  if (scope) {
    data.url = `/buckets/${scope}${url}`;
  } else {
    data.url = `/files${url}`;
  }

  if (data.isDir) {
    if (!data.url.endsWith("/")) data.url += "/";
    // Perhaps change the any
    data.items = data.items.map((item: any, index: any) => {
      item.index = index;
      item.url = `${data.url}${encodeURIComponent(item.name)}`;

      if (item.isDir) {
        item.url += "/";
      }

      return item;
    });
  }

  return data;
}

async function resourceAction(url: string, method: ApiMethod, content?: any, scope?: string) {
  const [pathWithoutQuery, queryPart] = url.split('?');
  let processedPath = pathWithoutQuery;

  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3" && scope) {
    // For S3 storage with scope, strip the bucket prefix from the URL
    const bucketMatch = processedPath.match(/^\/buckets\/([^/]+)/);
    if (bucketMatch) {
      processedPath = processedPath.slice(bucketMatch[0].length) || '/';
    }
  } else {
    processedPath = removePrefix(processedPath);
    
    if (appConfig.StorageType === "s3") {
      const bucketMatch = processedPath.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        processedPath = processedPath.slice(bucketMatch[0].length) || '/';
      }
    }
  }

  const opts: ApiOpts = {
    method,
  };

  if (content) {
    opts.body = content;
  }

  const baseUrl = `/api/resources`;
  const urlParams = new URLSearchParams();
  urlParams.set('path', processedPath);
  
  if (queryPart) {
    const originalParams = new URLSearchParams(queryPart);
    for (const [key, value] of originalParams) {
      urlParams.set(key, value);
    }
  }
  
  const res = await fetchURL(`${baseUrl}?${urlParams.toString()}`, opts, true, scope);

  return res;
}

export async function remove(url: string, scope?: string) {
  return resourceAction(url, "DELETE", undefined, scope);
}

export async function put(url: string, content = "", scope?: string) {
  return resourceAction(url, "PUT", content, scope);
}

export function download(format: any, ...files: string[]) {
  const appConfig = (window as any).FileBrowser || {};
  let url = `${baseURL}/api/raw`;
  const params = new URLSearchParams();

  if (files.length === 1) {
    let path = files[0];
    if (appConfig.StorageType === "s3") {
      const authStore = useAuthStore();
      const scope = authStore.user?.currentScope?.name;
      if (scope) {
        params.set('scope', scope);
        // Path should be relative to the bucket, strip bucket prefix if present
        if (path.startsWith('/buckets/' + scope)) {
          path = path.slice(('/buckets/' + scope).length) || '/';
        } else if (path.startsWith('/')) {
          path = path.slice(1);
        }
        params.set('path', path);
        if (format) {
          params.set('algo', format);
        }
        url += '?' + params.toString();
        window.open(url);
        return;
      }
    }
    params.set('path', removePrefix(path));
    url += '?' + params.toString();
  } else {
    let arg = "";

    for (const file of files) {
      arg += removePrefix(file) + ",";
    }

    arg = arg.substring(0, arg.length - 1);
    params.set('files', arg);
    url += '?' + params.toString();
  }

  if (format) {
    params.set('algo', format);
    // Reconstruct URL with updated params
    const newParams = new URLSearchParams(params.toString());
    url = `${baseURL}/api/raw?${newParams.toString()}`;
  }

  window.open(url);
}

export async function post(
  url: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any = () => {},
  scope?: string
) {
  // Use the pre-existing API if:
  const useResourcesApi =
    // a folder is being created
    url.endsWith("/") ||
    // We're not using http(s)
    (content instanceof Blob &&
      !["http:", "https:"].includes(window.location.protocol)) ||
    // Tus is disabled / not applicable
    !(await useTus(content));
  return useResourcesApi
    ? postResources(url, content, overwrite, onupload)
    : postTus(url, content, overwrite, onupload);
}

async function postResources(
  url: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any
) {
  url = removePrefix(url);

  // For S3 storage, also strip the bucket name from the URL
  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3") {
    // Handle both formats: /bucket/path and /buckets/bucket/path
    const bucketMatch = url.match(/^\/buckets\/([^/]+)/);
    if (bucketMatch) {
      // Remove /buckets/bucket prefix
      url = url.slice(bucketMatch[0].length) || '/';
    } else {
      // Also check for /bucket format (without /buckets prefix)
      const simpleBucketMatch = url.match(/^\/([^/]+)/);
      if (simpleBucketMatch) {
        url = url.slice(simpleBucketMatch[0].length) || '/';
      }
    }
  }

  let bufferContent: ArrayBuffer;
  if (
    content instanceof Blob &&
    !["http:", "https:"].includes(window.location.protocol)
  ) {
    bufferContent = await new Response(content).arrayBuffer();
  }

  const authStore = useAuthStore();
  
  // Get current scope/bucket for S3 storage
  const scope = appConfig.StorageType === "s3" ? authStore.user?.currentScope?.name : undefined;

  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    
    // Build URL with path and other parameters as query parameters
    const urlParams = new URLSearchParams();
    urlParams.set('path', url);
    urlParams.set('override', overwrite.toString());
    if (scope) {
      urlParams.set('scope', scope);
    }
    
    request.open(
      "POST",
      `${baseURL}/api/resources?${urlParams.toString()}`,
      true
    );
    request.setRequestHeader("X-Auth", authStore.jwt);

    if (typeof onupload === "function") {
      request.upload.onprogress = onupload;
    }

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText);
      } else if (request.status === 409) {
        reject(request.status);
      } else {
        reject(request.responseText);
      }
    };

    request.onerror = () => {
      reject(new Error("001 Connection aborted"));
    };

    request.send(bufferContent || content);
  });
}

function moveCopy(
  items: any[],
  copy = false,
  overwrite = false,
  rename = false
) {
  const layoutStore = useLayoutStore();
  const appConfig = (window as any).FileBrowser || {};
  const promises = [];

  for (const item of items) {
    const from = item.from;
    const to = item.to;
    
    let scope: string | undefined;
    let processedFrom = from;
    let processedTo = to ?? "";
    
    if (appConfig.StorageType === "s3") {
      const bucketMatch = from.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        scope = bucketMatch[1];
        processedFrom = from.slice(bucketMatch[0].length) || '/';
        
        const toBucketMatch = to?.match(/^\/buckets\/([^/]+)/);
        if (toBucketMatch) {
          processedTo = to.slice(toBucketMatch[0].length) || '/';
        }
      }
    } else {
      // For non-S3 storage, apply removePrefix
      processedFrom = removePrefix(from);
      processedTo = removePrefix(to ?? "");
    }
    
    // Build query parameters for PATCH request
    const urlParams = new URLSearchParams();
    urlParams.set('path', processedFrom);
    urlParams.set('action', copy ? "copy" : "rename");
    urlParams.set('destination', processedTo);
    urlParams.set('override', overwrite.toString());
    urlParams.set('rename', rename.toString());
    
    promises.push(resourceAction(`?${urlParams.toString()}`, "PATCH", undefined, scope));
  }
  layoutStore.closeHovers();
  return Promise.all(promises);
}

export function move(items: any[], overwrite = false, rename = false) {
  return moveCopy(items, false, overwrite, rename);
}

export function copy(items: any[], overwrite = false, rename = false) {
  return moveCopy(items, true, overwrite, rename);
}

export async function checksum(url: string, algo: ChecksumAlg, scope?: string) {
  // Extract path and query parameters
  const [path, query] = url.split('?');
  const urlParams = new URLSearchParams(query || '');
  urlParams.set('path', path);
  urlParams.set('checksum', algo);
  
  const data = await resourceAction(`?${urlParams.toString()}`, "GET", undefined, scope);
  return (await data.json()).checksums[algo];
}

export function getDownloadURL(file: ResourceItem, inline: any) {
  const params: Record<string, string> = {
    ...(inline && { inline: "true" }),
  };

  const appConfig = (window as any).FileBrowser || {};
  let path = file.path;

  // For S3 storage, use scope and path query parameters
  if (appConfig.StorageType === "s3") {
    const authStore = useAuthStore();
    const scope = authStore.user?.currentScope?.name;
    if (scope) {
      params.scope = scope;
      // Path should be relative to the bucket, strip bucket prefix if present
      if (path.startsWith('/buckets/' + scope)) {
        path = path.slice(('/buckets/' + scope).length) || '/';
      } else if (path.startsWith('/')) {
        path = path.slice(1); // Remove leading slash
      }
      params.path = path;
      return createURL("api/raw", params);
    }
  }

  return createURL("api/raw" + path, params);
}

export function getPreviewURL(file: ResourceItem, size: string) {
  const params = {
    inline: "true",
    key: Date.parse(file.modified),
  };

  return createURL("api/preview/" + size + file.path, params);
}

export function getSubtitlesURL(file: ResourceItem) {
  const params = {
    inline: "true",
  };

  return file.subtitles?.map((d) => createURL("api/subtitle" + d, params));
}

export async function usage(url: string, signal: AbortSignal, scope?: string) {
  const appConfig = (window as any).FileBrowser || {};

  // Process URL based on scope and storage type
  if (appConfig.StorageType === "s3") {
    if (scope) {
      // For S3 storage with scope, URL is already processed in Sidebar.vue
      // URL is already in the correct format (e.g., "/errors/")
      // No need to call removePrefix
    } else {
      url = removePrefix(url);
      // For S3 storage without scope, strip the bucket name from the URL
      // Handle both formats: /bucket/path and /buckets/bucket/path
      const bucketMatch = url.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        // Remove /buckets/bucket prefix
        url = url.slice(bucketMatch[0].length) || '/';
      } else {
        // Also check for /bucket format (without /buckets prefix)
        const simpleBucketMatch = url.match(/^\/([^/]+)/);
        if (simpleBucketMatch) {
          url = url.slice(simpleBucketMatch[0].length) || '/';
        }
      }
    }
  } else {
    // For non-S3 storage, always remove prefix
    url = removePrefix(url);
  }

  const urlParams = new URLSearchParams();
  // Decode first to avoid double encoding when path contains encoded characters
  const decodedUrl = decodeURIComponent(url);
  urlParams.set('path', decodedUrl);
  const res = await fetchURL(`/api/usage?${urlParams.toString()}`, { signal }, true, scope);

  try {
    return await res.json();
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name == "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }
}
