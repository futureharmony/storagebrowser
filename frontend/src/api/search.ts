import { fetchURL, removePrefix } from "./utils";
import url from "../utils/url";
import { stripS3BucketPrefix } from "@/utils/path";

export default async function search(base: string, query: string) {
  // Get storage type from global config
  const appConfig = (window as any).FileBrowser || {};
  const isS3 = appConfig.StorageType === "s3";

  let scope: string | undefined;
  let path: string;

  // First, extract bucket/scope from the base path before removing prefix
  // The base path is like /buckets/test1/path
  const bucketMatch = base.match(/^\/buckets\/([^/]+)(\/.*)?$/);

  if (bucketMatch) {
    scope = bucketMatch[1];
    const fullPath = bucketMatch[2] || "/";

    // Now remove the prefix for API call
    base = removePrefix(base);

    if (isS3) {
      // For S3 storage, strip the bucket prefix to get the actual path
      path = stripS3BucketPrefix(fullPath, scope);
    } else {
      // For local storage, use the full path without the bucket part
      // Since removePrefix already removed /buckets/, we need to get the remaining path
      path = fullPath;
    }
  } else {
    // Not a bucket path
    base = removePrefix(base);
    path = base;
  }

  // Ensure path is not empty and has proper format
  if (!path || path === "/") {
    path = "/";
  } else if (!path.startsWith("/")) {
    path = "/" + path;
  }

  // Build URL with query parameters
  const urlParams = new URLSearchParams();
  urlParams.set("path", path);
  urlParams.set("query", query);

  const res = await fetchURL(
    `/api/search?${urlParams.toString()}`,
    {},
    true,
    scope
  );

  let data = await res.json();

  // Determine correct base for URLs
  let urlBase: string;
  if (scope) {
    // Always use /buckets/{scope} prefix
    urlBase = `/buckets/${scope}${path}`;
  } else {
    // If no scope, use /buckets/ root
    urlBase = "/buckets/";
  }

  // Ensure urlBase ends with slash for directory
  if (!urlBase.endsWith("/")) {
    urlBase += "/";
  }

  data = data.map((item: ResourceItem & { dir: boolean }) => {
    item.url = urlBase + url.encodePath(item.path);

    if (item.dir) {
      item.url += "/";
    }

    return item;
  });

  return data;
}
