package sizlib

import "testing"

// Test checks a path both redursive and non-recursive for the sql-cfg*.json files.
// Test verifies that ignoreDirs will skip over a directory with sql-cfg.json in it.
func Test_FindDirsWithSQLCfg(t *testing.T) {

	// func FindDirsWithSQLCfg(pth string, ignoreDirs []string) (rv []string) {

	ignoreDirs := []string{"test1/old"}
	dirs := FindDirsWithSQLCfg("./test1/...", ignoreDirs)

	// fmt.Printf("dirs=%s\n", dirs)

	if len(dirs) == 2 && dirs[0] == "test1" && dirs[1] == "test1/test2" {
	} else {
		t.Errorf("Failed recursive test")
	}

	dirs = FindDirsWithSQLCfg("./test1", ignoreDirs)

	// fmt.Printf("2 dirs=%s\n", dirs)

	if len(dirs) == 1 && dirs[0] == "test1" {
	} else {
		t.Errorf("Failed non-recursive test")
	}
}
