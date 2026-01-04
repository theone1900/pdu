# PDU - PostgreSQL Data Unloader 迁移到 Golang 方案

## 1. 可行性分析

### 1.1 核心功能模块

PDU 主要由以下核心模块组成：

| 模块 | 功能 | 迁移复杂度 |
|------|------|------------|
| 命令处理 | 解析和执行用户命令 | 低 |
| 数据读取 | 读取和解析 PostgreSQL 数据文件 | 中 |
| 数据解码 | 解码各种 PostgreSQL 数据类型 | 中 |
| WAL 处理 | 解析和分析 WAL 文件 | 高 |
| 元数据管理 | 管理数据库元数据 | 低 |
| 动态数组 | 实现动态数组数据结构 | 低（可使用 Go 切片替代） |
| 字符串处理 | 处理变长字符串 | 低（可使用 Go 字符串和 bytes.Buffer） |
| TOAST 处理 | 处理压缩的大对象数据 | 中（需要集成 LZ4 和 zlib） |

### 1.2 技术可行性

| 技术点 | 可行性 | 解决方案 |
|--------|--------|----------|
| 直接文件访问 | 高 | 使用 Go 标准库 `os` 和 `io` 包 |
| 结构体解析 | 高 | 使用 Go 的 `encoding/binary` 包和结构体标签 |
| 并发处理 | 高 | 使用 Go 的 goroutine 和 channel |
| LZ4 解压缩 | 高 | 使用 Go 第三方库 `github.com/pierrec/lz4/v4` |
| zlib 解压缩 | 高 | 使用 Go 标准库 `compress/zlib` 包 |
| 命令行界面 | 高 | 使用 Go 标准库 `flag` 或第三方库 `github.com/spf13/cobra` |

### 1.3 优势

将 PDU 迁移到 Golang 有以下优势：

1. **跨平台支持**：Go 编译的二进制文件可以在不同平台上运行，无需重新编译
2. **内存安全**：Go 的垃圾回收机制和类型系统可以减少内存泄漏和指针错误
3. **并发支持**：内置的 goroutine 和 channel 可以更好地处理并发任务
4. **简洁的语法**：Go 的语法简洁明了，易于维护和扩展
5. **丰富的标准库**：减少对第三方库的依赖
6. **静态编译**：生成单个二进制文件，便于部署和分发
7. **更好的错误处理**：Go 的错误处理机制可以提高代码的可靠性

## 2. 架构设计

### 2.1 目录结构

```
pdu/
├── cmd/
│   └── pdu/
│       └── main.go           # 主入口文件
├── internal/
│   ├── cmd/                  # 命令处理模块
│   ├── decoder/              # 数据解码模块
│   ├── fileio/               # 文件 I/O 模块
│   ├── metadata/             # 元数据管理模块
│   ├── pager/                # 页面处理模块
│   ├── parser/               # 命令解析模块
│   ├── types/                # 数据类型定义
│   ├── wal/                  # WAL 处理模块
│   └── toast/                # TOAST 处理模块
├── pkg/
│   ├── array/                # 数组工具包
│   ├── pgtypes/              # PostgreSQL 数据类型定义
│   └── utils/                # 通用工具函数
├── go.mod                    # Go 模块定义
└── README.md                 # 项目说明文档
```

### 2.2 核心包设计

#### 2.2.1 internal/fileio

负责直接文件访问和操作，包括：
- 数据文件打开和关闭
- 页面读取和写入
- 文件系统操作

#### 2.2.2 internal/pager

负责 PostgreSQL 页面结构的解析和处理，包括：
- 页面头解析
- 行指针管理
- 元组提取

#### 2.2.3 internal/decoder

负责各种 PostgreSQL 数据类型的解码，包括：
- 数值类型解码
- 文本类型解码
- 布尔类型解码
- 复合类型解码
- JSON/XML 类型解码

#### 2.2.4 internal/wal

负责 WAL 文件的解析和分析，包括：
- WAL 文件头解析
- WAL 记录解码
- 事务信息提取

#### 2.2.5 internal/metadata

负责数据库元数据的管理，包括：
- 数据库结构管理
- 表结构管理
- 属性结构管理
- 模式结构管理

#### 2.2.6 internal/cmd

负责命令的执行，包括：
- 引导命令（bootstrap）
- 数据导出命令（unload）
- WAL 扫描命令（scan）
- 恢复命令（restore）
- 扫描删除表命令（dropscan）

#### 2.2.7 pkg/pgtypes

定义 PostgreSQL 数据类型的 Go 结构体，包括：
- PageHeaderData
- HeapPageHeaderData
- ItemIdData
- HeapTupleHeaderData
- 各种数据类型的编码规则

### 2.3 数据结构设计

将 C 语言的结构体转换为 Go 的结构体，例如：

**C 语言结构体**：
```c
typedef struct HeapPageHeaderData
{
    PageXLogRecPtr pd_lsn;
    uint16        pd_checksum;
    uint16        pd_flags;
    LocationIndex pd_lower;
    LocationIndex pd_upper;
    LocationIndex pd_special;
    uint16        pd_pagesize_version;
    TransactionId pd_prune_xid;
    ItemIdData    pd_linp[FLEXIBLE_ARRAY_MEMBER];
} HeapPageHeaderData;
```

**Go 结构体**：
```go
type PageXLogRecPtr struct {
    XLogID  uint32
    XRecOff uint32
}

type HeapPageHeaderData struct {
    PDLSN             PageXLogRecPtr
    PDChecksum        uint16
    PDFlags           uint16
    PDLower           uint16
    PDUpper           uint16
    PDSpecial         uint16
    PDPagesizeVersion uint16
    PDPruneXID        uint32
    // PDLinp is a flexible array, we'll handle it differently in Go
}
```

### 2.4 接口设计

定义模块之间的交互接口，例如：

```go
// FileReader 定义文件读取接口
type FileReader interface {
    Open(path string) error
    Close() error
    ReadPage(pageNumber int64) ([]byte, error)
    ReadBytes(offset int64, length int) ([]byte, error)
}

// PageParser 定义页面解析接口
type PageParser interface {
    ParsePage(pageData []byte) (*PageHeader, error)
    GetTuples(pageData []byte) ([]*Tuple, error)
}

// TupleDecoder 定义元组解码接口
type TupleDecoder interface {
    DecodeTuple(tupleData []byte, attrDesc []*AttributeDescriptor) (*DecodedTuple, error)
}
```

## 3. 迁移计划

### 3.1 分阶段迁移

| 阶段 | 任务 | 时间估算 |
|------|------|----------|
| 1 | 搭建 Go 项目结构和基础框架 | 1 周 |
| 2 | 迁移核心数据结构和工具函数 | 2 周 |
| 3 | 实现文件 I/O 和页面解析功能 | 2 周 |
| 4 | 实现数据解码功能 | 3 周 |
| 5 | 实现元数据管理功能 | 1 周 |
| 6 | 实现命令处理和执行功能 | 2 周 |
| 7 | 实现 WAL 处理功能 | 3 周 |
| 8 | 实现 TOAST 处理功能 | 1 周 |
| 9 | 测试和调试 | 2 周 |
| 10 | 性能优化和文档完善 | 1 周 |

### 3.2 关键里程碑

1. 完成基础框架搭建，能够编译和运行
2. 实现基本的数据文件读取和解析
3. 实现基本的数据类型解码
4. 实现完整的命令处理系统
5. 实现 WAL 文件处理功能
6. 完成全部功能测试
7. 发布第一个 Beta 版本

## 4. 技术选型

### 4.1 核心库

| 功能 | 库名 | 用途 |
|------|------|------|
| 命令行框架 | github.com/spf13/cobra | 命令行界面开发 |
| 配置管理 | github.com/spf13/viper | 配置文件处理 |
| LZ4 压缩 | github.com/pierrec/lz4/v4 | TOAST 数据解压缩 |
| zlib 压缩 | compress/zlib | 压缩数据处理 |
| 测试框架 | testing | 单元测试和集成测试 |

### 4.2 开发工具

| 工具 | 用途 |
|------|------|
| Go 1.21+ | 开发语言和工具链 |
| gofmt | 代码格式化 |
| golint | 代码质量检查 |
| go test | 测试执行 |
| dlv | 调试工具 |

## 5. 性能考虑

### 5.1 并发设计

利用 Go 的 goroutine 和 channel 实现并发处理：

1. **并行页面读取**：使用 goroutine 并行读取多个页面
2. **并行元组解码**：使用 goroutine 并行解码多个元组
3. **并行文件处理**：使用 goroutine 并行处理多个数据文件

### 5.2 内存管理

1. **减少内存分配**：使用 sync.Pool 复用对象，减少垃圾回收压力
2. **预分配内存**：对于已知大小的数据，预先分配内存
3. **避免不必要的复制**：使用指针和切片引用，避免数据复制

### 5.3 I/O 优化

1. **批量读取**：使用较大的缓冲区批量读取数据
2. **异步 I/O**：使用 Go 的异步 I/O 特性提高 I/O 性能
3. **文件系统优化**：使用适当的文件系统标志和选项

## 6. 兼容性考虑

1. **PostgreSQL 版本兼容性**：支持 PostgreSQL 14-18 版本
2. **跨平台兼容性**：支持 Linux、Windows 和 macOS
3. **数据格式兼容性**：确保生成的数据格式与原 PDU 工具兼容
4. **命令行兼容性**：保持与原 PDU 工具相同的命令行接口

## 7. 风险评估

### 7.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| WAL 格式复杂度 | 高 | 参考 PostgreSQL 源代码，逐步实现 |
| 数据类型多样性 | 中 | 按优先级实现，先支持常用类型 |
| 性能问题 | 中 | 进行性能测试和优化，使用并发处理 |
| 内存管理 | 低 | 使用 Go 的内存管理特性，避免内存泄漏 |

### 7.2 项目风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 迁移时间过长 | 高 | 分阶段迁移，优先实现核心功能 |
| 团队学习曲线 | 中 | 提供培训和文档，逐步熟悉 Go 语言 |
| 兼容性问题 | 中 | 进行充分的测试，确保与原工具兼容 |

## 8. 总结

将 PDU 迁移到 Golang 是可行的，并且可以带来以下好处：

1. **更好的跨平台支持**：可以在不同操作系统上运行，无需重新编译
2. **更高的代码可靠性**：Go 的类型系统和内存管理可以减少错误
3. **更好的并发性能**：内置的 goroutine 和 channel 可以提高处理效率
4. **更简洁的代码**：Go 的语法简洁明了，易于维护和扩展
5. **更好的生态系统**：丰富的标准库和第三方库可以加速开发

通过分阶段迁移计划，可以逐步实现功能，降低风险，确保项目成功。建议先实现核心功能，然后逐步扩展到完整功能。