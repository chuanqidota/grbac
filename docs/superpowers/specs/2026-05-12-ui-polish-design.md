# UI/UX Visual Polish Design Spec

## Context

The GRBAC frontend uses Element Plus with default styling and hardcoded colors throughout. The UI looks generic and unprofessional. This spec defines a comprehensive visual polish using CSS design tokens and Element Plus theme overrides — no layout or logic changes, purely visual improvement.

**Goal**: Transform the admin panel from "default Element Plus" to a polished, professional-looking interface while keeping the existing sidebar layout and component structure intact.

**Approach**: Create a centralized design token system (CSS variables), override Element Plus theme variables globally, define shared utility classes, then apply them consistently across all views.

---

## 1. Design Tokens

Create `src/styles/tokens.css` with CSS custom properties:

```css
:root {
  /* Colors */
  --color-primary: #409eff;
  --color-primary-light: #ecf5ff;
  --color-primary-dark: #337ecc;
  --color-success: #67c23a;
  --color-warning: #e6a23c;
  --color-danger: #f56c6c;
  --color-info: #909399;

  /* Text */
  --color-text-primary: #303133;
  --color-text-regular: #606266;
  --color-text-secondary: #909399;
  --color-text-placeholder: #c0c4cc;

  /* Background */
  --color-bg-page: #f5f7fa;
  --color-bg-card: #ffffff;
  --color-bg-overlay: rgba(0, 0, 0, 0.5);

  /* Border */
  --color-border: #e4e7ed;
  --color-border-light: #ebeef5;
  --color-border-extra-light: #f2f6fc;

  /* Spacing (4px base) */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;
  --space-2xl: 48px;

  /* Shadows */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 2px 8px rgba(0, 0, 0, 0.08);
  --shadow-lg: 0 4px 16px rgba(0, 0, 0, 0.1);

  /* Border radius */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  /* Typography */
  --font-size-xs: 12px;
  --font-size-sm: 13px;
  --font-size-base: 14px;
  --font-size-lg: 16px;
  --font-size-xl: 20px;
  --font-size-2xl: 24px;

  /* Sidebar */
  --sidebar-bg: #001529;
  --sidebar-width: 210px;
  --sidebar-collapsed-width: 64px;

  /* Header */
  --header-height: 56px;
  --header-bg: #ffffff;
}
```

---

## 2. Element Plus Theme Overrides

Create `src/styles/element-overrides.css` that maps EP's CSS variables to our tokens:

```css
:root {
  /* Primary color */
  --el-color-primary: var(--color-primary);
  --el-color-primary-light-3: #79bbff;
  --el-color-primary-light-5: #a0cfff;
  --el-color-primary-light-7: #c6e2ff;
  --el-color-primary-light-8: #d9ecff;
  --el-color-primary-light-9: #ecf5ff;
  --el-color-primary-dark-2: #337ecc;

  /* Border radius */
  --el-border-radius-small: var(--radius-sm);
  --el-border-radius-base: var(--radius-sm);
  --el-border-radius-medium: var(--radius-md);
  --el-border-radius-large: var(--radius-lg);
  --el-border-radius-round: 20px;
  --el-border-radius-circle: 100%;

  /* Font */
  --el-font-size-base: var(--font-size-base);
  --el-font-size-small: var(--font-size-sm);
  --el-font-size-extra-small: var(--font-size-xs);

  /* Background */
  --el-bg-color-page: var(--color-bg-page);

  /* Border */
  --el-border-color: var(--color-border);
  --el-border-color-light: var(--color-border-light);
  --el-border-color-extra-light: var(--color-border-extra-light);

  /* Text */
  --el-text-color-primary: var(--color-text-primary);
  --el-text-color-regular: var(--color-text-regular);
  --el-text-color-secondary: var(--color-text-secondary);
  --el-text-color-placeholder: var(--color-text-placeholder);

  /* Card */
  --el-card-border-radius: var(--radius-md);

  /* Dialog */
  --el-dialog-border-radius: var(--radius-lg);

  /* Table */
  --el-table-border-color: var(--color-border-light);
  --el-table-header-bg-color: #fafafa;

  /* Menu (sidebar) */
  --el-menu-bg-color: var(--sidebar-bg);
  --el-menu-text-color: #ffffffb3;
  --el-menu-active-color: #ffffff;
  --el-menu-hover-bg-color: #ffffff1a;

  /* Drawer */
  --el-drawer-padding-primary: var(--space-lg);

  /* Shadows */
  --el-box-shadow: var(--shadow-md);
  --el-box-shadow-light: var(--shadow-sm);
}
```

---

## 3. Shared Utility Classes

Create `src/styles/utilities.css` with reusable layout classes:

```css
/* Page header row: title + actions */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-lg);
}

.page-header h2 {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--color-text-primary);
}

.page-header .header-actions {
  display: flex;
  gap: var(--space-sm);
}

/* Search/filter bar above tables */
.search-bar {
  margin-bottom: var(--space-md);
  display: flex;
  gap: var(--space-md);
  align-items: center;
}

/* Content card wrapper */
.page-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
}

/* Stat card (dashboard) */
.stat-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
  transition: box-shadow 0.3s ease, transform 0.3s ease;
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* Page content wrapper */
.page-content {
  padding: var(--space-lg);
}
```

---

## 4. File Changes Summary

### New files to create:
- `src/styles/tokens.css` — design tokens
- `src/styles/element-overrides.css` — EP theme overrides
- `src/styles/utilities.css` — shared utility classes

### Files to modify:

| File | Changes |
|---|---|
| `src/main.ts` | Import the 3 new CSS files (before element-plus CSS) |
| `src/assets/base.css` | Remove or simplify — tokens replace its role |
| `src/assets/main.css` | Import tokens instead of base.css |
| `src/layouts/DefaultLayout.vue` | Use CSS variables instead of hardcoded colors, apply `--sidebar-bg`, `--header-bg`, `--shadow-sm` etc. |
| `src/views/dashboard/DashboardView.vue` | Use `stat-card`, `page-card`, `page-header` classes |
| `src/views/login/LoginView.vue` | Polish form card with `--radius-lg`, `--shadow-lg` |
| `src/views/user/UserView.vue` | Use `page-header`, `search-bar` classes, consistent table styling |
| `src/views/user/UserDetailDrawer.vue` | Better padding, consistent spacing |
| `src/views/system/SystemView.vue` | Use `page-header`, `search-bar`, `page-card` |
| `src/views/system/RoleView.vue` | Use shared classes, consistent table/dialog styling |
| `src/views/system/RoleAssignDrawer.vue` | Better tab content spacing, transfer widget styling |
| `src/views/system/MenuView.vue` | Use shared classes |
| `src/views/system/PermissionView.vue` | Use shared classes |
| `src/views/system/WebhookView.vue` | Use shared classes |
| `src/views/system/MemberView.vue` | Use shared classes |
| `src/views/audit/AuditLogView.vue` | Use shared classes |
| `src/views/profile/ChangePasswordView.vue` | Polish form card |
| `src/views/error/ForbiddenView.vue` | Use `page-card` centered layout |
| `src/views/error/NotFoundView.vue` | Use `page-card` centered layout |

### Unused files to delete:
- `src/views/system/RoleTab.vue` — unused by router
- `src/views/system/MenuTab.vue` — unused by router
- `src/views/system/PermissionTab.vue` — unused by router
- `src/views/system/WebhookTab.vue` — unused by router
- `src/views/system/MemberTab.vue` — unused by router

---

## 5. Visual Changes Per Component

### Layout (DefaultLayout.vue)
- Sidebar: keep `#001529` background, use `--shadow-md` for right border effect instead of hard line
- Header: use `--header-bg`, `--shadow-sm`, `--header-height` variables
- Logo: use `--font-size-xl` font weight 700
- Menu dividers: use `--color-border` variable
- Main content: use `--color-bg-page` background

### Dashboard
- Stat cards: use `stat-card` class with hover effect, icon in colored circle, value in `--font-size-2xl` bold
- Quick-action cards: use `page-card` class, icon + text layout, hover color transition
- Welcome section: larger username, subtle badge styling

### Login
- Form card: `--radius-lg` border radius, `--shadow-lg` shadow, `--space-xl` padding
- Input fields: wider (320px), `--space-md` gap between fields
- Button: full width, `--radius-md` radius
- Keep the purple gradient background

### Table Pages (User, Role, Menu, Permission, Member, Webhook, Audit)
- Page header: use `.page-header` with `.header-actions`
- Search bar: use `.search-bar` with consistent input width
- Table: `--el-table-header-bg-color: #fafafa`, lighter borders, consistent column widths
- Dialogs: 520px width, `--radius-lg` radius, `--space-lg` form padding
- Pagination: consistent bottom margin `--space-md`

### Drawers (RoleAssign, UserDetail)
- Use `--el-drawer-padding-primary` for consistent padding
- Tab content: `--space-lg` padding
- Tables in drawers: same styling as main tables
- Transfer widget: consistent with EP theme

### Error Pages
- Centered `page-card` with max-width 480px
- Large error code in `--color-text-secondary`
- Primary button to return home

---

## 6. Verification

1. `cd frontend && npx vue-tsc --noEmit` — type check passes
2. `cd frontend && npm run build` — build succeeds
3. Visual check each page:
   - Login: polished card on gradient
   - Dashboard: stat cards with hover, quick actions styled
   - User management: consistent header/search/table
   - System management: consistent with user management
   - Role/Menu/Permission/Member/Webhook: consistent styling
   - Audit log: consistent
   - Drawers: proper spacing
   - Error pages: centered cards
4. No functional changes — all existing behavior preserved
5. Sidebar collapse/expand still works
6. Dark sidebar text remains readable
