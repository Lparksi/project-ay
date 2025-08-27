# 开发环境设置文档

## 环境验证结果

### ✅ Git 分支管理
- 当前分支: `feature/merchant-geo-migration`
- 基于分支: `copilot/fix-eb1c132d-066b-4486-8a95-179ba0f3a178`
- 状态: 已创建迁移分支

### ✅ 代码备份
- 备份位置: `backup/merchant-components/`
- 备份内容:
  - `merchant/` - 现有商户组件目录
  - `dependency-analysis.md` - 依赖关系分析文档
  - `dev-environment-setup.md` - 本文档

### ✅ 数据库环境
- 数据库类型: PostgreSQL with PostGIS
- PostGIS 版本: 3.6 (USE_GEOS=1 USE_PROJ=1 USE_STATS=1)
- 容器状态: 运行中且健康
- 连接信息:
  - 主机: localhost (通过 Docker)
  - 端口: 5432 (内部)
  - 用户: vikunja
  - 数据库: vikunja
  - 密码: parksi2020

### ✅ Docker 环境
- Docker Compose 文件: `docker-compose.dev.yaml`
- 服务状态:
  - `ttt-db-1`: PostGIS 数据库 (健康)
  - `ttt-vikunja-1`: Vikunja 应用 (运行中)
- 端口映射: 3456:3456

## 开发环境配置

### 数据库配置
```yaml
# docker-compose.dev.yaml 中的数据库配置
db:
  image: postgis/postgis:17-master
  environment:
    POSTGRES_PASSWORD: parksi2020
    POSTGRES_USER: vikunja
    POSTGRES_DB: vikunja
  volumes:
    - ./db:/var/lib/postgresql/data
  restart: unless-stopped
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -h localhost -U $POSTGRES_USER"]
    interval: 2s
    start_period: 30s
```

### 应用配置
```yaml
# docker-compose.dev.yaml 中的应用配置
vikunja:
  build: .
  environment:
    VIKUNJA_SERVICE_PUBLICURL: http://127.0.0.1:3456
    VIKUNJA_DATABASE_HOST: db
    VIKUNJA_DATABASE_PASSWORD: parksi2020
    VIKUNJA_DATABASE_TYPE: postgres
    VIKUNJA_DATABASE_USER: vikunja
    VIKUNJA_DATABASE_DATABASE: vikunja
    VIKUNJA_SERVICE_JWTSECRET: parksi2020
  ports:
    - 3456:3456
  volumes:
    - ./files:/app/vikunja/files
    - ./frontend/src:/app/vikunja/frontend/src
    - ./frontend/dist:/app/vikunja/frontend/dist
```

## 开发工具和依赖

### 前端技术栈
- Vue.js 3 (Composition API)
- TypeScript
- Vue Router
- Vue I18n
- TDesign Vue Next (当前 UI 库)
- Vite (构建工具)

### 后端技术栈
- Go (Golang)
- GORM (ORM)
- PostgreSQL + PostGIS
- Gin (HTTP 框架)

### 地理信息相关
- PostGIS 3.6 (空间数据库扩展)
- 支持 GEOS (几何操作)
- 支持 PROJ (坐标系转换)
- 支持统计功能

## 迁移准备工作完成状态

### ✅ 已完成
1. **创建迁移分支**: `feature/merchant-geo-migration`
2. **备份现有代码**: 完整备份到 `backup/merchant-components/`
3. **依赖关系分析**: 详细分析现有组件依赖
4. **环境验证**: 确认 PostGIS 数据库正常运行
5. **开发环境文档**: 记录环境配置和设置

### 📋 下一步工作
1. **后端数据模型**: 创建 Merchant、MerchantTag、GeoPoint 模型
2. **数据库迁移**: 编写迁移脚本
3. **地理编码服务**: 集成第三方地理编码 API
4. **前端组件**: 开发新的 Vue 组件
5. **地图集成**: 集成地图 SDK (高德地图)

## 开发命令参考

### Docker 操作
```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yaml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yaml ps

# 查看日志
docker-compose -f docker-compose.dev.yaml logs -f vikunja

# 进入数据库
docker exec -it ttt-db-1 psql -U vikunja -d vikunja

# 停止服务
docker-compose -f docker-compose.dev.yaml down
```

### Git 操作
```bash
# 查看当前分支
git branch

# 切换到迁移分支
git checkout feature/merchant-geo-migration

# 提交更改
git add .
git commit -m "feat: merchant geo migration preparation"

# 推送分支
git push origin feature/merchant-geo-migration
```

### 前端开发
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 类型检查
npm run type-check
```

## 注意事项

### 数据安全
- 现有商户数据已备份
- 迁移过程中保持数据完整性
- 建议在迁移前创建数据库快照

### 兼容性
- 确保新组件与 Vikunja 主题一致
- 保持现有 API 接口兼容性
- 国际化文本需要更新

### 性能考虑
- 地图组件可能影响页面加载性能
- 大量商户数据需要分页处理
- 地理编码服务需要缓存机制

## 环境验证清单

- [x] Git 分支创建
- [x] 代码备份完成
- [x] PostGIS 数据库运行
- [x] Docker 环境正常
- [x] 依赖关系分析
- [x] 开发环境文档
- [ ] 后端模型设计
- [ ] 前端组件规划
- [ ] API 接口设计
- [ ] 测试计划制定