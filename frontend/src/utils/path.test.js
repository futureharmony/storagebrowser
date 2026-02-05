import { fileURLToPath } from "url";
import { dirname, join } from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// 使用动态导入
import(join(__dirname, "/path.ts"))
  .then(
    ({
      normalizePath,
      isS3BucketPath,
      extractS3BucketName,
      convertToS3Path,
      convertFromS3Path,
      getParentPath,
      getRelativePath,
      stripS3BucketPrefix,
    }) => {
      // 测试 normalizePath 函数
      console.log("=== normalizePath 测试 ===");
      console.log(normalizePath("/")); // 应该是 "/"
      console.log(normalizePath("/files")); // 应该是 "/files"
      console.log(normalizePath("/files/")); // 应该是 "/files"
      console.log(normalizePath("files/")); // 应该是 "/files"
      console.log(normalizePath("")); // 应该是 "/"
      console.log("");

      // 测试 isS3BucketPath 函数
      console.log("=== isS3BucketPath 测试 ===");
      console.log(isS3BucketPath("/buckets/test1")); // 应该是 true
      console.log(isS3BucketPath("/buckets/test1/")); // 应该是 true
      console.log(isS3BucketPath("/buckets/test1/files")); // 应该是 true
      console.log(isS3BucketPath("/files")); // 应该是 false
      console.log("");

      // 测试 extractS3BucketName 函数
      console.log("=== extractS3BucketName 测试 ===");
      console.log(extractS3BucketName("/buckets/test1")); // 应该是 "test1"
      console.log(extractS3BucketName("/buckets/test1/")); // 应该是 "test1"
      console.log(extractS3BucketName("/buckets/test1/files")); // 应该是 "test1"
      console.log(extractS3BucketName("/files")); // 应该是 null
      console.log("");

      // 测试 convertToS3Path 函数
      console.log("=== convertToS3Path 测试 ===");
      console.log(convertToS3Path("test1", "/")); // 应该是 "/buckets/test1"
      console.log(convertToS3Path("test1", "/files")); // 应该是 "/buckets/test1/files"
      console.log(convertToS3Path("test1", "/files/")); // 应该是 "/buckets/test1/files"
      console.log(convertToS3Path("test1", "files")); // 应该是 "/buckets/test1/files"
      console.log("");

      // 测试 convertFromS3Path 函数
      console.log("=== convertFromS3Path 测试 ===");
      console.log(convertFromS3Path("/buckets/test1")); // 应该是 { scope: "test1", path: "/" }
      console.log(convertFromS3Path("/buckets/test1/")); // 应该是 { scope: "test1", path: "/" }
      console.log(convertFromS3Path("/buckets/test1/files")); // 应该是 { scope: "test1", path: "/files" }
      console.log(convertFromS3Path("/buckets/test1/files/")); // 应该是 { scope: "test1", path: "/files" }
      console.log("");

      // 测试 getParentPath 函数
      console.log("=== getParentPath 测试 ===");
      console.log(getParentPath("/")); // 应该是 null
      console.log(getParentPath("/files")); // 应该是 "/"
      console.log(getParentPath("/files/")); // 应该是 "/"
      console.log(getParentPath("/files/documents")); // 应该是 "/files"
      console.log(getParentPath("/buckets/test1")); // 应该是 "/buckets"
      console.log(getParentPath("/buckets/test1/")); // 应该是 "/buckets"
      console.log(getParentPath("/buckets/test1/files")); // 应该是 "/buckets/test1"
      console.log("");

      // 测试 getRelativePath 函数
      console.log("=== getRelativePath 测试 ===");
      console.log(getRelativePath("/", "/")); // 应该是 ""
      console.log(getRelativePath("/", "/files")); // 应该是 "files"
      console.log(getRelativePath("/files", "/files")); // 应该是 ""
      console.log(getRelativePath("/files", "/files/documents")); // 应该是 "documents"
      console.log(getRelativePath("/files/documents", "/files")); // 应该是 ".."
      console.log(getRelativePath("/files/documents", "/files/images")); // 应该是 "../images"
      console.log(
        getRelativePath(
          "/buckets/test1/files",
          "/buckets/test1/files/documents"
        )
      ); // 应该是 "documents"
      console.log("");

      // 测试 stripS3BucketPrefix 函数
      console.log("=== stripS3BucketPrefix 测试 ===");
      console.log(stripS3BucketPrefix("/buckets/test1", "test1")); // 应该是 "/"
      console.log(stripS3BucketPrefix("/buckets/test1/", "test1")); // 应该是 "/"
      console.log(stripS3BucketPrefix("/buckets/test1/files", "test1")); // 应该是 "/files"
      console.log(stripS3BucketPrefix("/buckets/test1/files/", "test1")); // 应该是 "/files"
      console.log(stripS3BucketPrefix("/buckets/test2/files", "test1")); // 应该是 "/files"
      console.log(stripS3BucketPrefix("/files", "test1")); // 应该是 "/files"
      console.log(stripS3BucketPrefix("", "test1")); // 应该是 "/"

      console.log("\n✅ 所有测试完成！");
    }
  )
  .catch((err) => {
    console.error("测试失败:", err);
    process.exit(1);
  });
