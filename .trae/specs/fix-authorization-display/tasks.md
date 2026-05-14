# Tasks

- [x] Task 1: 修复 RoleAssignDrawer 菜单树勾选状态回显
  - [x] SubTask 1.1: 移除 `default-checked-keys` 属性，改用 `watch` + `nextTick` + `setCheckedKeys` 方式设置勾选
  - [x] SubTask 1.2: 添加获取叶子节点 ID 的辅助函数，确保 `setCheckedKeys` 只传入叶子节点 ID
  - [x] SubTask 1.3: 确保菜单树数据和已分配菜单 ID 都加载完成后再调用 `setCheckedKeys`

- [x] Task 2: 修复 MemberDetailDrawer 角色操作后数据刷新
  - [x] SubTask 2.1: 在 `handleAssignRoles` 成功后添加 `fetchMemberMenus()` 和 `fetchMemberPermissions()` 调用
  - [x] SubTask 2.2: 在 `handleRemoveRole` 成功后添加 `fetchMemberMenus()` 和 `fetchMemberPermissions()` 调用

- [x] Task 3: 修复 MemberDetailDrawer 菜单树只读展示
  - [x] SubTask 3.1: 移除菜单树 `el-tree` 的 `show-checkbox` 属性，改为纯展示模式

# Task Dependencies
- Task 1, Task 2, Task 3 互相独立，可并行执行
