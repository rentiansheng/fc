# code 目录

## 功能简介

本模块通过对 Go 语言 AST（抽象语法树）的分析，实现了对比两个代码文件中函数级别的差异，判断函数是否相同。适用于对比不同版本、不同实现的 Go 源代码文件，辅助代码一致性检查、重构、代码审查等场景。

## 主要特性

- 支持对比两个 Go 文件中所有函数（包括方法、匿名函数等）的实现差异。
- 通过 AST 结构，能忽略格式、注释等无关内容，仅关注函数实现的实际差异。
- 提供详细的对比结果，包括相同函数、不同函数的列表。
- 包含完整的单元测试，确保对比逻辑的准确性。

## 目录结构

```
code/
├── ast_compare.go        # AST 节点深度对比核心实现
├── ast_compare_test.go   # ast_compare 单元测试
├── diff.go               # 代码整体对比逻辑实现（主入口）
├── diff_test.go          # diff 相关单元测试
├── func_name.go          # 函数、方法、匿名函数的 AST 提取与定位
├── func_name_test.go     # func_name 相关单元测试
├── READEME.md            # 功能简述（建议重命名为 README.md）
```

## 使用方法

参考 `diff.go` 提供的 `CodeDiff` 结构体和 `Diff()` 方法，传入两个 `*ast.File` 及对应的 `*token.FileSet`，即可获得相同和不同的函数列表。

```go
cd := NewCodeDiff(f1, f2, f1Set, f2Set)
sames, diffs, err := cd.Diff()
```

更详细的用法可参考 `diff_test.go` 和 `ast_compare_test.go` 提供的测试代码。

## 测试

本目录提供丰富的单元测试用例，覆盖了函数、方法、匿名函数、结构体变更等多种场景。可直接通过 `go test` 运行。

```
go test ./code/...
```

## 注意事项

- 目前仅支持 Go 语言源文件的对比。
- 对于较大文件或极端复杂的 AST，性能可能受限于反射和遍历深度。
