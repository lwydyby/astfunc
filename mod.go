package astfunc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// FindGoMod 向上遍历目录寻找 go.mod 文件
func FindGoMod(path string) (string, *modfile.File) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	// 遍历路径直到达到根目录
	for {
		goModPath := filepath.Join(absPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return absPath, ParseGoMod(goModPath)
		}

		// 移动到上一级目录
		parentPath := filepath.Dir(absPath)
		if parentPath == absPath {
			// 如果已经到达根目录，则停止
			break
		}
		absPath = parentPath
	}

	panic(fmt.Errorf("go.mod not found"))
}

func GetModCachePath(modFile *modfile.File, modPath string, dependency string) string {
	path, version, ok := FindDependencyVersion(modFile, dependency)
	if !ok {
		if !strings.Contains(dependency, modFile.Module.Mod.Path) {
			return ""
		}
		return modPath + "/" + strings.ReplaceAll(dependency, modFile.Module.Mod.Path, "")
	}
	for _, replace := range modFile.Replace {
		if replace.Old.Path == path {
			if len(replace.New.Version) == 0 {
				return filepath.Join(modPath, replace.New.Path)
			}
			version = replace.New.Version
			break
		}
	}
	result := GetModPath() + "/" + path + "@" + version
	if path != dependency {
		result += strings.ReplaceAll(dependency, path, "")
	}
	return result
}

func ParseGoMod(goModPath string) *modfile.File {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		panic(err)
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		panic(err)
	}
	return mod
}

// FindDependencyVersion 查找特定依赖项的版本
func FindDependencyVersion(modFile *modfile.File, dependency string) (string, string, bool) {
	for _, req := range modFile.Require {
		if req.Mod.Path == dependency {
			return req.Mod.Path, req.Mod.Version, true
		}
	}
	// 很多场景path是项目内的目录
	for _, req := range modFile.Require {
		if strings.Contains(dependency, req.Mod.Path) {
			return req.Mod.Path, req.Mod.Version, true
		}
	}
	return "", "", false
}

func GetModPath() string {
	modCache := os.Getenv("GOMODCACHE")
	if modCache == "" {
		// 如果环境变量中没有GOMODCACHE，使用 `go env GOMODCACHE` 获取
		cmd := exec.Command("go", "env", "GOMODCACHE")
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		modCache = strings.TrimSpace(string(output))
	}
	return modCache
}

// LongestCommonSubstring 求最长子集,目前没有使用
func LongestCommonSubstring(s1, s2 string) string {
	len1 := len(s1)
	len2 := len(s2)

	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}
	maxLen := 0
	endIndex := len1
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
					endIndex = i
				}
			} else {
				dp[i][j] = 0
			}
		}
	}
	if maxLen == 0 {
		return ""
	}
	return s1[endIndex-maxLen : endIndex]
}
