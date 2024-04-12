# 目标
使用ast获取指定方法的入参,出参类型,以及这个方法中调用的方法的入参和出参

# 使用场景
将这些信息提供给AI工具,由AI工具生成单测,相比于单纯将方法提供给AI,AI可以获取更多的上下文,生成的单测也会更加准确

# 使用示例

```go
package main

import (
	"fmt"

	"github.com/lwydyby/astfunc"
)

func main() {
	result := astfunc.GetFuncWithDependency(".", "A")
	fmt.Println(result)
}
```

结果

```go

```