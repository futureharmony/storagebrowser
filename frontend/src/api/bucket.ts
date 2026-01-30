import { fetchURL, fetchJSON } from "./utils";

export interface Scope {
  name: string;
  rootPrefix: string;
}

export interface Bucket {
  name: string;
}

export async function list(): Promise<Bucket[]> {
  const response = await fetchJSON<Bucket[]>(`/api/buckets`, {});
  return response;
}

