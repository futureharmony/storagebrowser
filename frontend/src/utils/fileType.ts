/**
 * 根据文件扩展名检测文件类型
 * @param filename 文件名
 * @returns 文件类型字符串，用于CSS data-type属性
 */
export function detectFileType(filename: string): string {
  const ext = getExtension(filename).toLowerCase();

  // 图片类型
  const imageExts = [
    ".png",
    ".jpg",
    ".jpeg",
    ".gif",
    ".bmp",
    ".webp",
    ".svg",
    ".ico",
    ".tiff",
    ".tif",
  ];
  if (imageExts.includes(ext)) return "image";

  // PDF类型
  if (ext === ".pdf") return "pdf";

  // 视频类型
  const videoExts = [
    ".mp4",
    ".avi",
    ".mov",
    ".wmv",
    ".flv",
    ".mkv",
    ".webm",
    ".m4v",
    ".mpg",
    ".mpeg",
  ];
  if (videoExts.includes(ext)) return "video";

  // 音频类型
  const audioExts = [".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a", ".wma"];
  if (audioExts.includes(ext)) return "audio";

  // 文本类型
  const textExts = [
    ".txt",
    ".md",
    ".log",
    ".ini",
    ".cfg",
    ".conf",
    ".json",
    ".xml",
    ".yaml",
    ".yml",
    ".csv",
  ];
  if (textExts.includes(ext)) return "text";

  // 代码类型
  const codeExts = [
    ".js",
    ".ts",
    ".jsx",
    ".tsx",
    ".vue",
    ".html",
    ".css",
    ".scss",
    ".less",
    ".py",
    ".java",
    ".cpp",
    ".c",
    ".go",
    ".rs",
    ".php",
    ".rb",
    ".sh",
    ".bat",
    ".ps1",
  ];
  if (codeExts.includes(ext)) return "text";

  // 文档类型
  const docExts = [".doc", ".docx", ".odt", ".rtf", ".pages"];
  if (docExts.includes(ext)) return "text";

  // 表格类型
  const spreadsheetExts = [".xls", ".xlsx", ".ods", ".csv"];
  if (spreadsheetExts.includes(ext)) return "text";

  // 演示文稿类型
  const presentationExts = [".ppt", ".pptx", ".odp", ".key"];
  if (presentationExts.includes(ext)) return "text";

  // 压缩文件类型
  const archiveExts = [
    ".zip",
    ".rar",
    ".7z",
    ".tar",
    ".gz",
    ".bz2",
    ".xz",
    ".zst",
    ".cab",
  ];
  if (archiveExts.includes(ext)) return "blob";

  // 安装包类型
  const packageExts = [".pkg", ".deb", ".rpm", ".msi", ".exe", ".dmg", ".apk"];
  if (packageExts.includes(ext)) return "blob";

  // 默认返回 blob 类型
  return "blob";
}

/**
 * 获取文件扩展名
 * @param filename 文件名
 * @returns 文件扩展名（包括点号）
 */
export function getExtension(filename: string): string {
  const lastDotIndex = filename.lastIndexOf(".");
  if (lastDotIndex === -1) {
    return "";
  }
  return filename.substring(lastDotIndex);
}
