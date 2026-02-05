import { files as api } from "@/api";
import { useAuthStore } from "@/stores/auth";

// 核心接口定义
export interface ConflictResult {
  hasConflict: boolean;
  existingItems: string[];
  suggestedNewName?: string;
  duplicateNames: string[];
}

export interface ResolutionOptions {
  overwrite?: boolean;
  rename?: boolean;
  customName?: string;
}

export interface ConflictResolution {
  resolvedPaths: string[];
  action: 'overwrite' | 'rename' | 'skip';
}

// 资源项类型定义（与现有代码兼容）
export interface ResourceItem {
  name: string;
  isDir: boolean;
  url?: string;
  [key: string]: any;
}

// 待处理项类型定义
export interface ItemToProcess {
  from: string;
  to: string;
  name: string;
  [key: string]: any;
}

/**
 * 冲突检测服务类
 */
class ConflictDetector {
  /**
   * 检查源路径与目标路径之间的冲突
   * @param sourcePaths 源路径数组
   * @param destPath 目标路径
   * @param existingItems 目标路径下已存在的资源项
   * @returns 冲突检测结果
   */
  async checkConflict(
    sourcePaths: ItemToProcess[],
    destPath: string,
    existingItems: ResourceItem[]
  ): Promise<ConflictResult> {
    const existingNames = new Set<string>();
    for (const item of existingItems) {
      existingNames.add(item.name);
    }

    const duplicateNames: string[] = [];
    for (const item of sourcePaths) {
      if (existingNames.has(item.name)) {
        duplicateNames.push(item.name);
      }
    }

    const hasConflict = duplicateNames.length > 0;
    const result: ConflictResult = {
      hasConflict,
      existingItems: Array.from(existingNames),
      duplicateNames,
    };

    if (hasConflict && duplicateNames.length === 1) {
      result.suggestedNewName = this.generateNewName(duplicateNames[0], Array.from(existingNames));
    }

    return result;
  }

  /**
   * 解决冲突
   * @param conflict 冲突检测结果
   * @param options 解决选项
   * @returns 冲突解决结果
   */
  async resolveConflict(
    conflict: ConflictResult,
    options: ResolutionOptions
  ): Promise<ConflictResolution> {
    const { overwrite = false, rename = false, customName } = options;

    if (overwrite) {
      return {
        resolvedPaths: conflict.duplicateNames,
        action: 'overwrite'
      };
    }

    if (rename) {
      const renamedPaths = conflict.duplicateNames.map(name => {
        return customName || this.generateNewName(name, conflict.existingItems);
      });

      return {
        resolvedPaths: renamedPaths,
        action: 'rename'
      };
    }

    return {
      resolvedPaths: [],
      action: 'skip'
    };
  }

  /**
   * 生成不重复的新名称
   * @param originalName 原始名称
   * @param existingNames 已存在的名称数组
   * @returns 不重复的新名称
   */
  generateNewName(originalName: string, existingNames: string[]): string {
    const namePattern = /^(.+?)(?:\s*\((\d+)\))?(\.[^.]+)?$/;
    const match = originalName.match(namePattern);

    if (!match) {
      return this.findAvailableName(originalName, existingNames);
    }

    const [, baseName, versionStr, extension] = match;
    let version = versionStr ? parseInt(versionStr, 10) : 0;

    let newName: string;
    do {
      version++;
      newName = `${baseName} (${version})${extension || ''}`;
    } while (existingNames.includes(newName));

    return newName;
  }

  /**
   * 查找可用的名称（简单版本）
   * @param originalName 原始名称
   * @param existingNames 已存在的名称数组
   * @returns 可用的新名称
   */
  private findAvailableName(originalName: string, existingNames: string[]): string {
    let version = 1;
    let newName: string;

    do {
      newName = `${originalName} (${version})`;
      version++;
    } while (existingNames.includes(newName));

    return newName;
  }

  /**
   * 获取目标路径下的现有资源项
   * @param destPath 目标路径
   * @returns 现有资源项数组
   */
  async getExistingItems(destPath: string): Promise<ResourceItem[]> {
    const authStore = useAuthStore();
    const scope = authStore.user?.currentScope?.name;

    // 对于S3 bucket路径，需要移除/buckets/test1/前缀
    let fetchPath = destPath;
    if (scope && fetchPath.match(/^\/buckets\/[^/]+/)) {
      const bucketMatch = fetchPath.match(/^\/buckets\/([^/]+)/);
      if (bucketMatch) {
        fetchPath = fetchPath.slice(bucketMatch[0].length) || '/';
      }
    }

    try {
      const data = await api.fetch(fetchPath, undefined, scope);
      return data.items || [];
    } catch (error) {
      console.error('Failed to fetch existing items:', error);
      return [];
    }
  }

  /**
   * 检查并解决冲突的综合方法
   * @param sourcePaths 源路径数组
   * @param destPath 目标路径
   * @param options 解决选项
   * @returns 冲突检测和解决结果
   */
  async checkAndResolveConflict(
    sourcePaths: ItemToProcess[],
    destPath: string,
    options: ResolutionOptions = {}
  ): Promise<{ conflict: ConflictResult; resolution?: ConflictResolution }> {
    const existingItems = await this.getExistingItems(destPath);
    const conflict = await this.checkConflict(sourcePaths, destPath, existingItems);

    if (conflict.hasConflict && (options.overwrite || options.rename || options.customName)) {
      const resolution = await this.resolveConflict(conflict, options);
      return { conflict, resolution };
    }

    return { conflict };
  }
}

// 创建单例实例
const conflictDetector = new ConflictDetector();

export default conflictDetector;
