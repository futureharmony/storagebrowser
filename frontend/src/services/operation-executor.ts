import { files as api } from "@/api";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";
import buttons from "@/utils/buttons";

export interface OperationOptions {
  overwrite?: boolean;
  rename?: boolean;
  customNames?: Record<string, string>;
}

export interface OperationResult {
  success: boolean;
  message?: string;
  affectedItems?: string[];
  redirectPath?: string;
  error?: string;
}

export interface MoveOperation {
  sourcePaths: string[];
  destPath: string;
  items: FileItem[];
}

// 资源项类型定义（与现有代码兼容）
export interface FileItem {
  name: string;
  isDir: boolean;
  url?: string;
  [key: string]: any;
}

export class OperationExecutor {
  private authStore = useAuthStore();
  private fileStore = useFileStore();
  private layoutStore = useLayoutStore();

  /**
   * 构建 API 请求 payload
   */
  buildApiPayload(items: FileItem[], destPath: string, scope?: string): any[] {
    return items.map((item) => ({
      from: item.url,
      to: destPath + encodeURIComponent(item.name),
      name: item.name,
    }));
  }

  /**
   * 处理成功结果
   */
  handleSuccessResult(
    _result: any,
    mode: "move" | "copy",
    items: any[],
    destPath: string
  ): OperationResult {
    const affectedItems = items.map((item) => item.name);
    const firstItemPath = items[0].to;

    // 设置预选择项
    this.fileStore.preselect = removePrefix(firstItemPath);

    const operationResult: OperationResult = {
      success: true,
      message:
        mode === "move"
          ? "Move operation completed successfully"
          : "Copy operation completed successfully",
      affectedItems,
    };

    // 处理导航逻辑
    if (mode === "move") {
      operationResult.redirectPath = destPath;
    } else {
      // 复制操作：如果是同一路径，需要刷新；否则导航
      const currentPath = window.location.pathname;
      if (currentPath === destPath) {
        this.fileStore.reload = true;
      } else {
        operationResult.redirectPath = destPath;
      }
    }

    return operationResult;
  }

  /**
   * 执行移动操作
   */
  async executeMove(
    operation: MoveOperation,
    options?: OperationOptions
  ): Promise<OperationResult> {
    const { items, destPath } = operation;
    const { overwrite = false, rename = false } = options || {};

    try {
      buttons.loading("move");

      const scope = this.authStore.user?.currentScope?.name;

      // 构建 API payload
      const payload = this.buildApiPayload(items, destPath, scope);

      // 检查冲突
      let fetchPath = destPath;
      if (scope && fetchPath.match(/^\/buckets\/[^/]+/)) {
        const bucketMatch = fetchPath.match(/^\/buckets\/([^/]+)/);
        if (bucketMatch) {
          fetchPath = fetchPath.slice(bucketMatch[0].length) || "/";
        }
      }
      const dstItems = (await api.fetch(fetchPath, undefined, scope)).items;
      const conflict = upload.checkConflict(payload, dstItems);

      if (conflict && !overwrite && !rename) {
        buttons.done("move");
        return {
          success: false,
          error: "Conflict detected: File already exists at destination",
        };
      }

      // 执行移动操作
      await api.move(payload, overwrite, rename);

      buttons.success("move");

      return this.handleSuccessResult({}, "move", payload, destPath);
    } catch (error) {
      buttons.done("move");
      console.error("Move operation failed:", error);
      return {
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      };
    }
  }

  /**
   * 执行复制操作
   */
  async executeCopy(
    operation: MoveOperation,
    options?: OperationOptions
  ): Promise<OperationResult> {
    const { items, destPath } = operation;
    const { overwrite = false, rename = false } = options || {};

    try {
      buttons.loading("copy");

      const scope = this.authStore.user?.currentScope?.name;

      // 构建 API payload
      const payload = this.buildApiPayload(items, destPath, scope);

      // 检查冲突
      let fetchPath = destPath;
      if (scope && fetchPath.match(/^\/buckets\/[^/]+/)) {
        const bucketMatch = fetchPath.match(/^\/buckets\/([^/]+)/);
        if (bucketMatch) {
          fetchPath = fetchPath.slice(bucketMatch[0].length) || "/";
        }
      }
      const dstItems = (await api.fetch(fetchPath, undefined, scope)).items;
      const conflict = upload.checkConflict(payload, dstItems);

      if (conflict && !overwrite && !rename) {
        buttons.done("copy");
        return {
          success: false,
          error: "Conflict detected: File already exists at destination",
        };
      }

      // 执行复制操作
      await api.copy(payload, overwrite, rename);

      buttons.success("copy");

      return this.handleSuccessResult({}, "copy", payload, destPath);
    } catch (error) {
      buttons.done("copy");
      console.error("Copy operation failed:", error);
      return {
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      };
    }
  }

  /**
   * 检查操作是否有冲突
   */
  async checkConflict(operation: MoveOperation): Promise<boolean> {
    const { items, destPath } = operation;
    const scope = this.authStore.user?.currentScope?.name;

    const payload = this.buildApiPayload(items, destPath, scope);

    let fetchPath = destPath;
    if (scope && fetchPath.match(/^\/buckets\/[^/]+/)) {
      const bucketMatch = fetchPath.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        fetchPath = fetchPath.slice(bucketMatch[0].length) || "/";
      }
    }

    const dstItems = (await api.fetch(fetchPath, undefined, scope)).items;
    return upload.checkConflict(payload, dstItems);
  }
}

// 导出单例实例
export const operationExecutor = new OperationExecutor();

export default operationExecutor;
