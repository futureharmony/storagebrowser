import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  isMobile,
  isSmallMobile,
  isVerySmallMobile,
  responsiveClass,
} from "./responsive";

describe("responsive utilities", () => {
  beforeEach(() => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(800);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("isMobile returns false for desktop width", () => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(800);
    expect(isMobile.value).toBe(false);
  });

  it("isMobile returns true for mobile width", () => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(600);
    expect(isMobile.value).toBe(true);
  });

  it("isSmallMobile returns true for small mobile width", () => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(400);
    expect(isSmallMobile.value).toBe(true);
  });

  it("isVerySmallMobile returns true for very small mobile width", () => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(350);
    expect(isVerySmallMobile.value).toBe(true);
  });

  it("responsiveClass returns correct classes", () => {
    vi.spyOn(window, "innerWidth", "get").mockReturnValue(600);
    expect(responsiveClass("desktop", "mobile")).toBe("mobile");
  });
});
