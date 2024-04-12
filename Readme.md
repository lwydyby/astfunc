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

```bash
方法所在项目module为:github.com/lwydyby/astfunc 
方法的package为: 1
需要添加单元测试的方法如下:
func A(a string) string {
        return B()
}
测试方法的入参结构为:

string
测试方法的返回结构为:

string
测试方法内调用的方法入参和返回分别为:

方法名: B
入参: 
返回类型: string

```
