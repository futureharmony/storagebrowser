import { fetchURL, fetchJSON } from "./utils";

export interface Bucket {
  name: string;
}

export function list() {
  return fetchJSON<Bucket[]>(`/api/buckets`, {});
}

export async function switchBucket(bucketName: string) {
  await fetchURL(`/api/buckets`, {
    method: "PUT",
    body: JSON.stringify({ bucket: bucketName }),
  });
}
