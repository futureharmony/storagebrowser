import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { baseURL } from "@/utils/constants";
import { upload as postTus, useTus } from "./tus";
import { createURL, fetchURL, removePrefix, StatusError } from "./utils";

export async function fetch(url: string, signal?: AbortSignal, scope?: string) {
  console.log('[API FETCH] original url:', url);

  const appConfig = (window as any).FileBrowser || {};

  // Process URL based on scope and storage type
  if (appConfig.StorageType === "s3") {
    if (scope) {
      // For S3 storage with scope, the URL is already processed in Files.vue
      // URL is already in the correct format (e.g., "/folder")
      // No need to call removePrefix
    } else {
      url = removePrefix(url);
      // For S3 storage without scope, strip the bucket name from the URL
      const bucketMatch = url.match(/^\/([^/]+)/);
      if (bucketMatch) {
        url = url.slice(bucketMatch[0].length) || '/';
      }
    }
  } else {
    // For non-S3 storage, always remove prefix
    url = removePrefix(url);
  }

  const res = await fetchURL(`/api/resources${url}`, { signal }, true, scope);

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
  // Extract query parameters before processing the path
  const [pathWithoutQuery, queryPart] = url.split('?');
  let processedPath = pathWithoutQuery;

  // For S3 storage with scope, the path is already processed (doesn't include bucket)
  // For non-S3 storage or without scope, apply removePrefix
  const appConfig = (window as any).FileBrowser || {};
  if (appConfig.StorageType === "s3" && scope) {
    // Path is already processed in moveCopy, no need to remove bucket/prefix
    processedPath = pathWithoutQuery;
  } else {
    processedPath = removePrefix(pathWithoutQuery);
    
    // For S3 storage without scope, strip the bucket name from the URL
    if (appConfig.StorageType === "s3") {
      const bucketMatch = processedPath.match(/^\/([^/]+)/);
      if (bucketMatch) {
        processedPath = processedPath.slice(bucketMatch[0].length) || '/';
      }
    }
  }

  // Reconstruct URL with processed path and original query parameters
  const finalUrl = queryPart ? `${processedPath}?${queryPart}` : processedPath;

  const opts: ApiOpts = {
    method,
  };

  if (content) {
    opts.body = content;
  }

  const res = await fetchURL(`/api/resources${finalUrl}`, opts, true, scope);

  return res;
}

export async function remove(url: string, scope?: string) {
  return resourceAction(url, "DELETE", undefined, scope);
}

export async function put(url: string, content = "", scope?: string) {
  return resourceAction(url, "PUT", content, scope);
}

export function download(format: any, ...files: string[]) {
  let url = `${baseURL}/api/raw`;

  if (files.length === 1) {
    url += removePrefix(files[0]) + "?";
  } else {
    let arg = "";

    for (const file of files) {
      arg += removePrefix(file) + ",";
    }

    arg = arg.substring(0, arg.length - 1);
    arg = encodeURIComponent(arg);
    url += `/?files=${arg}&`;
  }

  if (format) {
    url += `algo=${format}&`;
  }

  window.open(url);
}

export async function post(
  url: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any = () => {}
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
    const bucketMatch = url.match(/^\/([^/]+)/);
    if (bucketMatch) {
      url = url.slice(bucketMatch[0].length) || '/';
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
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open(
      "POST",
      `${baseURL}/api/resources${url}?override=${overwrite}`,
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
    
    const url = `${processedFrom}?action=${
      copy ? "copy" : "rename"
    }&destination=${encodeURIComponent(processedTo)}&override=${overwrite}&rename=${rename}`;
    promises.push(resourceAction(url, "PATCH", undefined, scope));
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

export async function checksum(url: string, algo: ChecksumAlg) {
  const data = await resourceAction(`${url}?checksum=${algo}`, "GET");
  return (await data.json()).checksums[algo];
}

export function getDownloadURL(file: ResourceItem, inline: any) {
  const params = {
    ...(inline && { inline: "true" }),
  };

  return createURL("api/raw" + file.path, params);
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
      const bucketMatch = url.match(/^\/([^/]+)/);
      if (bucketMatch) {
        url = url.slice(bucketMatch[0].length) || '/';
      }
    }
  } else {
    // For non-S3 storage, always remove prefix
    url = removePrefix(url);
  }

  const res = await fetchURL(`/api/usage${url}`, { signal }, true, scope);

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
