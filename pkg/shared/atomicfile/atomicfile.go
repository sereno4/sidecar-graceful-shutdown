package atomicfile

import (
"fmt"
"os"
"path/filepath"
)

func WriteAtomically(path string, data []byte) error {
dir := filepath.Dir(path)
tmpFile, err := os.CreateTemp(dir, ".tmp-*")
if err != nil {
return fmt.Errorf("criar temp file: %w", err)
}
tmpPath := tmpFile.Name()

cleaned := false
defer func() {
if !cleaned {
os.Remove(tmpPath)
}
}()

if _, err := tmpFile.Write(data); err != nil {
tmpFile.Close()
return fmt.Errorf("escrever temp file: %w", err)
}

if err := tmpFile.Sync(); err != nil {
tmpFile.Close()
return fmt.Errorf("fsync temp file: %w", err)
}

if err := tmpFile.Close(); err != nil {
return fmt.Errorf("fechar temp file: %w", err)
}

if err := os.Rename(tmpPath, path); err != nil {
return fmt.Errorf("rename atômico: %w", err)
}
cleaned = true

dirFile, err := os.Open(dir)
if err != nil {
return fmt.Errorf("abrir diretório: %w", err)
}
defer dirFile.Close()
if err := dirFile.Sync(); err != nil {
return fmt.Errorf("fsync diretório: %w", err)
}

return nil
}

func ReadIfExists(path string) ([]byte, error) {
data, err := os.ReadFile(path)
if err != nil {
if os.IsNotExist(err) {
return nil, nil
}
return nil, fmt.Errorf("ler arquivo: %w", err)
}
return data, nil
}

func Exists(path string) bool {
_, err := os.Stat(path)
return err == nil
}
