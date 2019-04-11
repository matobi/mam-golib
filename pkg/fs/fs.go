package fs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"crypto/md5"
	"encoding/hex"
	"io"
	"path/filepath"
	"sort"
)

// ScanDir returns content of dir. Returns empty list if dir does not exist.
func ScanDir(dirname string) []os.FileInfo {
	empty := []os.FileInfo{}
	if !PathExists(dirname) {
		return empty
	}
	dir, err := os.Open(dirname)
	if err != nil {
		return empty
	}
	list, err := dir.Readdir(-1)
	dir.Close()
	if err != nil {
		return empty
	}
	return list
}

// FindFiles Returns all files/dirs in a dir matching given regexp.
func FindFiles(dir string, r *regexp.Regexp) ([]os.FileInfo, error) {
	var result []os.FileInfo
	for _, f := range ScanDir(dir) {
		if r.MatchString(f.Name()) {
			result = append(result, f)
		}
	}
	return result, nil
}

// FindSuffix Returns all files in dir with given suffix.
func FindSuffix(dir, suffix string) []string {
	selection := []string{}
	for _, f := range ScanDir(dir) {
		if f.IsDir() || suffix != Suffix(f.Name()) {
			continue
		}
		selection = append(selection, path.Join(dir, f.Name()))
	}
	sort.Strings(selection)
	return selection
}

// DirSize Calcs sum of all fiels in dir recursively.
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// Suffix Get the suffix of a file name, Suffix is defined as everything
// after the first dot, not including the dot.
// Note the standard go file functions defines suffix as everything
// after the last dot, including the dot.
func Suffix(filename string) string {
	name := path.Base(filename) // Make sure we only have filename, not full path

	// Many files in aspera_test have suffixes like "mxf.xml" or "mov.aspx" or similar.
	// We need to check the full siffix, not only the last part.
	index := strings.Index(name, ".")
	if index < 0 || index == len(name)-1 {
		return ""
	}
	return strings.ToLower(name[index+1:])
}

// WithoutSuffix Returns file name or path with suffix removed.
func WithoutSuffix(filename string) string {
	suffix := Suffix(filename)
	if suffix == "" {
		return filename
	}
	return filename[0 : len(filename)-len(suffix)-1]
}

// PathExists Check if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir Returns true if path exist and is a dir.
func IsDir(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.IsDir()
}

// IsFile Returns true if path exist and is a file.
func IsFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}

// CreateDirIfNotExist Create dir if not exists.
func CreateDirIfNotExist(dir string) error {
	if PathExists(dir) {
		return nil
	}
	return os.MkdirAll(dir, os.ModePerm)
}

// RemoveDirIfEmpty Deletes a dir if it is empty.
func RemoveDirIfEmpty(dir string) error {
	if !PathExists(dir) {
		return nil
	}
	files := ScanDir(dir)
	if len(files) > 0 {
		return nil // not empty
	}
	return os.Remove(dir)
}

// RemoveFile Removs a file. Ignor if not exists or is a dir.
func RemoveFile(path string) {
	f, err := os.Stat(path)
	if err != nil {
		return // file probably not exist.
	}
	if f.IsDir() {
		log.Printf("RemoveFile called for dir; path=%s\n", path)
		return // Don't delete dir
	}
	if err := os.Remove(path); err != nil {
		log.Printf("RemoveFile failed; path=%s; err=%+v\n", path, err)
	}
}

// MovePath Moves a path to new destination.
func MovePath(src string, dest string) error {
	if !PathExists(src) {
		return fmt.Errorf("MovePath; source missing; %s", src)
	}
	destDir := path.Dir(dest)
	CreateDirIfNotExist(destDir)
	return RunCmd("mv", []string{src, dest})
}

// MoveDir moves a subdir to another root.
func MoveDir(srcRoot string, destRoot string, subdir string) error {
	src := path.Join(srcRoot, subdir)
	dest := path.Join(destRoot, subdir)

	// If source dir is empty, we remove it, but do not delete dest dir.
	if err := RemoveDirIfEmpty(src); err != nil {
		return err
	}
	if !PathExists(src) {
		return nil
	}

	// Source dir exists and is not empty.
	// Clear dest dir and move source dir.
	if err := RemoveAll(dest, 4); err != nil {
		return err
	}
	if err := RunCmd("mv", []string{src, dest}); err != nil {
		return err
	}
	return nil
}

func CopyFile(src string, dest string) error {
	if !IsFile(src) {
		return fmt.Errorf("CopyFile; source missing; %s", src)
	}
	if PathExists(dest) {
		return fmt.Errorf("CopyFile; dest already exist; %s", dest)
	}
	destDir := path.Dir(dest)
	if !IsDir(destDir) {
		return fmt.Errorf("CopyFile; dest dir missing; %s", destDir)
	}
	return RunCmd("cp", []string{src, dest})
}

func MoveFile(src string, dest string) error {
	if !IsFile(src) {
		return fmt.Errorf("MoveFile; source missing; %s", src)
	}
	if PathExists(dest) {
		return fmt.Errorf("MoveFile; dest already exist; %s", dest)
	}
	destDir := path.Dir(dest)
	if !IsDir(destDir) {
		return fmt.Errorf("MoveFile; dest dir missing; %s", destDir)
	}
	return RunCmd("mv", []string{src, dest})
}

// LoadJSON Read json from file.
func LoadJSON(f string, iface interface{}) error {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, iface)
}

// SaveJSON Writes a json stuct to file.
func SaveJSON(f string, iface interface{}) error {
	if !PathExists(filepath.Dir(f)) {
		return fmt.Errorf("Missing dir; %s", f)
	}

	content, err := json.MarshalIndent(iface, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f, content, 0644)
}

// SaveXML Writes a json stuct to file.
func SaveXML(f string, iface interface{}) error {
	if !PathExists(filepath.Dir(f)) {
		return fmt.Errorf("Missing dir; %s", f)
	}

	content, err := xml.MarshalIndent(iface, "", " ")
	if err != nil {
		return err
	}
	content = []byte(xml.Header + string(content))
	return ioutil.WriteFile(f, content, 0644)
}

// LoadJSON Read json from file.
func LoadXML(f string, iface interface{}) error {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	return xml.Unmarshal(raw, iface)
}

// ReadLines Read all lines from a file
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines, scanner.Err()
}

// GetHouseDir Get full path to house dir from given root dir.
func GetHouseDir(root string, houseID string) (string, error) {
	length := len(houseID)
	if length < 6 || length > 16 || length%2 != 0 {
		return "", fmt.Errorf("bad houseID; %s", houseID)
	}

	var buf bytes.Buffer
	x := "xxxxxxxxxxxxxxxx"
	for i := 2; i <= length; i += 2 {
		buf.WriteString("/")
		buf.WriteString(houseID[0:i])
		buf.WriteString(x[0 : length-i])
	}
	houseDir := path.Clean(path.Join(root, buf.String()))
	return houseDir, nil
}

// RemoveAll remove dir recursively. Ignored if level of subdirs > maxDepth.
func RemoveAll(root string, maxDepth int) error {
	fi, err := os.Stat(root)
	if err != nil {
		return nil // Ignore if path does not exist.
	}
	if !fi.IsDir() {
		os.Remove(root) // return single file
		return nil
	}

	// Verify removed dir is not too deep. To make sure we are not trying to remove wrong path.
	if maxDepth >= 0 {
		rootDepth := len(strings.Split(root, "/"))
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			depth := len(strings.Split(path, "/")) - rootDepth
			if depth > maxDepth {
				return fmt.Errorf("removeAll too deep; %d; %s; %s", depth, root, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return os.RemoveAll(root)
}

// SetDefaultPermissions set file permissions to default user/group recursively.
func SetDefaultPermissions(path string) error {
	if err := ChmodR(path, 0775); err != nil {
		return fmt.Errorf("failed to chmod: %s", path)
	}
	if err := ChownR(path, 2000, 2000); err != nil {
		return fmt.Errorf("failed to chown: %s", path)
	}
	return nil
}

// ChmodR set file mod recursively.
func ChmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(path, mode)
		}
		return err
	})
}

// ChownR set file uid/gid recursively.
func ChownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

// RunCmd Run command and use current proc stdout/stderr.
func RunCmd(cmdName string, cmdArgs []string) error {
	cmdLog := fmt.Sprintf("%s %s", cmdName, strings.Join(cmdArgs, " "))
	//log.Info().Str("cmd", cmdLog).Msg("will execute command")

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed execute command; cmd=%s; err=%+v", cmdLog, err)
	}
	return nil
}

// RunCmdWithOutput Run command and return stdout.
func RunCmdWithOutput(cmdName string, cmdArgs []string) ([]byte, error) {
	cmdLog := fmt.Sprintf("%s %s", cmdName, strings.Join(cmdArgs, " "))
	//log.Info().Str("cmd", cmdLog).Msg("will execute command")

	out, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return []byte(""), fmt.Errorf("failed execute command; cmd=%s; err=%+v", cmdLog, err)
	}
	return out, nil
}

// HashMd5 calc checksum for file.
func HashMd5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	value := hex.EncodeToString(hash.Sum(nil))
	return value, nil
}
