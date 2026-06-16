package atomicfile

import (
"os"
"path/filepath"
"testing"
)

func TestWriteAtomically(t *testing.T) {
tmpDir := t.TempDir()
target := filepath.Join(tmpDir, "test-file.txt")
data := []byte("dados criticos do batch processor")

if err := WriteAtomically(target, data); err != nil {
t.Fatalf("WriteAtomically falhou: %v", err)
}

read, err := os.ReadFile(target)
if err != nil {
t.Fatalf("ler arquivo: %v", err)
}
if string(read) != string(data) {
t.Errorf("conteúdo diferente: got %s, want %s", read, data)
}
}

func TestReadIfExists_NotExists(t *testing.T) {
tmpDir := t.TempDir()
path := filepath.Join(tmpDir, "nao-existe.txt")

data, err := ReadIfExists(path)
if err != nil {
t.Fatalf("ReadIfExists retornou erro: %v", err)
}
if data != nil {
t.Errorf("esperado nil, got %v", data)
}
}

func TestWriteAtomically_RaceCondition(t *testing.T) {
tmpDir := t.TempDir()
target := filepath.Join(tmpDir, "race-file.txt")

done := make(chan bool, 10)
for i := 0; i < 10; i++ {
go func(id int) {
data := []byte(string(rune('0' + id)))
WriteAtomically(target, data)
done <- true
}(i)
}

for i := 0; i < 10; i++ {
<-done
}

read, err := os.ReadFile(target)
if err != nil {
t.Fatalf("ler arquivo: %v", err)
}
if len(read) != 1 {
t.Errorf("arquivo corrompido: %v", read)
}
}
