import { fetchURL, fetchJSON } from "./utils";

export interface Scope {
  name: string;
  rootPrefix: string;
}

export interface Bucket {
  name: string;
}

export interface BucketSettings {
  name: string;
  versioning: boolean;
  objectLock: boolean;
  objectLockDays: number;
  retentionMode: string;
}

export async function list(): Promise<Bucket[]> {
  const response = await fetchJSON<Bucket[]>(`/api/buckets`, {});
  return response;
}

export async function create(settings: BucketSettings): Promise<Bucket> {
  await fetchURL(`/api/buckets`, {
    method: "POST",
    body: JSON.stringify(settings),
  });
  return { name: settings.name };
}

export async function remove(name: string): Promise<void> {
  await fetchURL(`/api/buckets/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
}

export async function getSettings(name: string): Promise<BucketSettings> {
  const response = await fetchJSON<BucketSettings>(
    `/api/buckets/${encodeURIComponent(name)}`,
    {}
  );
  return response;
}

export async function updateSettings(
  settings: BucketSettings
): Promise<BucketSettings> {
  const response = await fetchJSON<BucketSettings>(
    `/api/buckets/${encodeURIComponent(settings.name)}`,
    {
      method: "PUT",
      body: JSON.stringify(settings),
    }
  );
  return response;
}
