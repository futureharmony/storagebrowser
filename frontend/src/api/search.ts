import { fetchURL, removePrefix } from "./utils";
import url from "../utils/url";
import { stripS3BucketPrefix } from "@/utils/path";

export default async function search(base: string, query: string) {
  // Get storage type from global config
  const appConfig = (window as any).FileBrowser || {};
  const isS3 = appConfig.StorageType === "s3";

  let scope: string | undefined;
  let path: string;

  if (isS3) {
    // For S3 storage, extract scope from /buckets/{scope}/ pattern
    base = removePrefix(base);
    const bucketMatch = base.match(/^\/buckets\/([^/]+)/);

    if (bucketMatch) {
      scope = bucketMatch[1];
      // Strip the bucket prefix to get the actual path
      path = stripS3BucketPrefix(base, scope);
    } else {
      // Not a bucket path, use as-is
      path = base;
    }
  } else {
    // For local storage
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

  // Determine correct base for URLs based on storage type
  let urlBase: string;
  if (isS3 && scope) {
    // For S3 with scope, use bucket path
    urlBase = `/buckets/${scope}${path}`;
  } else {
    // For local storage, use the original base (with /files prefix added)
    urlBase = `/files${base}`;
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
