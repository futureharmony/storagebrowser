/**
 * 路径处理工具函数模块
 * 统一处理应用中的路径标准化、S3存储桶路径转换等逻辑
 */

/**
 * 标准化路径，确保路径格式一致
 * - 去除末尾斜杠（根路径 "/" 除外）
 * - 确保路径以斜杠开头
 */
export function normalizePath(path: string): string {
  if (!path) return "/";

  // 确保路径以斜杠开头
  let normalized = path.startsWith("/") ? path : `/${path}`;

  // 去除末尾斜杠，但保留根路径 "/"
  if (normalized.length > 1 && normalized.endsWith("/")) {
    normalized = normalized.slice(0, -1);
  }

  return normalized;
}

/**
 * 判断是否为S3存储桶路径
 * 匹配格式: /buckets/{bucketName}/...
 */
export function isS3BucketPath(path: string): boolean {
  if (!path) return false;
  return /^\/buckets\/[^/]+/.test(path);
}

/**
 * 从S3路径中提取存储桶名称
 * @param path S3路径，格式为 /buckets/{bucketName}/...
 * @returns 存储桶名称或null
 */
export function extractS3BucketName(path: string): string | null {
  if (!path) return null;

  const bucketMatch = path.match(/^\/buckets\/([^/]+)/);
  return bucketMatch ? bucketMatch[1] : null;
}

/**
 * 将本地路径转换为S3格式路径
 * @param scope 存储桶名称（scope）
 * @param path 本地路径（相对于存储桶的路径）
 * @returns S3格式路径，格式为 /buckets/{scope}{path}
 */
export function convertToS3Path(scope: string, path: string): string {
  if (!scope) return path;

  // 标准化路径
  const normalizedPath = normalizePath(path);

  // 构建S3路径
  return `/buckets/${scope}${normalizedPath === "/" ? "" : normalizedPath}`;
}

/**
 * 从S3路径转换回本地路径格式
 * @param s3Path S3路径，格式为 /buckets/{bucketName}/...
 * @returns 包含scope和本地路径的对象
 */
export function convertFromS3Path(s3Path: string): {
  scope: string | null;
  path: string;
} {
  const bucketMatch = s3Path.match(/^\/buckets\/([^/]+)/);

  if (bucketMatch) {
    const scope = bucketMatch[1];
    const path = s3Path.slice(bucketMatch[0].length) || "/";
    return { scope, path };
  }

  return { scope: null, path: s3Path };
}

/**
 * 获取父路径
 * @param path 当前路径
 * @returns 父路径或null（如果已经是根路径）
 */
export function getParentPath(path: string): string | null {
  if (!path || path === "/" || path === "/buckets") return null;

  // 处理S3存储桶根路径的情况
  if (path.match(/^\/buckets\/[^/]+(\/)?$/)) {
    return "/buckets";
  }

  const normalizedPath = normalizePath(path);
  const lastSlashIndex = normalizedPath.lastIndexOf("/");

  if (lastSlashIndex === 0) return "/";

  return normalizedPath.slice(0, lastSlashIndex);
}

/**
 * 获取相对路径
 * @param base 基准路径
 * @param target 目标路径
 * @returns 从基准路径到目标路径的相对路径
 */
export function getRelativePath(base: string, target: string): string {
  const normalizedBase = normalizePath(base);
  const normalizedTarget = normalizePath(target);

  // 如果路径相同，返回空字符串
  if (normalizedBase === normalizedTarget) {
    return "";
  }

  // 如果基准路径是根路径
  if (normalizedBase === "/") {
    return normalizedTarget.slice(1);
  }

  // 检查目标路径是否以基准路径开头
  if (normalizedTarget.startsWith(normalizedBase)) {
    const relative = normalizedTarget.slice(normalizedBase.length);
    return relative.startsWith("/") ? relative.slice(1) : relative;
  }

  // 复杂情况：计算相对路径（处理不同的目录层级）
  const baseParts = normalizedBase.split("/").filter(Boolean);
  const targetParts = normalizedTarget.split("/").filter(Boolean);

  let commonLength = 0;
  while (
    commonLength < baseParts.length &&
    commonLength < targetParts.length &&
    baseParts[commonLength] === targetParts[commonLength]
  ) {
    commonLength++;
  }

  const upLevels = baseParts.length - commonLength;
  const downParts = targetParts.slice(commonLength);

  const relativeParts = [];
  for (let i = 0; i < upLevels; i++) {
    relativeParts.push("..");
  }
  relativeParts.push(...downParts);

  return relativeParts.join("/");
}

/**
 * 处理S3路径，去除存储桶前缀（用于API请求）
 * @param path 原始路径
 * @param scope 当前scope（存储桶名称）
 * @returns 处理后的路径（去除存储桶前缀）
 */
export function stripS3BucketPrefix(path: string, scope?: string): string {
  if (!path) return "/";

  // 如果有scope，并且路径是该存储桶的路径，则去除前缀
  if (scope && path.startsWith(`/buckets/${scope}`)) {
    return path.slice(`/buckets/${scope}`.length) || "/";
  }

  // 如果是S3路径但没有指定scope，或者scope不匹配，则检查是否有其他存储桶前缀
  const bucketMatch = path.match(/^\/buckets\/([^/]+)/);
  if (bucketMatch) {
    return path.slice(bucketMatch[0].length) || "/";
  }

  return path;
}
