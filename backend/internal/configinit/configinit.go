package configinit

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

func Create(templatePath, targetPath, serverDir string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("target configuration already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	marker := "encryption_key: \"\""
	if strings.Count(string(data), marker) != 1 {
		return fmt.Errorf("template must contain exactly one empty database encryption key")
	}
	key, err := randomKey()
	if err != nil {
		return err
	}
	output := strings.Replace(string(data), marker, "encryption_key: \""+key+"\"", 1)
	if serverDir != "" {
		output = strings.Replace(output, "    dir: \"/opt/cephtower\"", "    dir: \""+serverDir+"\"", 1)
	}
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	name := file.Name()
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(name)
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if _, err := file.WriteString(output); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, targetPath); err != nil {
		return err
	}
	ok = true
	return nil
}
func randomKey() (string, error) {
	result := make([]byte, 32)
	for i := range result {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[value.Int64()]
	}
	return string(result), nil
}
