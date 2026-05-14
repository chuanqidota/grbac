# 修复授权展示数据 Bug Spec

## Why
角色管理的授权抽屉中，已授权的菜单和接口没有正确回显勾选状态；人员管理的授权抽屉中，分配/移除角色后，对应的菜单和接口数据没有刷新，且菜单树勾选状态展示不正确。

## What Changes
- 修复 RoleAssignDrawer 中菜单树 `default-checked-keys` 时序问题：改用 `setCheckedKeys` 方法在数据加载完成后设置勾选状态
- 修复 RoleAssignDrawer 中菜单树只应传入叶子节点 ID 的问题
- 修复 MemberDetailDrawer 中分配角色后未刷新菜单和接口数据的问题
- 修复 MemberDetailDrawer 中移除角色后未刷新菜单和接口数据的问题
- 修复 MemberDetailDrawer 中菜单树使用 `show-checkbox` 但未设置勾选状态的问题（只读展示改为不显示 checkbox）

## Impact
- Affected code: `frontend/src/views/system/RoleAssignDrawer.vue`, `frontend/src/views/system/MemberDetailDrawer.vue`

## ADDED Requirements

### Requirement: 角色授权抽屉正确回显已授权资源
系统应当在角色授权抽屉打开时，正确展示已授权的菜单勾选状态和接口权限选中状态。

#### Scenario: 打开角色授权抽屉查看已分配菜单
- **WHEN** 用户点击角色管理的"授权"按钮
- **THEN** 菜单树中已分配的菜单应正确显示为勾选状态，且仅叶子节点被勾选（父节点根据子节点自动显示半选或全选）

#### Scenario: 打开角色授权抽屉查看已分配接口权限
- **WHEN** 用户点击角色管理的"授权"按钮
- **THEN** 穿梭框右侧应正确展示已分配的接口权限

### Requirement: 人员授权抽屉角色变更后刷新关联数据
系统应当在人员授权抽屉中分配或移除角色后，自动刷新该人员的菜单和接口权限数据。

#### Scenario: 分配角色后菜单和接口数据刷新
- **WHEN** 用户在人员授权抽屉中为成员分配新角色
- **THEN** 该成员的可访问菜单和接口权限应立即刷新展示

#### Scenario: 移除角色后菜单和接口数据刷新
- **WHEN** 用户在人员授权抽屉中移除成员的角色
- **THEN** 该成员的可访问菜单和接口权限应立即刷新展示

### Requirement: 人员授权抽屉菜单树只读展示
系统应当在人员授权抽屉的菜单标签页中，以只读方式展示成员可访问的菜单树，不显示复选框。

#### Scenario: 查看成员可访问菜单
- **WHEN** 用户打开人员授权抽屉的"分配菜单"标签页
- **THEN** 应以树形结构展示成员可访问的菜单，不显示复选框（因为这是只读展示）

## MODIFIED Requirements

### Requirement: RoleAssignDrawer 菜单树勾选逻辑
菜单树不再使用 `default-checked-keys` 属性（仅在初始渲染时生效），改为在菜单树数据和已分配菜单 ID 都加载完成后，通过 `setCheckedKeys` 方法设置勾选状态，且仅传入叶子节点 ID。

### Requirement: MemberDetailDrawer 角色操作后数据刷新
分配角色和移除角色的操作完成后，除了刷新角色列表，还需同时刷新菜单和接口权限数据。
