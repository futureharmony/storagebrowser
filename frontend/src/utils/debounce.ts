import { debounce } from "lodash-es";

/**
 * 防抖工具函数
 * @param func 要防抖的函数
 * @param delay 延迟时间（毫秒）
 * @returns 防抖后的函数
 */
export function createDebouncedHandler<T extends (...args: any[]) => void>(
  func: T,
  delay: number = 300
): (...args: Parameters<T>) => void {
  return debounce(func, delay);
}

/**
 * 防抖的键盘事件处理器
 * @param handler 原始键盘事件处理函数
 * @param delay 延迟时间（毫秒）
 * @returns 防抖后的键盘事件处理函数
 */
export function createDebouncedKeyHandler(
  handler: (event: KeyboardEvent) => void,
  delay: number = 100
): (event: KeyboardEvent) => void {
  return debounce(handler, delay);
}