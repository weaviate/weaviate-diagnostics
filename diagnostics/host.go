package diagnostics

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/pbnjay/memory"
)

type HostInfo struct {
	OperatingSystem string
	Architecture    string
	Cores           uint32
	MemorySizeGB    uint64
	DiskUsage       string
}

func getDiskUse(diskPath string) string {
	fs := syscall.Statfs_t{}

	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return "n/a"
	}

	err := syscall.Statfs(diskPath, &fs)
	if err != nil {
		return "n/a"
	}

	return fmt.Sprintf("Total %d GB, Available: %d GB", fs.Blocks*uint64(fs.Bsize)/1024/1024/1024, fs.Bavail*uint64(fs.Bsize)/1024/1024/1024)
}

func getHostInfo() HostInfo {
	hostInfo := HostInfo{}

	hostInfo.Cores = uint32(runtime.NumCPU())

	hostInfo.MemorySizeGB = memory.TotalMemory() / 1024 / 1024

	hostInfo.OperatingSystem = runtime.GOOS
	hostInfo.Architecture = runtime.GOARCH
	hostInfo.DiskUsage = getDiskUse("/var/lib/weaviate")
	return hostInfo
}
