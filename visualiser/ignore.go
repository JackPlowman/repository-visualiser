package main

// defaultIgnoreList contains a list of default file and folder patterns to ignore.
var defaultIgnoreList = []string{
	".git",         // Git folder
	"node_modules", // Node.js dependencies
	"vendor",       // Vendor folder for Go dependencies
	"*.log",        // Log files
	"*.tmp",        // Temporary files
	"*.swp",        // Swap files
	".DS_Store",    // macOS metadata files
	"*.exe",        // Executable files
	"*.dll",        // Dynamic-link library files
	"*.so",         // Shared object files
	"*.o",          // Object files
	"*.a",          // Archive files
	"*.pyc",        // Compiled Python files
	"*.class",      // Compiled Java files
	"*.jar",        // Java archive files
	"*.war",        // Web application archive files
	"*.zip",        // Zip archives
	"*.tar.gz",     // Tarball archives
	"*.7z",         // 7-Zip archives
	"*.bak",        // Backup files
	"*.old",        // Old files
	"*.orig",       // Original files
}
