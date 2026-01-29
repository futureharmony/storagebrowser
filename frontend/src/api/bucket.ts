import { fetchURL, fetchJSON } from "./utils";

export interface Scope {
  name: string;
  rootPrefix: string;
}

export interface BucketListResponse {
  availableScopes: Scope[];
  currentScope: Scope;
}

export interface Bucket {
  name: string;
}

export async function list(): Promise<Bucket[]> {
  const response = await fetchJSON<BucketListResponse>(`/api/buckets`, {});
  // Convert availableScopes to Bucket array for backward compatibility
  return response.availableScopes.map(scope => ({ name: scope.name }));
}

export async function listWithScopes(): Promise<BucketListResponse> {
  return await fetchJSON<BucketListResponse>(`/api/buckets`, {});
}
