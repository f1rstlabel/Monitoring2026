package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	argStr := strings.Join(os.Args, " ")
	f, _ := os.OpenFile("d:/Project/backup/vue project/bin/git_args.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(argStr + "\n")
		f.Close()
	}

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("git version 2.45.0.windows.1")
		return
	}

	dummyPkg := `{"name":"git-dummy","version":"1.0.0","main":"index.js"}`

	for i, arg := range os.Args {
		if arg == "clone" {
			var targetDir string
			for j := i + 1; j < len(os.Args); j++ {
				if !strings.HasPrefix(os.Args[j], "-") {
					targetDir = os.Args[j]
				}
			}
			if targetDir != "" {
				parentDir := filepath.Dir(targetDir)
				_ = os.MkdirAll(targetDir, 0755)
				_ = os.MkdirAll(parentDir, 0755)
				_ = os.WriteFile(filepath.Join(targetDir, "package.json"), []byte(dummyPkg), 0644)
				_ = os.WriteFile(filepath.Join(parentDir, "package.json"), []byte(dummyPkg), 0644)
				_ = os.WriteFile(filepath.Join(targetDir, "index.js"), []byte("module.exports = {};"), 0644)
				_ = os.WriteFile(filepath.Join(parentDir, "index.js"), []byte("module.exports = {};"), 0644)
				fmt.Printf("Cloned into '%s'...\n", targetDir)
			}
			return
		}
	}

	fmt.Println("git stub ok")
}
