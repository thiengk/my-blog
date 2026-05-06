---
title: "Hướng Dẫn Khai Thác Tối Đa Sức Mạnh Kiro"
description: "Tài liệu toàn diện về Kiro AI - từ Steering, Specs, Hooks đến Powers và MCP. Học cách setup dự án, tự động hóa workflow và tối ưu hiệu suất làm việc với AI."
date: 2026-05-06
category: "cong-nghe"
tags: ["kiro", "ai", "development", "automation", "productivity", "tutorial", "devops"]
draft: false
coverImage: "https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=400&fit=crop"
---

> Tài liệu này áp dụng cho **mọi dự án**, không riêng dự án nào cụ thể.
> 
> Mục tiêu: giúp bạn hiểu cần chuẩn bị gì, cấu trúc ra sao, và triển khai thế nào để khai thác tối đa Kiro.

---

## 📑 Mục Lục

1. [Tổng Quan Khả Năng](#1-tổng-quan-khả-năng)
2. [Cấu Trúc Thư Mục .kiro/](#2-cấu-trúc-thư-mục-kiro)
3. [Steering - Điều Hướng AI](#3-steering---điều-hướng-ai)
4. [Specs - Phát Triển Theo Đặc Tả](#4-specs---phát-triển-theo-đặc-tả)
5. [Hooks - Tự Động Hóa](#5-hooks---tự-động-hóa)
6. [Powers - Mở Rộng Khả Năng](#6-powers---mở-rộng-khả-năng)
7. [MCP - Model Context Protocol](#7-mcp---model-context-protocol)
8. [Chat Context - Tận Dụng Ngữ Cảnh](#8-chat-context---tận-dụng-ngữ-cảnh)
9. [Quy Trình Làm Việc Tối Ưu](#9-quy-trình-làm-việc-tối-ưu)
10. [Checklist Khởi Tạo Dự Án Mới](#10-checklist-khởi-tạo-dự-án-mới)

---

## 1. Tổng Quan Khả Năng

Kiro là một **AI development environment**, không chỉ là chat. Các khả năng chính:

| Khả năng | Mô tả |
|-----------|--------|
| Đọc/viết code | Đọc hiểu toàn bộ codebase, viết code mới, refactor |
| Chạy terminal | Thực thi commands, build, test, deploy |
| Spec workflow | Biến ý tưởng → requirements → design → tasks → code |
| Hooks | Tự động hóa quy trình khi có sự kiện trong IDE |
| Steering | Luôn tuân theo rules/standards bạn đặt ra, mọi cuộc chat |
| Powers/MCP | Mở rộng khả năng với external tools |
| Web search | Tra cứu thông tin mới nhất từ internet |
| Phân tích ảnh/tài liệu | Đọc mockup UI, PDF, DOCX kéo vào chat |

**Nguyên tắc cốt lõi**: Context và rules càng rõ ràng → Output càng chính xác và nhất quán.

---

## 2. Cấu Trúc Thư Mục .kiro/

Mỗi dự án cần một thư mục `.kiro/` ở root. Đây là "bộ não" giúp Kiro hiểu dự án của bạn.

### Cấu trúc đầy đủ

```
.kiro/
├── steering/              # [QUAN TRỌNG NHẤT] Rules và context cho AI
│   ├── language.md        # Ngôn ngữ giao tiếp (VD: tiếng Việt)
│   ├── product.md         # Mô tả sản phẩm, tính năng, kiến trúc
│   ├── structure.md       # Cấu trúc thư mục dự án
│   ├── tech.md            # Tech stack, commands (build/test/run)
│   ├── standards.md       # Coding standards, conventions
│   ├── patterns.md        # Design patterns đang dùng
│   ├── testing.md         # Testing strategy
│   └── {custom}.md        # Bất kỳ context nào bạn muốn AI nhớ
├── specs/                 # Đặc tả features/bugfixes
│   └── {feature-name}/
│       ├── .config.kiro   # Config (specType, workflowType)
│       ├── requirements.md
│       ├── design.md
│       └── tasks.md
├── hooks/                 # Automation hooks (JSON files)
│   └── {hook-name}.json
└── settings/
    └── mcp.json           # MCP server configuration
```

### Mức độ ưu tiên khi setup

| Ưu tiên | Thành phần | Lý do |
|---------|-----------|-------|
| 🔴 Bắt buộc | steering/ (ít nhất product + tech) | Kiro cần hiểu dự án |
| 🟡 Nên có | steering/standards.md | Đảm bảo code nhất quán |
| 🟡 Nên có | Hooks cơ bản (lint, test) | Tự động kiểm tra chất lượng |
| 🟢 Tùy chọn | specs/ | Dùng khi có feature phức tạp |
| 🟢 Tùy chọn | settings/mcp.json | Khi cần external tools |

---

## 3. Steering - Điều Hướng AI

### Steering là gì?

Steering files (`.kiro/steering/*.md`) là các file markdown chứa rules và context mà Kiro **luôn nhớ** trong mọi cuộc hội thoại. Bạn viết một lần, Kiro tuân theo mãi mãi.

### 3 chế độ inclusion

| Chế độ | Front-matter | Khi nào active | Use case |
|--------|-------------|----------------|----------|
| **Always** (mặc định) | Không cần | Mọi cuộc chat | Product info, tech stack, standards |
| **File Match** | `inclusion: fileMatch` + `fileMatchPattern` | Khi đọc file matching | Language-specific rules |
| **Manual** | `inclusion: manual` | Khi user gõ #TênFile | Reference docs ít dùng |

### Template steering files cho mọi dự án

#### language.md - Ngôn ngữ giao tiếp

```markdown
# Ngôn ngữ giao tiếp

Luôn trả lời bằng **tiếng Việt** khi tương tác với người dùng.

## Quy tắc
- Giải thích, tóm tắt, hướng dẫn: tiếng Việt
- Tên biến, hàm, class: tiếng Anh (chuẩn lập trình)
- Comment trong code: tiếng Anh hoặc Việt tuỳ ngữ cảnh
- Commit message: tiếng Việt hoặc Anh tuỳ team
```

#### product.md - Mô tả sản phẩm

```markdown
# Product Overview

[Mô tả ngắn gọn sản phẩm là gì, phục vụ ai]

## Tính năng chính
- Feature A: [mô tả]
- Feature B: [mô tả]

## Kiến trúc tổng quan
[Diagram hoặc mô tả flow: Frontend → Backend → Database]

## Deployment
- Frontend: [hosting platform]
- Backend: [hosting platform]
- Database: [service]
```

#### tech.md - Tech stack & Commands

```markdown
# Tech Stack

## Frontend
- Framework: [VD: React, Vue, Astro, Next.js]
- CSS: [VD: TailwindCSS, styled-components]
- State: [VD: Zustand, Redux, Svelte stores]

## Backend
- Language: [VD: Go, Node.js, Python]
- Framework: [VD: Gin, Express, FastAPI]
- Database: [VD: PostgreSQL, MongoDB]
- Cache: [VD: Redis]

## Commands
\`\`\`bash
# Install
[command]

# Dev
[command]

# Build
[command]

# Test
[command]

# Lint
[command]
\`\`\`
```

#### structure.md - Cấu trúc dự án

```markdown
# Project Structure

\`\`\`
src/
├── components/  # [mô tả]
├── pages/       # [mô tả]
├── lib/         # [mô tả]
├── services/    # [mô tả]
└── types/       # [mô tả]
\`\`\`

[Giải thích ngắn về convention đặt tên file, tổ chức module]
```

#### standards.md - Coding standards

```markdown
# Coding Standards

## Naming Conventions
- Files: [kebab-case / PascalCase / camelCase]
- Functions: [camelCase / snake_case]
- Constants: [UPPER_SNAKE_CASE]
- Types/Interfaces: [PascalCase, prefix I hay không]

## Error Handling
- [Mô tả pattern: try-catch, Result type, error wrapping...]

## Testing
- Unit test cho mọi business logic
- Integration test cho API endpoints
- Test file đặt cạnh source file / trong folder __tests__

## Git Conventions
- Commit format: \`type: description\`
- Types: feat, fix, refactor, docs, test, chore
- Branch: feature/xxx, fix/xxx, refactor/xxx

## Code Review Checklist
- [ ] Error handling đầy đủ
- [ ] Input validation
- [ ] No hardcoded values
- [ ] Tests cover happy path + edge cases
```

#### Conditional steering (File Match)

```markdown
---
inclusion: fileMatch
fileMatchPattern: '**/*.go'
---

# Go-Specific Rules

- Luôn handle errors, không dùng \`_\` bỏ qua error
- Dùng \`context.Context\` cho mọi function có I/O
- Interface nhỏ (1-3 methods), đặt ở package consumer
- Table-driven tests
```

```markdown
---
inclusion: fileMatch
fileMatchPattern: '**/*.{ts,tsx,svelte}'
---

# Frontend-Specific Rules

- Components phải có TypeScript props interface
- Dùng semantic HTML elements
- Accessibility: aria-label cho interactive elements
- Responsive: mobile-first approach
```

### Tham chiếu file trong steering

Dùng cú pháp `#[[file:path]]` để Kiro tự động đọc thêm context:

```markdown
# API Documentation

Tham khảo API spec:
#[[file:docs/openapi.yaml]]

Database schema:
#[[file:migrations/schema.sql]]
```

---

## 4. Specs - Phát Triển Theo Đặc Tả

### Spec workflow là gì?

Quy trình có hệ thống biến ý tưởng thành code:

```
Ý tưởng → Requirements → Design → Tasks → Implementation
         (bạn review)  (bạn review) (bạn review) (Kiro code)
```

Mỗi bước bạn review và confirm trước khi tiếp. Đảm bảo code đúng từ đầu, tránh refactor tốn thời gian.

### Khi nào dùng Spec?

| Dùng Spec ✅ | Chat trực tiếp ❌ |
|-------------|-------------------|
| Feature mới phức tạp (> 3 files) | Thay đổi nhỏ (1-2 files) |
| Fix bug phức tạp (nhiều nguyên nhân) | Fix bug đơn giản (rõ nguyên nhân) |
| Refactor lớn | Rename, move file |
| Tính năng cần thiết kế DB schema | Thêm 1 field vào API |
| Feature liên quan nhiều components | Sửa CSS, UI nhỏ |

### 2 loại Spec

**Feature Spec** - Tính năng mới:
- **Requirements-first**: Bạn có business needs rõ → thu thập requirements trước → design sau
- **Design-first**: Bạn biết technical approach → thiết kế trước → derive requirements sau

**Bugfix Spec** - Sửa lỗi:
- Phân tích bug condition (điều kiện gây bug)
- Viết test chứng minh bug tồn tại
- Fix code
- Verify test pass

### Cách bắt đầu

Nói tự nhiên, Kiro sẽ hỏi bạn chọn workflow:

```
"Tôi muốn thêm feature X"     → Feature Spec
"App bị crash khi làm Y"      → Bugfix Spec
"Implement hệ thống Z"        → Feature Spec
```

### Thực thi Tasks

Sau khi có task list:
- "Chạy task 1" → Kiro implement task cụ thể
- "Chạy tất cả tasks" → Kiro implement tuần tự
- Review kết quả → feedback → iterate

---

## 5. Hooks - Tự Động Hóa

### Hooks là gì?

Hooks tự động trigger hành động khi có sự kiện trong IDE. Giống CI/CD nhưng chạy local, real-time.

### Các loại event

| Event | Trigger khi | Ví dụ use case |
|-------|------------|----------------|
| fileEdited | User save file | Auto-lint, auto-format |
| fileCreated | User tạo file mới | Generate boilerplate |
| fileDeleted | User xóa file | Cleanup imports |
| promptSubmit | User gửi message | Pre-process input |
| agentStop | Agent hoàn thành | Summary, notification |
| preToolUse | Trước khi tool chạy | Access control, review |
| postToolUse | Sau khi tool chạy | Validate output |
| preTaskExecution | Trước khi task bắt đầu | Setup environment |
| postTaskExecution | Sau khi task xong | Run tests |
| userTriggered | User bấm nút | Manual workflows |

### 2 loại action

| Action | Mô tả | Khi nào dùng |
|--------|--------|-------------|
| askAgent | Gửi prompt cho AI | Review, validate, suggest |
| runCommand | Chạy shell command | Lint, test, build, format |

### Schema hook file

```json
{
  "name": "Tên hook (hiển thị)",
  "version": "1.0.0",
  "description": "Mô tả hook làm gì (optional)",
  "when": {
    "type": "eventType",
    "patterns": ["glob patterns (cho file events)"],
    "toolTypes": ["categories hoặc regex (cho tool events)"]
  },
  "then": {
    "type": "askAgent hoặc runCommand",
    "prompt": "prompt cho AI (askAgent)",
    "command": "shell command (runCommand)"
  }
}
```

### Hooks mẫu cho mọi dự án

**Auto-format khi save:**

```json
{
  "name": "Format on Save",
  "version": "1.0.0",
  "when": {
    "type": "fileEdited",
    "patterns": ["*.ts", "*.tsx", "*.js", "*.jsx"]
  },
  "then": {
    "type": "runCommand",
    "command": "npx prettier --write ${file}"
  }
}
```

**Auto-test khi save:**

```json
{
  "name": "Run Related Tests",
  "version": "1.0.0",
  "when": {
    "type": "fileEdited",
    "patterns": ["*.ts", "*.tsx"]
  },
  "then": {
    "type": "runCommand",
    "command": "npx vitest --run --reporter=verbose"
  }
}
```

**Review trước khi write code:**

```json
{
  "name": "Pre-Write Review",
  "version": "1.0.0",
  "when": {
    "type": "preToolUse",
    "toolTypes": ["write"]
  },
  "then": {
    "type": "askAgent",
    "prompt": "Verify: 1) Tuân theo coding standards 2) Error handling đầy đủ 3) Không có security issues 4) Naming conventions đúng"
  }
}
```

**Test sau mỗi task:**

```json
{
  "name": "Post-Task Tests",
  "version": "1.0.0",
  "when": {
    "type": "postTaskExecution"
  },
  "then": {
    "type": "runCommand",
    "command": "npm test"
  }
}
```

**Go test khi save .go:**

```json
{
  "name": "Go Test on Save",
  "version": "1.0.0",
  "when": {
    "type": "fileEdited",
    "patterns": ["*.go"]
  },
  "then": {
    "type": "runCommand",
    "command": "go test ./..."
  }
}
```

### Cách tạo hook

3 cách:
1. Chat: "Tạo hook auto-lint khi save file .ts"
2. Command Palette → "Open Kiro Hook UI"
3. Explorer sidebar → "Agent Hooks" section

---

## 6. Powers - Mở Rộng Khả Năng

### Powers là gì?

Powers là packages gồm documentation, workflow guides (steering), và MCP servers. Cài thêm để Kiro có khả năng chuyên biệt cho domain cụ thể.

### Cách quản lý

- Chat: "Mở powers panel" hoặc "Cài power X"
- Kiro mở panel để browse và install
- Powers tự động integrate vào workflow

### Ví dụ use cases

| Power | Khả năng thêm |
|-------|---------------|
| Build a Power | Tạo custom power cho team |
| AWS Documentation | Tra cứu AWS docs real-time |
| Database tools | Query DB trực tiếp từ chat |

---

## 7. MCP - Model Context Protocol

### MCP là gì?

Protocol cho phép Kiro kết nối external tools/servers. Mở rộng khả năng vượt xa built-in tools.

### Cấu hình

**Workspace level**: `.kiro/settings/mcp.json`  
**User level (global)**: `~/.kiro/settings/mcp.json`

Precedence: user config < workspace config (workspace override user)

```json
{
  "mcpServers": {
    "server-name": {
      "command": "uvx",
      "args": ["package-name@latest"],
      "env": {
        "ENV_VAR": "value"
      },
      "disabled": false,
      "autoApprove": ["tool-name-1", "tool-name-2"]
    }
  }
}
```

### Yêu cầu cài đặt

1. Cài `uv` (Python package manager):
   - `pip install uv`
   - Hoặc Homebrew: `brew install uv`
   - Hoặc xem: https://docs.astral.sh/uv/getting-started/installation/

2. `uvx` sẽ tự download và chạy MCP servers (không cần install riêng từng server)

### MCP servers phổ biến

| Server | Mục đích |
|--------|----------|
| awslabs.aws-documentation-mcp-server | Tra cứu AWS docs |
| mcp-server-postgres | Query PostgreSQL |
| mcp-server-github | GitHub API |
| mcp-server-filesystem | File operations nâng cao |
| mcp-server-fetch | HTTP requests |

### Quản lý

- Servers tự reconnect khi config thay đổi
- Command Palette → tìm "MCP" để xem commands
- `autoApprove`: list tool names không cần confirm mỗi lần gọi

---

## 8. Chat Context - Tận Dụng Ngữ Cảnh

### Các cách cung cấp context

| Cách | Mô tả | Khi nào dùng |
|------|--------|-------------|
| #File | Kéo file cụ thể vào | Hỏi về file cụ thể |
| #Folder | Kéo folder vào | Hỏi về module/package |
| #Problems | Problems trong file hiện tại | Fix lỗi IDE đang báo |
| #Terminal | Output terminal gần nhất | Debug command errors |
| #Git Diff | Thay đổi git chưa commit | Code review |
| Kéo ảnh | Phân tích hình ảnh | UI mockup, error screenshot |
| Kéo PDF/DOCX | Đọc tài liệu | Specs, requirements docs |

### Tips tận dụng context

| Tình huống | Cách làm |
|-----------|----------|
| Implement UI từ mockup | Kéo ảnh mockup + "Implement giao diện này" |
| Fix lỗi từ screenshot | Kéo screenshot + "Fix lỗi này" |
| Code review | #Git Diff + "Review changes" |
| Fix IDE errors | #Problems + "Fix tất cả" |
| Debug terminal error | #Terminal + "Giải thích và fix" |
| Hỏi về logic cụ thể | #File + "Giải thích function X" |
| Implement từ spec doc | Kéo PDF + "Implement theo spec này" |

---

## 9. Quy Trình Làm Việc Tối Ưu

### Daily workflow

```
┌─────────────────────────────────────────────────────┐
│  1. MỞ DỰ ÁN                                       │
│     → Kiro tự load steering, hiểu context ngay     │
├─────────────────────────────────────────────────────┤
│  2. FEATURE PHỨC TẠP                                │
│     "Tôi muốn thêm feature X"                      │
│     → Spec: Requirements → Design → Tasks → Code   │
├─────────────────────────────────────────────────────┤
│  3. THAY ĐỔI NHỎ                                   │
│     "Thêm field Y vào API Z"                       │
│     → Chat trực tiếp, implement ngay               │
├─────────────────────────────────────────────────────┤
│  4. FIX BUG                                         │
│     "Lỗi khi user làm X"                           │
│     → Bugfix spec (phức tạp) hoặc fix ngay (đơn giản) │
├─────────────────────────────────────────────────────┤
│  5. REVIEW                                          │
│     "#Git Diff review cho tôi"                      │
│     → AI review code, suggest improvements          │
├─────────────────────────────────────────────────────┤
│  6. REFACTOR                                        │
│     "Refactor module X cho clean hơn"              │
│     → Spec (lớn) hoặc chat (nhỏ)                  │
├─────────────────────────────────────────────────────┤
│  7. HOOKS CHẠY NỀN                                  │
│     Lint, test, format tự động                      │
│     → Bạn focus logic, Kiro lo quality              │
└─────────────────────────────────────────────────────┘
```

### 5 nguyên tắc vàng

**1. Steering càng chi tiết → Output càng chính xác**
- Viết rõ conventions, patterns, do's and don'ts
- Kiro đọc steering MỌI lần chat, không cần nhắc lại

**2. Dùng Spec cho feature phức tạp**
- Feature > 3 files → dùng Spec
- Tránh "code ngay" rồi phải refactor

**3. Hooks cho repetitive tasks**
- Lint mỗi lần save? → Hook
- Test mỗi lần xong task? → Hook
- Review trước khi write? → Hook

**4. Context rõ ràng = Output tốt**
- Kéo file liên quan vào chat
- Mô tả expected behavior cụ thể
- Cung cấp error messages đầy đủ

**5. Iterate, không one-shot**
- Review từng bước trong spec
- Feedback cụ thể: "phần X tốt, nhưng Y cần đổi vì Z"
- Steering files nên update khi dự án phát triển

### Anti-patterns (tránh làm)

| ❌ Tránh | ✅ Nên làm |
|---------|-----------|
| Nói "code feature X" không context | Mô tả rõ feature, expected behavior |
| Không có steering files | Tạo ít nhất product.md + tech.md |
| Dùng Spec cho mọi thứ nhỏ | Chat trực tiếp cho thay đổi < 3 files |
| Không review output | Review từng bước, feedback cụ thể |
| Copy-paste error không context | Kéo #Terminal hoặc #Problems |

---

## 10. Checklist Khởi Tạo Dự Án Mới

Khi bắt đầu dự án mới với Kiro, follow checklist này:

### Bước 1: Tạo .kiro/steering/ (5-15 phút)

- [ ] `language.md` - Ngôn ngữ giao tiếp
- [ ] `product.md` - Sản phẩm là gì, features, kiến trúc
- [ ] `tech.md` - Tech stack, commands (install/dev/build/test)
- [ ] `structure.md` - Cấu trúc thư mục, giải thích từng folder
- [ ] `standards.md` - Coding conventions, naming, error handling, git

### Bước 2: Tạo hooks cơ bản (2-5 phút)

- [ ] Auto-lint/format khi save
- [ ] Auto-test khi save (optional, có thể chậm)
- [ ] Post-task test (chạy test sau mỗi spec task)

### Bước 3: Setup MCP (optional, 5 phút)

- [ ] Cài `uv` nếu chưa có
- [ ] Tạo `.kiro/settings/mcp.json` với servers cần thiết

### Bước 4: Bắt đầu làm việc

- [ ] Feature phức tạp đầu tiên → dùng Spec workflow
- [ ] Thay đổi nhỏ → chat trực tiếp
- [ ] Review output → feedback → iterate
- [ ] Update steering files khi phát hiện pattern mới

---

## 📚 Phụ Lục: Quick Reference

### Câu lệnh chat hữu ích

```bash
# Feature mới
"Tôi muốn thêm [feature]"

# Fix bug
"[Mô tả bug], fix giúp tôi"

# Giải thích code
"#[file] giải thích function X"

# Review
"#Git Diff review cho tôi"

# Tạo hook
"Tạo hook [mô tả hành vi]"

# Refactor
"Refactor [file/module] cho [mục tiêu]"

# Hỏi kiến trúc
"Giải thích flow từ [A] đến [B] trong dự án"

# Generate tests
"Viết unit tests cho [file/function]"

# Setup
"Setup [tool/framework] cho dự án này"
```

### Steering file front-matter options

```yaml
---
# File Match: active khi đọc file matching pattern
inclusion: fileMatch
fileMatchPattern: '**/*.go'
---

---
# Manual: active khi user gõ #TênFile trong chat
inclusion: manual
---

# Không có front-matter = Always active (mặc định)
```

---

## 🎯 Kết Luận

Bạn đã học:
- ✅ Cấu trúc thư mục `.kiro/` và vai trò từng phần
- ✅ Steering files để điều hướng AI
- ✅ Spec workflow cho features phức tạp
- ✅ Hooks để tự động hóa quy trình
- ✅ Powers và MCP để mở rộng khả năng
- ✅ Chat context để tận dụng ngữ cảnh
- ✅ Quy trình làm việc tối ưu

**Bước tiếp theo:** Áp dụng vào dự án của bạn! Bắt đầu với steering files cơ bản, sau đó dần thêm hooks và specs khi cần.

---

*"Context và rules càng rõ ràng → Output càng chính xác và nhất quán."*

**Chúc bạn thành công với Kiro! 🚀**
