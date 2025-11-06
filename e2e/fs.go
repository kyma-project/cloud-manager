package e2e

import (
	"fmt"
	"strings"
)

type FileOperationFunc func(rootDir string) []string

func EchoOperation(message string) FileOperationFunc {
	return func(rootDir string) []string {
		return []string{
			``,
			`# EchoOperation`,
			fmt.Sprintf(`echo "%s"`, message),
			``,
		}
	}
}

func FileExistsOperation(path string) FileOperationFunc {
	path = strings.TrimPrefix(path, "/")
	return func(rootDir string) []string {
		return []string{
			``,
			`# FileExistsOperation`,
			fmt.Sprintf(`[ -e "%s/%s" ] || (echo "File %s does not exist"; exit 1)`, rootDir, path, path),
			``,
		}
	}
}

func FileContainsOperation(path, content string) FileOperationFunc {
	path = strings.TrimPrefix(path, "/")
	return func(rootDir string) []string {
		return []string{
			``,
			`# FileContainsOperation`,
			fmt.Sprintf(`FILE=%s/%s`, rootDir, path),
			fmt.Sprintf(`DIR=%s`, rootDir),
			fmt.Sprintf(`CONTENT="%s"`, content),
			`if ! grep $CONTENT $FILE; then`,
			`  echo "content '$CONTENT' not found"`,
			`  echo "directory '$DIR' content:"`,
			`  ls -la $DIR || true`,
			`  echo "file content:"`,
			`  cat $FILE || true`,
			`  exit 1`,
			`fi`,
			``,
		}
	}
}

func DeleteFileOperation(path string) FileOperationFunc {
	path = strings.TrimPrefix(path, "/")
	return func(rootDir string) []string {
		return []string{
			``,
			`# DeleteFileOperation`,
			fmt.Sprintf(`rm -rf %s/%s`, rootDir, path),
			``,
		}
	}
}

func CreateFileOperation(path, content string) FileOperationFunc {
	path = strings.TrimPrefix(path, "/")
	return func(rootDir string) []string {
		return []string{
			``,
			`# CreateFileOperation`,
			fmt.Sprintf(`echo "%s" > %s/%s`, content, rootDir, path),
			``,
		}
	}
}

func AppendFileOperation(path, content string) FileOperationFunc {
	path = strings.TrimPrefix(path, "/")
	return func(rootDir string) []string {
		return []string{
			``,
			`# CreateFileOperation`,
			fmt.Sprintf(`echo "%s" >> %s/%s`, content, rootDir, path),
			``,
		}
	}
}

func CombineFileOperations(ops ...FileOperationFunc) FileOperationFunc {
	return func(rootDir string) []string {
		var result []string
		for _, op := range ops {
			result = append(result, op(rootDir)...)
		}
		return result
	}
}
