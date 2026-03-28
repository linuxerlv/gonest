# 项目管理系统设计文档

> 创建日期：2026-03-28
> 目标：验证 gonest 框架在生产环境下的完整能力和稳定性

---

## 1. 项目概述

### 1.1 目标

构建一个完整企业级项目管理系统，验证 gonest 框架以下能力：
- **路由与中间件**：Controller 模式、Guard、Interceptor、Pipe
- **认证授权**：JWT + Casbin RBAC
- **实时通信**：WebSocket/SSE
- **依赖注入**：模块化 DI 容器
- **配置管理**：Koanf 多源配置
- **日志系统**：Zap 结构化日志
- **任务调度**：Cron + TaskQueue（可选）
- **ORM 集成**：GORM + SQLite

### 1.2 功能范围

| 模块 | 核心功能 |
|------|----------|
| User | 注册、登录、JWT认证、角色管理、个人信息 |
| Team | 创建团队、邀请成员、成员管理、团队项目 |
| Project | 创建项目、项目配置、状态管理、进度追踪 |
| Task | 任务创建、状态流转、子任务、评论、看板视图 |
| Document | 文档创建、Markdown编辑、版本管理 |
| Activity | 操作日志、活动记录、审计追踪 |
| Notification | 实时通知、WebSocket推送、已读管理 |

---

## 2. 技术架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (Vue3 Monorepo)                  │
│  packages: core (API/WS) + ui + features (project/task...)  │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ REST API + WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Backend (gonest + GORM)                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Entry Point (cmd/server/main.go)                    │  │
│  │  - Builder模式创建WebApplication                      │  │
│  │  - DI注册所有模块服务                                 │  │
│  │  - 注册模块路由                                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                              │                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Modules (DDD 模块化)                                │  │
│  │  ├── user/       ├── project/   ├── task/            │  │
│  │  ├── team/       ├── document/  ├── activity/        │  │
│  │  └── notification/                                   │  │
│  │  每个模块: domain/handler/service/repository         │  │
│  └──────────────────────────────────────────────────────┘  │
│                              │                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Shared Infrastructure                               │  │
│  │  ├── database (GORM + SQLite)                        │  │
│  │  ├── auth (JWT + Casbin RBAC)                        │  │
│  │  ├── websocket (Hub + Broadcast)                     │  │
│  │  ├── config (Koanf YAML)                             │  │
│  │  ├── logger (Zap)                                    │  │
│  │  └── middleware (Recovery/CORS/RateLimit)            │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   SQLite DB     │
                    │  (data.db)      │
                    └─────────────────┘
```

### 2.2 项目目录结构

```
project-management/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # 应用入口
│   │
│   ├── modules/
│   │   ├── user/
│   │   │   ├── domain/
│   │   │   │   ├── entity.go        # User model
│   │   │   │   ├── repository.go    # 接口定义
│   │   │   │   └── dto.go           # DTO结构
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go  # 认证处理器
│   │   │   │   ├── user_handler.go  # 用户处理器
│   │   │   │   └── routes.go        # 路由定义
│   │   │   ├── service/
│   │   │   │   ├── user_service.go  # 业务逻辑
│   │   │   ├── repository/
│   │   │   │   ├── user_repo.go     # GORM实现
│   │   │   └── module.go            # 模块注册
│   │   │
│   │   ├── team/                    # 同上结构
│   │   ├── project/
│   │   ├── task/
│   │   ├── document/
│   │   ├── activity/
│   │   └── notification/
│   │
│   ├── shared/
│   │   ├── database/
│   │   │   ├── db.go                # GORM初始化
│   │   │   └── migrate.go           # 自动迁移
│   │   ├── auth/
│   │   │   ├── jwt.go               # JWT Provider
│   │   │   ├── casbin.go            # Casbin Enforcer
│   │   │   └── middleware.go        # Auth中间件
│   │   ├── websocket/
│   │   │   ├── hub.go               # WebSocket Hub
│   │   │   ├── client.go            # Client管理
│   │   │   └── message.go           # 消息定义
│   │   ├── config/
│   │   │   ├── config.go            # 配置加载
│   │   │   └── types.go             # 配置结构体
│   │   ├── logger/
│   │   │   └── logger.go            # Zap初始化
│   │   ├── middleware/
│   │   │   ├── cors.go              # CORS配置
│   │   │   ├── recovery.go          # Panic恢复
│   │   │   ├── ratelimit.go         # 限流
│   │   │   └── requestid.go         # 请求ID
│   │   └── response/
│   │   │   ├── response.go          # 统一响应格式
│   │   │   └── errors.go            # 错误定义
│   │
│   ├── config.yaml                  # 配置文件
│   ├── model.conf                   # Casbin模型
│   ├── policy.csv                   # Casbin策略
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── packages/
│   │   ├── core/
│   │   │   ├── src/
│   │   │   │   ├── api/
│   │   │   │   │   ├── client.ts    # Axios封装
│   │   │   │   │   ├── auth.ts      # 认证API
│   │   │   │   │   ├── project.ts
│   │   │   │   │   ├── task.ts
│   │   │   │   │   └── team.ts
│   │   │   │   ├── types/
│   │   │   │   │   ├── user.ts
│   │   │   │   │   ├── project.ts
│   │   │   │   │   ├── task.ts
│   │   │   │   │   └── common.ts
│   │   │   │   ├── websocket/
│   │   │   │   │   ├── client.ts    # WS客户端
│   │   │   │   │   └── types.ts
│   │   │   │   ├── auth/
│   │   │   │   │   ├── token.ts     # Token管理
│   │   │   │   │   └── guard.ts     # 路由守卫
│   │   │   │   ├── utils/
│   │   │   │   │   ├── date.ts
│   │   │   │   │   ├── validation.ts
│   │   │   │   │   └── format.ts
│   │   │   └── package.json
│   │   │
│   │   ├── ui/
│   │   │   ├── src/
│   │   │   │   ├── components/
│   │   │   │   │   ├── Board/
│   │   │   │   │   │   ├── Board.vue
│   │   │   │   │   │   ├── Column.vue
│   │   │   │   │   ├── TaskCard/
│   │   │   │   │   │   ├── TaskCard.vue
│   │   │   │   │   │   ├── TaskDetail.vue
│   │   │   │   │   ├── ProjectCard/
│   │   │   │   │   │   ├── ProjectCard.vue
│   │   │   │   │   ├── Navbar/
│   │   │   │   │   │   ├── Navbar.vue
│   │   │   │   │   ├── Sidebar/
│   │   │   │   │   │   ├── Sidebar.vue
│   │   │   │   │   ├── Notification/
│   │   │   │   │   │   ├── NotificationPanel.vue
│   │   │   │   │   ├── User/
│   │   │   │   │   │   ├── Avatar.vue
│   │   │   │   │   │   ├── UserSelect.vue
│   │   │   │   │   ├── Form/
│   │   │   │   │   │   ├── FormBuilder.vue
│   │   │   │   │   ├── Comment/
│   │   │   │   │   │   ├── CommentList.vue
│   │   │   │   │   ├── Activity/
│   │   │   │   │   │   ├── ActivityLog.vue
│   │   │   │   │   ├── layouts/
│   │   │   │   │   │   ├── MainLayout.vue
│   │   │   │   │   │   ├── AuthLayout.vue
│   │   │   │   │   ├── styles/
│   │   │   │   │   │   ├── theme.scss
│   │   │   │   │   │   ├── variables.scss
│   │   │   │   │   │   ├── mixins.scss
│   │   │   │   └── package.json
│   │   │
│   │   ├── features/
│   │   │   ├── project/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── ProjectList.vue
│   │   │   │   │   │   ├── ProjectDetail.vue
│   │   │   │   │   │   ├── ProjectCreate.vue
│   │   │   │   │   │   ├── ProjectSettings.vue
│   │   │   │   │   ├── store/
│   │   │   │   │   │   ├── projectStore.ts
│   │   │   │   │   ├── hooks/
│   │   │   │   │   │   ├── useProject.ts
│   │   │   │   │   │   ├── useProjectTasks.ts
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   ├── task/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── TaskBoard.vue
│   │   │   │   │   │   ├── TaskList.vue
│   │   │   │   │   │   ├── TaskDetail.vue
│   │   │   │   │   ├── store/
│   │   │   │   │   │   ├── taskStore.ts
│   │   │   │   │   ├── components/
│   │   │   │   │   │   ├── TaskForm.vue
│   │   │   │   │   │   ├── TaskComment.vue
│   │   │   │   │   │   ├── TaskStatusBadge.vue
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   ├── team/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── TeamList.vue
│   │   │   │   │   │   ├── TeamDetail.vue
│   │   │   │   │   │   ├── TeamMembers.vue
│   │   │   │   │   ├── store/
│   │   │   │   │   │   ├── teamStore.ts
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   ├── user/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── Login.vue
│   │   │   │   │   │   ├── Register.vue
│   │   │   │   │   │   ├── Profile.vue
│   │   │   │   │   ├── store/
│   │   │   │   │   │   ├── userStore.ts
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   ├── notification/
│   │   │   │   ├── src/
│   │   │   │   │   ├── store/
│   │   │   │   │   │   ├── notificationStore.ts
│   │   │   │   │   ├── hooks/
│   │   │   │   │   │   ├── useWebSocket.ts
│   │   │   │   │   │   ├── useNotification.ts
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   └── dashboard/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── Dashboard.vue
│   │   │   │   │   │   ├── Overview.vue
│   │   │   │   │   ├── components/
│   │   │   │   │   │   ├── StatsCard.vue
│   │   │   │   │   │   ├── RecentActivity.vue
│   │   │   │   │   ├── package.json
│   │   │   │
│   │   │   └── document/
│   │   │   │   ├── src/
│   │   │   │   │   ├── views/
│   │   │   │   │   │   ├── DocumentList.vue
│   │   │   │   │   │   ├── DocumentEditor.vue
│   │   │   │   │   ├── components/
│   │   │   │   │   │   ├── MarkdownEditor.vue
│   │   │   │   │   ├── package.json
│   │   │
│   │   └── app/
│   │   │   ├── src/
│   │   │   │   ├── main.ts
│   │   │   │   ├── App.vue
│   │   │   │   ├── router/
│   │   │   │   │   ├── index.ts
│   │   │   │   │   ├── routes.ts
│   │   │   │   │   ├── guards.ts
│   │   │   │   ├── assets/
│   │   │   │   │   ├── global.scss
│   │   │   │   ├── vite.config.ts
│   │   │   │   ├── tsconfig.json
│   │   │   │   ├── package.json
│   │   │
│   ├── pnpm-workspace.yaml
│   ├── package.json
│   ├── tsconfig.base.json
│   └── .gitignore
│
└── docs/
    └── specs/
        └── 2026-03-28-project-management-system-design.md
```

---

## 3. 领域模型设计

### 3.1 实体关系图

```
User ─────┬───────────────────────────────────────
          │ owns                                  │ member of
          ▼                                       ▼
        Team ────────┐                      TeamUser (M2M)
          │          │
          │ has      │
          ▼          │
       Project ──────┤
          │          │
          │ contains │
          ▼          │
        Task ────────┤
          │          │
          │ has      │
          ▼          │
      Document       │
          │          │
          │ logs     │
          ▼          │
      Activity ──────┤
          │          │
          │ creates  │
          ▼          │
   Notification ─────┘
```

### 3.2 实体定义

#### User 模块

```go
// modules/user/domain/entity.go
type User struct {
    gorm.Model
    Username    string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email       string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password    string    `gorm:"size:255;not null" json:"-"` // bcrypt hash
    Avatar      string    `gorm:"size:255" json:"avatar"`
    Role        string    `gorm:"size:20;default:'member'" json:"role"` // admin/project_admin/member
    Status      string    `gorm:"size:20;default:'active'" json:"status"` // active/inactive/banned
    Teams       []Team    `gorm:"many2many:team_users;" json:"teams,omitempty"`
    LastLoginAt *time.Time `json:"lastLoginAt"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// modules/user/domain/dto.go
type RegisterDTO struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type UserResponseDTO struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Avatar    string    `json:"avatar"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"createdAt"`
}

type UpdateProfileDTO struct {
    Username string `json:"username" validate:"omitempty,min=3,max=50"`
    Avatar   string `json:"avatar" validate:"omitempty,url"`
}
```

#### Team 模块

```go
// modules/team/domain/entity.go
type Team struct {
    gorm.Model
    Name        string    `gorm:"size:100;not null" json:"name"`
    Description string    `gorm:"size:500" json:"description"`
    Logo        string    `gorm:"size:255" json:"logo"`
    OwnerID     uint      `gorm:"not null" json:"ownerId"`
    Owner       User      `gorm:"foreignKey:OwnerID" json:"owner"`
    Members     []User    `gorm:"many2many:team_users;" json:"members"`
    Projects    []Project `gorm:"foreignKey:TeamID" json:"projects,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

type TeamUser struct {
    TeamID    uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"primaryKey"`
    Role      string    `gorm:"size:20;default:'member'" json:"role"` // owner/admin/member
    JoinedAt  time.Time `gorm:"autoCreateTime" json:"joinedAt"`
}

// modules/team/domain/dto.go
type CreateTeamDTO struct {
    Name        string `json:"name" validate:"required,min=2,max=100"`
    Description string `json:"description" validate:"omitempty,max=500"`
    Logo        string `json:"logo" validate:"omitempty,url"`
}

type AddMemberDTO struct {
    UserID uint   `json:"userId" validate:"required"`
    Role   string `json:"role" validate:"required,oneof=admin member"`
}

type TeamResponseDTO struct {
    ID          uint              `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Logo        string            `json:"logo"`
    Owner       UserResponseDTO   `json:"owner"`
    Members     []UserResponseDTO `json:"members"`
    MemberCount int               `json:"memberCount"`
    ProjectCount int              `json:"projectCount"`
    CreatedAt   time.Time         `json:"createdAt"`
}
```

#### Project 模块

```go
// modules/project/domain/entity.go
type Project struct {
    gorm.Model
    Name          string     `gorm:"size:100;not null" json:"name"`
    Key           string     `gorm:"uniqueIndex;size:10;not null" json:"key"` // 项目标识（如 PM-001）
    Description   string     `gorm:"type:text" json:"description"`
    Status        string     `gorm:"size:20;default:'active'" json:"status"` // active/archived/completed/on_hold
    Priority      string     `gorm:"size:20;default:'medium'" json:"priority"` // high/medium/low/critical
    Visibility    string     `gorm:"size:20;default:'private'" json:"visibility"` // public/private
    TeamID        uint       `gorm:"not null;index" json:"teamId"`
    Team          Team       `gorm:"foreignKey:TeamID" json:"team"`
    OwnerID       uint       `gorm:"not null;index" json:"ownerId"`
    Owner         User       `gorm:"foreignKey:OwnerID" json:"owner"`
    StartDate     *time.Time `json:"startDate"`
    EndDate       *time.Time `json:"endDate"`
    Progress      int        `gorm:"default:0" json:"progress"` // 0-100
    TaskCount     int        `gorm:"default:0" json:"taskCount"`
    CompletedCount int       `gorm:"default:0" json:"completedCount"`
    CreatedAt     time.Time  `json:"createdAt"`
    UpdatedAt     time.Time  `json:"updatedAt"`
}

// modules/project/domain/dto.go
type CreateProjectDTO struct {
    Name        string     `json:"name" validate:"required,min=2,max=100"`
    Key         string     `json:"key" validate:"required,min=2,max=10"`
    Description string     `json:"description"`
    TeamID      uint       `json:"teamId" validate:"required"`
    Priority    string     `json:"priority" validate:"omitempty,oneof=high medium low critical"`
    StartDate   *time.Time `json:"startDate"`
    EndDate     *time.Time `json:"endDate"`
    Visibility  string     `json:"visibility" validate:"omitempty,oneof=public private"`
}

type UpdateProjectDTO struct {
    Name        string     `json:"name" validate:"omitempty,min=2,max=100"`
    Description string     `json:"description"`
    Status      string     `json:"status" validate:"omitempty,oneof=active archived completed on_hold"`
    Priority    string     `json:"priority" validate:"omitempty,oneof=high medium low critical"`
    Progress    int        `json:"progress" validate:"omitempty,min=0,max=100"`
    EndDate     *time.Time `json:"endDate"`
}

type ProjectResponseDTO struct {
    ID            uint            `json:"id"`
    Name          string          `json:"name"`
    Key           string          `json:"key"`
    Description   string          `json:"description"`
    Status        string          `json:"status"`
    Priority      string          `json:"priority"`
    Visibility    string          `json:"visibility"`
    Team          TeamResponseDTO `json:"team"`
    Owner         UserResponseDTO `json:"owner"`
    Progress      int             `json:"progress"`
    TaskCount     int             `json:"taskCount"`
    CompletedCount int            `json:"completedCount"`
    StartDate     *time.Time      `json:"startDate"`
    EndDate       *time.Time      `json:"endDate"`
    CreatedAt     time.Time       `json:"createdAt"`
}
```

#### Task 模块

```go
// modules/task/domain/entity.go
type Task struct {
    gorm.Model
    Title         string     `gorm:"size:200;not null" json:"title"`
    Description   string     `gorm:"type:text" json:"description"`
    Status        string     `gorm:"size:20;default:'todo';index" json:"status"` // todo/in_progress/review/done/cancelled
    Priority      string     `gorm:"size:20;default:'medium';index" json:"priority"` // high/medium/low/urgent
    Type          string     `gorm:"size:20;default:'task'" json:"type"` // task/bug/feature/improvement
    ProjectID     uint       `gorm:"not null;index" json:"projectId"`
    Project       Project    `gorm:"foreignKey:ProjectID" json:"project"`
    AssigneeID    *uint      `gorm:"index" json:"assigneeId"`
    Assignee      *User      `gorm:"foreignKey:AssigneeID" json:"assignee"`
    ReporterID    uint       `gorm:"not null;index" json:"reporterId"`
    Reporter      User       `gorm:"foreignKey:ReporterID" json:"reporter"`
    ParentID      *uint      `gorm:"index" json:"parentId"` // 子任务
    Children      []Task     `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Tags          []string   `gorm:"type:json;serializer:json" json:"tags"`
    DueDate       *time.Time `gorm:"index" json:"dueDate"`
    EstimatedHours *float64  `json:"estimatedHours"`
    ActualHours   *float64   `json:"actualHours"`
    Order         int        `gorm:"default:0" json:"order"` // 看板排序
    Comments      []Comment  `gorm:"foreignKey:TaskID" json:"comments,omitempty"`
    Attachments   []Attachment `gorm:"foreignKey:TaskID" json:"attachments,omitempty"`
    CompletedAt   *time.Time `json:"completedAt"`
    CreatedAt     time.Time  `json:"createdAt"`
    UpdatedAt     time.Time  `json:"updatedAt"`
}

type Comment struct {
    gorm.Model
    TaskID    uint      `gorm:"not null;index" json:"taskId"`
    Task      Task      `gorm:"foreignKey:TaskID"`
    UserID    uint      `gorm:"not null;index" json:"userId"`
    User      User      `gorm:"foreignKey:UserID" json:"user"`
    Content   string    `gorm:"type:text;not null" json:"content"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type Attachment struct {
    gorm.Model
    TaskID     uint   `gorm:"not null;index" json:"taskId"`
    FileName   string `gorm:"size:255;not null" json:"fileName"`
    FilePath   string `gorm:"size:500;not null" json:"filePath"`
    FileSize   int64  `json:"fileSize"`
    MimeType   string `gorm:"size:100" json:"mimeType"`
    UploadedBy uint   `gorm:"not null" json:"uploadedBy"`
    CreatedAt  time.Time `json:"createdAt"`
}

// modules/task/domain/dto.go
type CreateTaskDTO struct {
    Title         string   `json:"title" validate:"required,min=2,max=200"`
    Description   string   `json:"description"`
    ProjectID     uint     `json:"projectId" validate:"required"`
    AssigneeID    *uint    `json:"assigneeId"`
    Priority      string   `json:"priority" validate:"omitempty,oneof=high medium low urgent"`
    Type          string   `json:"type" validate:"omitempty,oneof=task bug feature improvement"`
    ParentID      *uint    `json:"parentId"`
    Tags          []string `json:"tags"`
    DueDate       *time.Time `json:"dueDate"`
    EstimatedHours *float64 `json:"estimatedHours"`
}

type UpdateTaskDTO struct {
    Title         string     `json:"title" validate:"omitempty,min=2,max=200"`
    Description   string     `json:"description"`
    Status        string     `json:"status" validate:"omitempty,oneof=todo in_progress review done cancelled"`
    Priority      string     `json:"priority" validate:"omitempty,oneof=high medium low urgent"`
    AssigneeID    *uint      `json:"assigneeId"`
    Tags          []string   `json:"tags"`
    DueDate       *time.Time `json:"dueDate"`
    EstimatedHours *float64  `json:"estimatedHours"`
    ActualHours   *float64   `json:"actualHours"`
    Order         int        `json:"order"`
}

type UpdateStatusDTO struct {
    Status string `json:"status" validate:"required,oneof=todo in_progress review done cancelled"`
}

type CreateCommentDTO struct {
    Content string `json:"content" validate:"required,min=1,max=5000"`
}

type TaskResponseDTO struct {
    ID            uint             `json:"id"`
    Title         string           `json:"title"`
    Description   string           `json:"description"`
    Status        string           `json:"status"`
    Priority      string           `json:"priority"`
    Type          string           `json:"type"`
    Project       ProjectResponseDTO `json:"project"`
    Assignee      *UserResponseDTO `json:"assignee"`
    Reporter      UserResponseDTO  `json:"reporter"`
    ParentID      *uint            `json:"parentId"`
    Children      []TaskResponseDTO `json:"children,omitempty"`
    Tags          []string         `json:"tags"`
    DueDate       *time.Time       `json:"dueDate"`
    EstimatedHours *float64        `json:"estimatedHours"`
    ActualHours   *float64         `json:"actualHours"`
    Order         int              `json:"order"`
    Comments      []CommentDTO     `json:"comments,omitempty"`
    CompletedAt   *time.Time       `json:"completedAt"`
    CreatedAt     time.Time        `json:"createdAt"`
    UpdatedAt     time.Time        `json:"updatedAt"`
}

type CommentDTO struct {
    ID        uint            `json:"id"`
    Content   string          `json:"content"`
    User      UserResponseDTO `json:"user"`
    CreatedAt time.Time       `json:"createdAt"`
}
```

#### Document 模块

```go
// modules/document/domain/entity.go
type Document struct {
    gorm.Model
    Title       string    `gorm:"size:200;not null" json:"title"`
    Content     string    `gorm:"type:longtext" json:"content"`
    Type        string    `gorm:"size:20;default:'markdown'" json:"type"` // markdown/text/wiki/html
    ProjectID   uint      `gorm:"not null;index" json:"projectId"`
    Project     Project   `gorm:"foreignKey:ProjectID" json:"project"`
    AuthorID    uint      `gorm:"not null;index" json:"authorId"`
    Author      User      `gorm:"foreignKey:AuthorID" json:"author"`
    Version     int       `gorm:"default:1" json:"version"`
    IsPublic    bool      `gorm:"default:false" json:"isPublic"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

type DocumentVersion struct {
    gorm.Model
    DocumentID uint      `gorm:"not null;index" json:"documentId"`
    Content    string    `gorm:"type:longtext" json:"content"`
    Version    int       `gorm:"not null" json:"version"`
    EditorID   uint      `gorm:"not null" json:"editorId"`
    Editor     User      `gorm:"foreignKey:EditorID" json:"editor"`
    EditedAt   time.Time `gorm:"autoCreateTime" json:"editedAt"`
}
```

#### Activity 模块

```go
// modules/activity/domain/entity.go
type Activity struct {
    gorm.Model
    Type        string    `gorm:"size:50;not null;index" json:"type"` // created/updated/deleted/commented/assigned/status_changed/archived
    EntityType  string    `gorm:"size:50;not null;index" json:"entityType"` // project/task/document/comment
    EntityID    uint      `gorm:"not null;index" json:"entityId"`
    UserID      uint      `gorm:"not null;index" json:"userId"`
    User        User      `gorm:"foreignKey:UserID" json:"user"`
    ProjectID   *uint     `gorm:"index" json:"projectId"`
    TaskID      *uint     `gorm:"index" json:"taskId"`
    Content     string    `gorm:"type:text" json:"content"` // JSON详情
    Metadata    string    `gorm:"type:text" json:"metadata"` // 额外元数据
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type ActivityDTO struct {
    ID          uint            `json:"id"`
    Type        string          `json:"type"`
    EntityType  string          `json:"entityType"`
    EntityID    uint            `json:"entityId"`
    User        UserResponseDTO `json:"user"`
    Content     string          `json:"content"`
    CreatedAt   time.Time       `json:"createdAt"`
}
```

#### Notification 模块

```go
// modules/notification/domain/entity.go
type Notification struct {
    gorm.Model
    Type        string    `gorm:"size:50;not null;index" json:"type"` // task_assigned/comment/mention/deadline/project_invite/status_update
    Title       string    `gorm:"size:200;not null" json:"title"`
    Content     string    `gorm:"type:text" json:"content"`
    UserID      uint      `gorm:"not null;index" json:"userId"`
    User        User      `gorm:"foreignKey:UserID" json:"user"`
    IsRead      bool      `gorm:"default:false;index" json:"isRead"`
    EntityID    uint      `json:"entityId"`
    EntityType  string    `gorm:"size:50" json:"entityType"`
    ActionURL   string    `gorm:"size:255" json:"actionUrl"` // 跳转链接
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type NotificationDTO struct {
    ID         uint      `json:"id"`
    Type       string    `json:"type"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    IsRead     bool      `json:"isRead"`
    EntityID   uint      `json:"entityId"`
    EntityType string    `json:"entityType"`
    ActionURL  string    `json:"actionUrl"`
    CreatedAt  time.Time `json:"createdAt"`
}
```

---

## 4. API 设计

### 4.1 RESTful API 规范

**统一响应格式**：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... },
  "timestamp": "2026-03-28T10:30:00Z"
}

// 错误响应
{
  "code": 400,
  "message": "validation failed",
  "errors": [
    {"field": "username", "message": "username is required"}
  ],
  "timestamp": "2026-03-28T10:30:00Z"
}
```

### 4.2 API 路由表

#### Auth & User API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | public |
| POST | `/api/v1/auth/login` | 登录（返回JWT） | public |
| POST | `/api/v1/auth/logout` | 登出 | authenticated |
| POST | `/api/v1/auth/refresh` | 刷新Token | authenticated |
| GET | `/api/v1/auth/profile` | 当前用户信息 | authenticated |
| PUT | `/api/v1/auth/profile` | 更新个人信息 | authenticated |
| PUT | `/api/v1/auth/password` | 修改密码 | authenticated |
| GET | `/api/v1/users` | 用户列表 | admin |
| GET | `/api/v1/users/:id` | 用户详情 | authenticated |
| PUT | `/api/v1/users/:id/role` | 修改用户角色 | admin |

#### Team API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| POST | `/api/v1/teams` | 创建团队 | authenticated |
| GET | `/api/v1/teams` | 我的团队列表 | authenticated |
| GET | `/api/v1/teams/:id` | 团队详情 | team_member |
| PUT | `/api/v1/teams/:id` | 更新团队 | team_admin |
| DELETE | `/api/v1/teams/:id` | 删除团队 | team_owner |
| POST | `/api/v1/teams/:id/members` | 添加成员 | team_admin |
| GET | `/api/v1/teams/:id/members` | 成员列表 | team_member |
| DELETE | `/api/v1/teams/:id/members/:userId` | 移除成员 | team_admin |
| PUT | `/api/v1/teams/:id/members/:userId/role` | 修改成员角色 | team_owner |
| GET | `/api/v1/teams/:id/projects` | 团队项目列表 | team_member |

#### Project API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| POST | `/api/v1/projects` | 创建项目 | authenticated |
| GET | `/api/v1/projects` | 项目列表（可筛选） | authenticated |
| GET | `/api/v1/projects/:id` | 项目详情 | project_member |
| PUT | `/api/v1/projects/:id` | 更新项目 | project_admin |
| DELETE | `/api/v1/projects/:id` | 删除项目 | project_admin |
| GET | `/api/v1/projects/:id/tasks` | 项目任务列表 | project_member |
| GET | `/api/v1/projects/:id/tasks/board` | 看板视图数据 | project_member |
| GET | `/api/v1/projects/:id/documents` | 项目文档列表 | project_member |
| GET | `/api/v1/projects/:id/activities` | 项目活动日志 | project_member |
| GET | `/api/v1/projects/:id/stats` | 项目统计 | project_member |
| POST | `/api/v1/projects/:id/members` | 添加项目成员 | project_admin |
| GET | `/api/v1/projects/:id/members` | 项目成员列表 | project_member |

#### Task API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| POST | `/api/v1/tasks` | 创建任务 | authenticated |
| GET | `/api/v1/tasks/:id` | 任务详情 | project_member |
| PUT | `/api/v1/tasks/:id` | 更新任务 | task_assignee/reporter |
| DELETE | `/api/v1/tasks/:id` | 删除任务 | project_admin/reporter |
| PUT | `/api/v1/tasks/:id/status` | 更新任务状态 | task_assignee |
| PUT | `/api/v1/tasks/:id/assignee` | 指派任务 | project_admin |
| POST | `/api/v1/tasks/:id/comments` | 添加评论 | project_member |
| GET | `/api/v1/tasks/:id/comments` | 评论列表 | project_member |
| POST | `/api/v1/tasks/:id/attachments` | 上传附件 | project_member |
| GET | `/api/v1/tasks/:id/attachments` | 附件列表 | project_member |
| GET | `/api/v1/tasks/:id/subtasks` | 子任务列表 | project_member |
| POST | `/api/v1/tasks/:id/subtasks` | 创建子任务 | project_member |
| GET | `/api/v1/tasks/search` | 搜索任务 | authenticated |

#### Document API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| POST | `/api/v1/documents` | 创建文档 | project_member |
| GET | `/api/v1/documents/:id` | 文档详情 | project_member |
| PUT | `/api/v1/documents/:id` | 更新文档 | author |
| DELETE | `/api/v1/documents/:id` | 删除文档 | author/project_admin |
| GET | `/api/v1/documents/:id/history` | 文档历史版本 | project_member |
| POST | `/api/v1/documents/:id/restore/:version` | 恢复版本 | author |

#### Notification API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| GET | `/api/v1/notifications` | 通知列表 | authenticated |
| GET | `/api/v1/notifications/unread` | 未读通知 | authenticated |
| PUT | `/api/v1/notifications/:id/read` | 标记已读 | authenticated |
| PUT | `/api/v1/notifications/read-all` | 全部已读 | authenticated |
| DELETE | `/api/v1/notifications/:id` | 删除通知 | authenticated |
| WS | `/ws` | WebSocket连接 | authenticated |

#### Activity API

| Method | Path | 描述 | 权限 |
|--------|------|------|------|
| GET | `/api/v1/activities` | 全局活动日志 | authenticated |
| GET | `/api/v1/activities/user/:userId` | 用户活动 | authenticated |

---

## 5. 权限模型（Casbin RBAC）

### 5.1 角色定义

| 角色 | 描述 | 权限范围 |
|------|------|----------|
| `admin` | 全局管理员 | 所有操作 |
| `team_owner` | 团队所有者 | 团队管理 + 项目管理 |
| `team_admin` | 团队管理员 | 团队成员管理 + 项目管理 |
| `team_member` | 团队成员 | 查看团队 + 创建项目 |
| `project_admin` | 项目管理员 | 项目管理 + 任务管理 |
| `project_member` | 项目成员 | 查看项目 + 创建任务 + 评论 |

### 5.2 Casbin 模型配置

```conf
# model.conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _
g2 = _, _, _  # 项目角色继承
g3 = _, _, _  # 团队角色继承

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act) || 
    g2(r.sub, p.sub, r.obj) || 
    g3(r.sub, p.sub, r.obj)
```

### 5.3 策略规则

```csv
# policy.csv - 基础角色权限
p, admin, /api/v1/*, *
p, team_owner, /api/v1/teams/*, *
p, team_admin, /api/v1/teams/:id/members/*, *
p, project_admin, /api/v1/projects/*, *
p, project_admin, /api/v1/tasks/*, *
p, project_member, /api/v1/projects/:id, GET
p, project_member, /api/v1/tasks, POST
p, project_member, /api/v1/tasks/:id, GET|PUT
p, project_member, /api/v1/tasks/:id/comments/*, *

# 用户角色分配
g, user:1, admin
g, user:2, team_owner, team:1
g, user:3, project_admin, project:1
g, user:4, project_member, project:1
```

### 5.4 动态权限中间件

```go
// shared/auth/casbin.go
func (m *CasbinMiddleware) Handle(ctx abstract.ContextAbstract, next func() error) error {
    // 1. 获取用户ID
    userID := auth.GetUserID(ctx)
    if userID == "" {
        return abstract.Unauthorized("未授权")
    }
    
    // 2. 获取请求路径和方法
    path := ctx.Path()
    method := ctx.Method()
    
    // 3. 构建Casbin主体
    sub := fmt.Sprintf("user:%s", userID)
    
    // 4. 检查全局角色权限
    if ok, _ := m.enforcer.Enforce(sub, path, method); ok {
        return next()
    }
    
    // 5. 检查项目级权限（从路径提取项目ID）
    projectID := extractProjectID(path)
    if projectID != "" {
        projectSub := fmt.Sprintf("user:%s:project:%s", userID, projectID)
        if ok, _ := m.enforcer.Enforce(projectSub, path, method); ok {
            return next()
        }
    }
    
    // 6. 检查团队级权限
    teamID := extractTeamID(path)
    if teamID != "" {
        teamSub := fmt.Sprintf("user:%s:team:%s", userID, teamID)
        if ok, _ := m.enforcer.Enforce(teamSub, path, method); ok {
            return next()
        }
    }
    
    return abstract.Forbidden("无权限访问")
}
```

---

## 6. WebSocket 实时通知

### 6.1 WebSocket Hub 设计

```go
// shared/websocket/hub.go
type Hub struct {
    clients    map[uint]*Client        // userID -> Client
    channels   map[string]map[uint]bool // channel -> userIDs
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    ID        uint
    Conn      *websocket.Conn
    Send      chan *Message
    Channels  map[string]bool          // subscribed channels
    Hub       *Hub
}

type Message struct {
    Type      string `json:"type"`      // task_updated/comment/notification
    Channel   string `json:"channel"`   // project:1/team:1/user:1
    Payload   any    `json:"payload"`
    Timestamp int64  `json:"timestamp"`
}

// 消息类型
const (
    MsgTaskUpdated      = "task_updated"
    MsgTaskAssigned     = "task_assigned"
    MsgTaskComment      = "task_comment"
    MsgProjectUpdated   = "project_updated"
    MsgNotification     = "notification"
    MsgUserMention      = "user_mention"
    MsgDeadlineReminder = "deadline_reminder"
)
```

### 6.2 WebSocket 处理器

```go
// shared/websocket/handler.go
type WebSocketHandler struct {
    hub    *Hub
    jwt    *auth.JWTProvider
}

func (h *WebSocketHandler) Handle(ctx abstract.ContextAbstract) error {
    // 1. 从Query获取token
    token := ctx.Query("token")
    if token == "" {
        return abstract.Unauthorized("缺少token")
    }
    
    // 2. 验证JWT
    claims, err := h.jwt.ValidateToken(token)
    if err != nil {
        return abstract.Unauthorized("token无效")
    }
    
    // 3. 升级为WebSocket连接
    hc := ctx.(*core.HttpContext)
    conn, err := websocket.Accept(hc.Request(), hc.ResponseWriter(), nil)
    if err != nil {
        return err
    }
    
    // 4. 创建客户端并注册
    client := &Client{
        ID:       claims.UserID,
        Conn:     conn,
        Send:     make(chan *Message, 100),
        Channels: make(map[string]bool),
        Hub:      h.hub,
    }
    h.hub.Register(client)
    
    // 5. 启动读写goroutine
    go client.writePump()
    go client.readPump(h)
    
    return nil
}
```

### 6.3 通知推送示例

```go
// modules/task/service/task_service.go
func (s *TaskService) UpdateTask(id uint, dto UpdateTaskDTO, userID uint) (*Task, error) {
    task, err := s.repo.Update(id, dto)
    if err != nil {
        return nil, err
    }
    
    // 1. 推送WebSocket通知
    if task.AssigneeID != nil {
        s.hub.BroadcastToUser(*task.AssigneeID, &Message{
            Type:    MsgTaskUpdated,
            Channel: fmt.Sprintf("project:%d", task.ProjectID),
            Payload: task,
        })
    }
    
    // 2. 创建数据库通知
    s.notificationService.Create(Notification{
        Type:        "task_updated",
        Title:       fmt.Sprintf("任务更新：%s", task.Title),
        UserID:      *task.AssigneeID,
        EntityID:    task.ID,
        EntityType:  "task",
        ActionURL:   fmt.Sprintf("/projects/%d/tasks/%d", task.ProjectID, task.ID),
    })
    
    // 3. 记录活动日志
    s.activityService.Log(Activity{
        Type:        "updated",
        EntityType:  "task",
        EntityID:    task.ID,
        UserID:      userID,
        ProjectID:   &task.ProjectID,
        TaskID:      &task.ID,
        Content:     toJSON(dto),
    })
    
    return task, nil
}
```

---

## 7. 前端架构

### 7.1 技术栈

- **Vue 3** + **TypeScript**
- **Pinia** 状态管理
- **Vue Router** 路由
- **Element Plus** UI组件库
- **Axios** HTTP客户端
- **pnpm** + **workspace** Monorepo管理

### 7.2 核心功能模块

#### core 包（公共基础）

```typescript
// packages/core/src/api/client.ts
import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
})

// 请求拦截器 - 添加JWT
client.interceptors.request.use(config => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 - 统一错误处理
client.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      clearToken()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default client
```

```typescript
// packages/core/src/websocket/client.ts
export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  
  connect(token: string) {
    const wsUrl = `${WS_BASE_URL}?token=${token}`
    this.ws = new WebSocket(wsUrl)
    
    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      this.emit('connected')
    }
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data)
      this.emit(message.type, message.payload)
    }
    
    this.ws.onclose = () => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => this.connect(token), 3000)
        this.reconnectAttempts++
      }
    }
  }
  
  subscribe(channel: string) {
    this.ws?.send(JSON.stringify({ action: 'subscribe', channel }))
  }
  
  unsubscribe(channel: string) {
    this.ws?.send(JSON.stringify({ action: 'unsubscribe', channel }))
  }
}
```

#### ui 包（组件库）

主要组件：
- `Board` - 看板组件（拖拽支持）
- `TaskCard` - 任务卡片
- `ProjectCard` - 项目卡片
- `Navbar` - 顶部导航
- `Sidebar` - 侧边栏
- `NotificationPanel` - 通知面板
- `CommentList` - 评论列表
- `ActivityLog` - 活动日志
- `MarkdownEditor` - 文档编辑器

#### features 包（功能模块）

每个功能模块包含：
- `views/` - 页面组件
- `store/` - Pinia状态管理
- `hooks/` - Vue组合式函数
- `components/` - 专属组件

---

## 8. 实现要点

### 8.1 后端关键实现

#### 模块注册模式

```go
// modules/user/module.go
type UserModule struct {
    services *core.ServiceCollection
}

func Register(services *core.ServiceCollection) {
    // 注册Repository
    services.AddScoped(func(s abstract.ServiceCollectionAbstract) UserRepository {
        db := core.GetService[*gorm.DB](s)
        return NewUserRepository(db)
    })
    
    // 注册Service
    services.AddScoped(func(s abstract.ServiceCollectionAbstract) UserService {
        repo := core.GetService[UserRepository](s)
        hub := core.GetService[*WebSocketHub](s)
        return NewUserService(repo, hub)
    })
}

func RegisterRoutes(app *core.WebApplication, services *core.ServiceCollection) {
    handler := NewUserHandler(
        core.GetService[UserService](services),
        core.GetService[*auth.JWTProvider](services),
    )
    
    api := app.Group("/api/v1")
    handler.RegisterRoutes(api)
}
```

#### Controller 模式

```go
// modules/user/handler/auth_handler.go
type AuthHandler struct {
    userService UserService
    jwtProvider *auth.JWTProvider
}

func (h *AuthHandler) Routes(r abstract.RouterAbstract) {
    r.POST("/auth/register", h.Register)
    r.POST("/auth/login", h.Login)
    r.POST("/auth/logout", h.Logout).Guard(&AuthGuard{})
    r.GET("/auth/profile", h.Profile).Guard(&AuthGuard{})
    r.PUT("/auth/profile", h.UpdateProfile).Guard(&AuthGuard{})
}

func (h *AuthHandler) Register(ctx abstract.ContextAbstract) error {
    var dto RegisterDTO
    if err := ctx.Bind(&dto); err != nil {
        return abstract.BadRequest("数据格式错误")
    }
    
    user, err := h.userService.Register(&dto)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, response.Success(user))
}

func (h *AuthHandler) Login(ctx abstract.ContextAbstract) error {
    var dto LoginDTO
    if err := ctx.Bind(&dto); err != nil {
        return abstract.BadRequest("数据格式错误")
    }
    
    user, token, err := h.userService.Login(&dto)
    if err != nil {
        return abstract.Unauthorized("登录失败")
    }
    
    return ctx.JSON(200, response.Success(map[string]any{
        "user":  user,
        "token": token,
    }))
}
```

#### GORM Repository 实现

```go
// modules/task/repository/task_repo.go
type TaskRepository interface {
    Create(task *Task) error
    Update(id uint, updates map[string]any) (*Task, error)
    Delete(id uint) error
    FindByID(id uint) (*Task, error)
    FindByProject(projectID uint, filters *TaskFilters) ([]Task, error)
    FindByAssignee(userID uint) ([]Task, error)
    CountByStatus(projectID uint) (map[string]int, error)
}

type GormTaskRepository struct {
    db *gorm.DB
}

func (r *GormTaskRepository) Create(task *Task) error {
    return r.db.Create(task).Error
}

func (r *GormTaskRepository) FindByProject(projectID uint, filters *TaskFilters) ([]Task, error) {
    query := r.db.Where("project_id = ?", projectID)
    
    if filters.Status != "" {
        query = query.Where("status = ?", filters.Status)
    }
    if filters.AssigneeID != nil {
        query = query.Where("assignee_id = ?", *filters.AssigneeID)
    }
    if filters.Priority != "" {
        query = query.Where("priority = ?", filters.Priority)
    }
    
    var tasks []Task
    err := query.Order("order ASC, created_at DESC").Find(&tasks).Error
    return tasks, err
}

func (r *GormTaskRepository) CountByStatus(projectID uint) (map[string]int, error) {
    var results []struct {
        Status string
        Count  int
    }
    
    err := r.db.Model(&Task{}).
        Select("status, count(*) as count").
        Where("project_id = ?", projectID).
        Group("status").
        Find(&results).Error
    
    counts := make(map[string]int)
    for _, r := range results {
        counts[r.Status] = r.Count
    }
    return counts, err
}
```

### 8.2 前端关键实现

#### Pinia Store 示例

```typescript
// packages/features/task/src/store/taskStore.ts
import { defineStore } from 'pinia'
import { taskApi } from '@pm/core/api'
import type { Task, TaskFilters } from '@pm/core/types'

export const useTaskStore = defineStore('task', {
  state: () => ({
    tasks: [] as Task[],
    currentTask: null as Task | null,
    filters: {} as TaskFilters,
    loading: false,
  }),
  
  actions: {
    async fetchProjectTasks(projectId: number) {
      this.loading = true
      try {
        this.tasks = await taskApi.getProjectTasks(projectId, this.filters)
      } finally {
        this.loading = false
      }
    },
    
    async createTask(task: CreateTaskDTO) {
      const newTask = await taskApi.create(task)
      this.tasks.push(newTask)
      return newTask
    },
    
    async updateTaskStatus(taskId: number, status: string) {
      await taskApi.updateStatus(taskId, status)
      const task = this.tasks.find(t => t.id === taskId)
      if (task) task.status = status
    },
  },
})
```

#### WebSocket Hook

```typescript
// packages/features/notification/src/hooks/useWebSocket.ts
import { onMounted, onUnmounted } from 'vue'
import { WebSocketClient } from '@pm/core/websocket'
import { useNotificationStore } from '../store/notificationStore'
import { getToken } from '@pm/core/auth'

export function useWebSocket() {
  const ws = new WebSocketClient()
  const notificationStore = useNotificationStore()
  
  onMounted(() => {
    ws.connect(getToken())
    
    ws.on('task_updated', (payload) => {
      notificationStore.addNotification({
        type: 'task_updated',
        title: '任务更新',
        content: payload.title,
        entityId: payload.id,
      })
    })
    
    ws.on('notification', (payload) => {
      notificationStore.addNotification(payload)
    })
  })
  
  onUnmounted(() => {
    ws.disconnect()
  })
  
  return {
    subscribe: ws.subscribe,
    unsubscribe: ws.unsubscribe,
  }
}
```

---

## 9. 配置文件

### 9.1 后端配置

```yaml
# backend/config.yaml
server:
  port: "8080"
  name: "pm-server"
  mode: "debug"  # debug/release

database:
  type: "sqlite"
  path: "data.db"
  # 若切换MySQL/PostgreSQL:
  # type: "mysql"
  # host: "localhost"
  # port: 3306
  # user: "root"
  # password: "password"
  # name: "pm_db"

jwt:
  secret: "your-secret-key-change-in-production"
  access_token_ttl: 3600      # 1小时
  refresh_token_ttl: 86400    # 24小时
  issuer: "pm-server"

casbin:
  model_path: "model.conf"
  policy_path: "policy.csv"

websocket:
  enabled: true
  path: "/ws"
  max_connections: 1000
  ping_interval: 30
  ping_timeout: 10

log:
  level: "info"  # debug/info/warn/error
  format: "json" # json/console
  output: "stdout" # stdout/file

cors:
  allow_origins:
    - "http://localhost:5173"
    - "http://localhost:3000"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Authorization"
    - "Content-Type"
  allow_credentials: true
  max_age: 86400

rate_limit:
  enabled: true
  limit: 100       # 每分钟100次
  window: 60       # 60秒
```

---

## 10. 测试策略

### 10.1 后端测试

- **单元测试**：Service层业务逻辑测试
- **集成测试**：HTTP API端点测试（使用gonest testutil）
- **E2E测试**：完整业务流程测试

### 10.2 前端测试

- **组件测试**：UI组件渲染测试（Vitest）
- **Store测试**：Pinia状态管理测试
- **E2E测试**：用户操作流程测试（Playwright）

---

## 11. 验证目标清单

| 框架能力 | 验证方式 |
|----------|----------|
| Controller模式 | 模块化路由注册，每个模块使用Controller模式 |
| Guard拦截器 | AuthGuard保护路由，RoleGuard角色检查 |
| JWT认证 | 登录/注册流程，Token验证中间件 |
| Casbin RBAC | 动态权限检查，项目级/团队级权限 |
| WebSocket | 实时通知推送，任务状态变更通知 |
| DI容器 | 模块服务注册，跨模块服务依赖 |
| 配置管理 | YAML配置加载，多环境配置支持 |
| Zap日志 | 结构化日志，请求日志中间件 |
| Recovery中间件 | Panic恢复，错误统一处理 |
| CORS中间件 | 前端跨域请求支持 |
| RateLimit中间件 | API限流保护 |
| GORM集成 | SQLite CRUD，复杂查询，事务 |
| 模块化扩展 | DDD模块结构，模块独立注册 |

---

## 12. 下一步行动

确认设计后，将进入实现规划阶段：
1. 创建项目骨架（前后端目录结构）
2. 后端基础模块（database/auth/config/logger）
3. 核心业务模块（user/project/task/team）
4. WebSocket实时通知
5. 前端Monorepo搭建
6. 核心页面开发（Dashboard/看板/任务详情）
7. 集成测试与验证