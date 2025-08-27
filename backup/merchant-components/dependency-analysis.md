# 现有商户组件依赖关系分析

## 组件概览

### 1. 前端组件
- `frontend/src/views/merchant/ListMerchants.vue` - 商户列表页面
- `frontend/src/views/merchant/MerchantEdit.vue` - 商户编辑/创建页面

### 2. 服务层
- `frontend/src/services/merchant.ts` - 商户服务类
- `frontend/src/models/merchant.ts` - 商户数据模型

### 3. 路由配置
- `/merchants` -> `merchants.index` -> `ListMerchants.vue`
- `/merchants/new` -> `merchants.create` -> `MerchantEdit.vue`
- `/merchants/:id/edit` -> `merchants.edit` -> `MerchantEdit.vue`

### 4. 导航集成
- `frontend/src/components/home/Navigation.vue` 中包含商户菜单项

## 依赖关系分析

### ListMerchants.vue 依赖项
1. **UI 组件库**: TDesign Vue Next
   - t-space, t-card, t-table, t-button, t-input, t-dialog, t-upload 等
2. **路由**: Vue Router
   - useRouter, useRoute
3. **国际化**: Vue I18n
   - useI18n, $t
4. **服务**: 
   - MerchantService (CRUD 操作)
   - MerchantModel (数据模型)
5. **消息通知**: 
   - success, error 函数
6. **图标**: Icon 组件

### MerchantEdit.vue 依赖项
1. **UI 组件**: 
   - Card, XButton (自定义组件)
   - 原生 HTML 表单元素
2. **路由**: Vue Router
3. **国际化**: Vue I18n
4. **服务**:
   - MerchantService
   - LabelService (标签服务)
   - MerchantModel, LabelModel
5. **消息通知**: success, error 函数

### MerchantService 依赖项
1. **基础服务**: AbstractService
2. **HTTP 客户端**: axios
3. **认证**: getToken 辅助函数
4. **数据模型**: MerchantModel, IMerchant 接口

### MerchantModel 依赖项
1. **基础模型**: AbstractModel
2. **用户模型**: UserModel, IUser 接口
3. **接口定义**: IMerchant

## 当前数据模型字段

### Merchant 模型字段
- id: number
- title: string (商户名称)
- legalRepresentative: string (法人代表)
- businessAddress: string (营业地址)
- businessDistrict: string (营业区域)
- validTime: string (有效时间)
- trafficConditions: string (交通状况)
- fixedEvents: string (固定事件)
- terminalType: string (终端类型)
- specialTimePeriods: string (特殊时间段)
- customFilters: string (自定义过滤器，JSON 格式)
- owner: IUser (所有者)
- created: Date (创建时间)
- updated: Date (更新时间)

## 功能特性分析

### ListMerchants.vue 功能
1. **数据展示**: 表格形式展示商户列表
2. **搜索过滤**: 实时搜索和多条件过滤
3. **分页**: 支持分页和每页数量调整
4. **批量操作**: 多选和批量删除
5. **导入功能**: Excel 文件批量导入
6. **导航**: 创建、编辑、删除操作
7. **任务集成**: 选择商户用于任务创建

### MerchantEdit.vue 功能
1. **表单编辑**: 完整的商户信息编辑表单
2. **标签映射**: 自定义标签替换功能
3. **数据验证**: 基础的表单验证
4. **创建/更新**: 支持新建和编辑模式

## 需要迁移的关键点

### 1. 数据模型扩展
- 需要添加地理位置相关字段 (lng, lat, geocode_*)
- 需要支持标签关联 (多对多关系)
- 需要地理点历史记录

### 2. UI 组件升级
- 从 TDesign 迁移到 Vikunja 原生组件
- 添加地图组件集成
- 改进标签选择器

### 3. 服务层扩展
- 添加地理编码服务
- 添加标签管理服务
- 扩展商户服务支持地理功能

### 4. 新增功能
- 地图展示和交互
- 地理编码和反向地理编码
- 标签管理系统
- 批量地理编码处理

## 风险评估

### 高风险项
1. **数据库迁移**: 需要保持现有数据完整性
2. **UI 组件兼容性**: TDesign 到 Vikunja 组件的迁移
3. **路由冲突**: 确保新旧路由不冲突

### 中风险项
1. **服务接口变更**: API 接口可能需要调整
2. **国际化文本**: 需要更新翻译文件
3. **权限系统**: 确保权限控制正确集成

### 低风险项
1. **样式调整**: CSS 样式适配
2. **图标更新**: 图标组件适配

## 迁移策略建议

### 阶段 1: 准备和备份
- ✅ 创建迁移分支
- ✅ 备份现有组件
- ✅ 分析依赖关系

### 阶段 2: 后端开发
- 创建新的数据模型
- 实现地理编码服务
- 扩展 API 端点

### 阶段 3: 前端开发
- 创建新的 Vue 组件
- 集成地图功能
- 实现标签管理

### 阶段 4: 集成和测试
- 路由切换
- 功能测试
- 性能优化

### 阶段 5: 清理
- 移除旧组件
- 清理未使用代码
- 文档更新