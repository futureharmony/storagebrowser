# Bucket Configuration UI & CRUD Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix bucket CRUD operations and simplify the configuration UI while maintaining all functionality.

**Architecture:** 
- Frontend: Fix unit conversion in Buckets.vue, add proper error handling with toast messages, add admin permission check
- Backend: Already complete, ensure proper error messages returned

**Tech Stack:** Vue.js 3 + TypeScript (frontend), Go (backend)

---

## Task 1: Fix Unit Conversion in Bucket Creation

**Files:**
- Modify: `frontend/src/views/settings/Buckets.vue:427-441`

**Step 1: Read current createBucket function**

The current function (lines 427-441) doesn't use `quotaUnit` and `retentionUnit`:

```typescript
const createBucket = async () => {
  if (!trimmedBucketName.value || isCreateDisabled.value) return;
  creating.value = true;
  try {
    const settings = { ...newBucketSettings.value, name: trimmedBucketName.value };
    await create(settings);
    closeCreateModal();
    await loadBuckets();
    $showSuccess(t("settings.bucketCreated"));
  } catch (err) {
    console.error(err);
  } finally {
    creating.value = false;
  }
};
```

**Step 2: Add unit conversion logic**

Replace the createBucket function with:

```typescript
const createBucket = async () => {
  if (!trimmedBucketName.value || isCreateDisabled.value) return;
  creating.value = true;
  try {
    let quotaStorageGB = newBucketSettings.value.quotaStorageGB;
    if (quotaEnabled.value) {
      const unitMultiplier: Record<string, number> = {
        MiB: 1 / 1024,
        GiB: 1,
        TiB: 1024,
        PiB: 1024 * 1024,
      };
      quotaStorageGB = Math.floor(quotaStorageGB * (unitMultiplier[quotaUnit.value] || 1));
    }

    let objectLockDays = newBucketSettings.value.objectLockDays;
    if (retentionEnabled.value && retentionUnit.value === "year") {
      objectLockDays = objectLockDays * 365;
    }

    const settings = {
      ...newBucketSettings.value,
      name: trimmedBucketName.value,
      quotaStorageGB,
      objectLockDays,
    };
    await create(settings);
    closeCreateModal();
    await loadBuckets();
    $showSuccess(t("settings.bucketCreated"));
  } catch (err) {
    console.error(err);
    $showError(err instanceof Error ? err.message : t("error.createBucketFailed"));
  } finally {
    creating.value = false;
  }
};
```

**Step 3: Verify typecheck**

Run: `cd frontend && pnpm run typecheck`
Expected: PASS (no new errors)

---

## Task 2: Add Error Handling with Toast Messages

**Files:**
- Modify: `frontend/src/views/settings/Buckets.vue:436-440, 469-473, 488-490`

**Step 1: Add error toast function**

Add to script section (around line 314 after $showSuccess):

```typescript
interface IToastError {
  (message: string): void;
}

const $showError = inject<IToastError>("$showError")!;
```

**Step 2: Update createBucket error handling**

Change lines 436-440 from:
```typescript
  } catch (err) {
    console.error(err);
  }
```

To:
```typescript
  } catch (err) {
    console.error(err);
    $showError(err instanceof Error ? err.message : t("error.createBucketFailed"));
  }
```

**Step 3: Update saveSettings error handling**

Change lines 469-473 from:
```typescript
  } catch (err) {
    console.error(err);
  }
```

To:
```typescript
  } catch (err) {
    console.error(err);
    $showError(err instanceof Error ? err.message : t("error.updateBucketFailed"));
  }
```

**Step 4: Update deleteBucket error handling**

Change lines 488-490 from:
```typescript
  } catch (err) {
    console.error(err);
  }
```

To:
```typescript
  } catch (err) {
    console.error(err);
    $showError(err instanceof Error ? err.message : t("error.deleteBucketFailed"));
  }
```

**Step 5: Verify typecheck**

Run: `cd frontend && pnpm run typecheck`
Expected: PASS

---

## Task 3: Add Admin Permission Check

**Files:**
- Modify: `frontend/src/views/settings/Buckets.vue:302-309`

**Step 1: Add auth store import**

Add after line 302:
```typescript
import { useAuthStore } from "@/stores/auth";
```

**Step 2: Add auth store usage**

Add after line 324:
```typescript
const authStore = useAuthStore();
const isAdmin = computed(() => authStore.user?.perm?.admin ?? false);
```

**Step 3: Update template to check admin**

Update line 17-20 to:
```vue
<button class="button" @click="showCreateModal = true" :disabled="!isAdmin">
  <i class="material-icons">add</i>
  {{ t("settings.createBucket") }}
</button>
```

Update line 62-71 to show settings/delete buttons only for admin:
```vue
<button class="icon-button" @click="editBucket(bucket.name)" :title="t('settings.settings')" v-if="isAdmin">
  <i class="material-icons">settings</i>
</button>
<button
  class="icon-button icon-danger"
  @click="confirmDelete(bucket.name)"
  :title="t('buttons.delete')"
  v-if="isAdmin"
>
  <i class="material-icons">delete</i>
</button>
```

**Step 4: Verify typecheck**

Run: `cd frontend && pnpm run typecheck`
Expected: PASS

---

## Task 4: Simplify UI - Consolidate Form Sections

**Files:**
- Modify: `frontend/src/views/settings/Buckets.vue:86-212`

**Step 1: Read current create modal structure**

The current modal has nested sections with collapsible options. Simplify by:
- Remove nested "settings-section" divs
- Flatten the form layout
- Use more intuitive toggle labels

**Step 2: Simplify template (lines 86-212)**

Replace the complex nested structure with a simpler flat layout:

```vue
<div v-if="showCreateModal" class="modal-wrapper" @click.self="closeCreateModal">
  <div class="modal modal-lg">
    <div class="modal-header">
      <h3>{{ t("settings.createBucket") }}</h3>
      <button class="icon-button" @click="closeCreateModal">
        <i class="material-icons">close</i>
      </button>
    </div>
    <div class="modal-content">
      <div class="form-group">
        <label>{{ t("settings.bucketName") }} *</label>
        <input
          v-model="newBucketSettings.name"
          type="text"
          :placeholder="t('settings.bucketNamePlaceholder')"
          :class="{ 'input-error': showNameError }"
          @keyup.enter="createBucket"
        />
        <span v-if="showNameError" class="error-text">
          {{ t("settings.bucketNameError") }}
        </span>
      </div>

      <div class="form-row-simple">
        <div class="toggle-item">
          <label>{{ t("settings.versioning") }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="newBucketSettings.versioning" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="toggle-item">
          <label>{{ t("settings.objectLock") }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="newBucketSettings.objectLock" @change="onObjectLockChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="toggle-item" v-if="newBucketSettings.objectLock">
          <label>{{ t("settings.retention") }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="retentionEnabled" />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="toggle-item">
          <label>{{ t("settings.bucketQuota") }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="quotaEnabled" />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>

      <div v-if="newBucketSettings.objectLock && retentionEnabled" class="retention-options">
        <div class="form-row">
          <div class="form-group">
            <label>{{ t("settings.retentionMode") }}</label>
            <div class="radio-group">
              <label class="radio-item" :class="{ selected: newBucketSettings.retentionMode === 'COMPLIANCE' }">
                <input type="radio" v-model="newBucketSettings.retentionMode" value="COMPLIANCE" />
                <span>{{ t("settings.compliance") }}</span>
              </label>
              <label class="radio-item" :class="{ selected: newBucketSettings.retentionMode === 'GOVERNANCE' }">
                <input type="radio" v-model="newBucketSettings.retentionMode" value="GOVERNANCE" />
                <span>{{ t("settings.governance") }}</span>
              </label>
            </div>
          </div>
          <div class="form-group">
            <label>{{ t("settings.validity") }}</label>
            <div class="input-with-unit">
              <input v-model.number="newBucketSettings.objectLockDays" type="number" min="1" />
              <select v-model="retentionUnit">
                <option value="day">{{ t("settings.days") }}</option>
                <option value="year">{{ t("settings.years") }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div v-if="quotaEnabled" class="quota-options">
        <div class="form-group">
          <label>{{ t("settings.quotaSize") }}</label>
          <div class="input-with-unit">
            <input v-model.number="newBucketSettings.quotaStorageGB" type="number" min="1" />
            <select v-model="quotaUnit">
              <option value="MiB">MiB</option>
              <option value="GiB">GiB</option>
              <option value="TiB">TiB</option>
              <option value="PiB">PiB</option>
            </select>
          </div>
        </div>
      </div>
    </div>
    <div class="modal-footer">
      <button class="button button-secondary" @click="closeCreateModal">
        {{ t("buttons.cancel") }}
      </button>
      <button class="button" @click="createBucket" :disabled="isCreateDisabled || creating">
        <span v-if="creating" class="spinner"></span>
        {{ t("buttons.create") }}
      </button>
    </div>
  </div>
</div>
```

**Step 3: Add simplified CSS styles**

Add to style section:
```css
.form-row-simple {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  padding: 1rem 0;
  border-top: 1px solid var(--border-color);
  border-bottom: 1px solid var(--border-color);
}

.toggle-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.toggle-item label:first-child {
  margin-bottom: 0;
}
```

**Step 4: Verify typecheck**

Run: `cd frontend && pnpm run typecheck`
Expected: PASS

---

## Task 5: Add Error Messages to i18n

**Files:**
- Modify: `frontend/src/i18n/en.json`
- Modify: `frontend/src/i18n/zh-cn.json`

**Step 1: Add error messages to en.json**

Add to the "error" section (around line 125):
```json
"createBucketFailed": "Failed to create bucket. Please check your permissions and try again.",
"updateBucketFailed": "Failed to update bucket settings.",
"deleteBucketFailed": "Failed to delete bucket. The bucket may contain objects.",
```

**Step 2: Add error messages to zh-cn.json**

Add to the "error" section:
```json
"createBucketFailed": "创建存储桶失败，请检查权限后重试。",
"updateBucketFailed": "更新存储桶设置失败。",
"deleteBucketFailed": "删除存储桶失败，存储桶可能包含对象。",
```

---

## Task 6: Backend - Improve Error Messages

**Files:**
- Modify: `http/bucket.go`

**Step 1: Update error responses to be more descriptive**

Change all `return http.StatusBadRequest, nil` to include error messages:

Line 15-16:
```go
if d.server.StorageType != "s3" {
    return http.StatusBadRequest, errors.New("bucket operations require S3 storage type")
}
```

Line 46-48:
```go
if d.server.StorageType != "s3" {
    return http.StatusBadRequest, errors.New("bucket operations require S3 storage type")
}
```

Similarly update all other handlers.

**Step 2: Run go fmt**

Run: `make fmt`
Expected: PASS

---

## Task 7: Verify Complete Workflow

**Step 1: Run frontend typecheck**

Run: `cd frontend && pnpm run typecheck`
Expected: PASS

**Step 2: Run frontend lint**

Run: `cd frontend && pnpm run lint`
Expected: PASS (or only pre-existing errors)

**Step 3: Build frontend**

Run: `cd frontend && pnpm run build`
Expected: PASS

---

## Summary

After completing all tasks:
1. ✅ Unit conversion works correctly (MiB/GiB/TiB/PiB, days/years)
2. ✅ User-friendly error messages displayed via toast
3. ✅ Admin permission check on frontend
4. ✅ Simplified and cleaner UI form
5. ✅ Better backend error messages
