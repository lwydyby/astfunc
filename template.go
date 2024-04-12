package astfunc

var tpl = `
方法所在项目module为:{{.Mod}} 
方法的package为: {{.Package}}
需要添加单元测试的方法如下:
{{.Code}}
测试方法的入参结构为:
{{range .Params}}
{{.}}{{end}}
测试方法的返回结构为:
{{range .Returns}}
{{.}}{{end}}
测试方法内调用的方法入参和返回分别为:
{{range .Funcs}}
方法名: {{.MethodName}}
入参: {{range .Params}}{{.}}{{end}}
返回类型: {{range .Returns}}{{.}}{{end}}
{{end}}
`
