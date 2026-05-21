package license

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// machineIDFile 机器ID持久化文件（需从宿主机挂载）
// 容器环境下网卡 MAC、主机名、磁盘序列号都不稳定，必须持久化避免重启后变化
const machineIDFile = "/app/data/.machine_id"

var (
	cachedMachineID string
	machineIDOnce   sync.Once
)

// GetMachineID 获取当前机器的唯一指纹（优先级：环境变量 > 持久化文件 > 实时计算）
// 1. LCA_MACHINE_ID 环境变量（客户可手动锁定）
// 2. /app/data/.machine_id 持久化文件（首次启动生成后写入，后续重启复用）
// 3. 实时采集 MAC+CPU+磁盘+主机名 → SHA256
func GetMachineID() string {
	machineIDOnce.Do(func() {
		cachedMachineID = resolveMachineID()
	})
	return cachedMachineID
}

func resolveMachineID() string {
	// 优先级1：环境变量（运维可控，客户购买后可锁定机器ID防止上漂）
	if id := strings.TrimSpace(os.Getenv("LCA_MACHINE_ID")); id != "" {
		return id
	}

	// 优先级2：读取持久化文件
	if data, err := os.ReadFile(machineIDFile); err == nil {
		if id := strings.TrimSpace(string(data)); id != "" {
			return id
		}
	}

	// 优先级3：实时计算并写入文件
	parts := []string{
		getMACAddress(),
		getCPUSerial(),
		getDiskSerial(),
		getHostname(),
	}
	raw := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(raw))
	id := fmt.Sprintf("%x", hash)

	// 尝试持久化（失败不阻断主流程，仅记录警告）
	if err := persistMachineID(id); err != nil {
		fmt.Printf("[License] 警告：未能持久化机器ID到 %s: %v（容器重启后机器ID可能变化）\n", machineIDFile, err)
	}
	return id
}

// persistMachineID 将机器ID写入持久化文件
func persistMachineID(id string) error {
	dir := filepath.Dir(machineIDFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(machineIDFile, []byte(id), 0644)
}

// getMACAddress 获取首个非loopback网卡的MAC地址
func getMACAddress() string {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getMACUnix()
	case "windows":
		return getMACWindows()
	}
	return "unknown-mac"
}

func getMACUnix() string {
	out, err := exec.Command("sh", "-c", "ifconfig 2>/dev/null | grep -oE '([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}' | head -1").Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	// fallback: try ip link
	out, err = exec.Command("sh", "-c", "ip link 2>/dev/null | grep -oE 'link/ether ([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}' | head -1 | awk '{print $2}'").Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	return "unknown-mac"
}

func getMACWindows() string {
	out, err := exec.Command("cmd", "/c", "wmic nic where PhysicalAdapter=true get MACAddress /format:list 2>nul").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "MACAddress=") {
				return strings.TrimPrefix(line, "MACAddress=")
			}
		}
	}
	return "unknown-mac"
}

// getCPUSerial 获取CPU序列号/标识
func getCPUSerial() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sh", "-c", "sysctl -n machdep.cpu.brand_string 2>/dev/null").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	case "linux":
		out, err := exec.Command("sh", "-c", "cat /proc/cpuinfo 2>/dev/null | grep -m1 'model name' | cut -d: -f2").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
		out, err = exec.Command("sh", "-c", "dmidecode -s processor-id 2>/dev/null | head -1").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	case "windows":
		out, err := exec.Command("cmd", "/c", "wmic cpu get ProcessorId /format:list 2>nul").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "ProcessorId=") {
					return strings.TrimPrefix(line, "ProcessorId=")
				}
			}
		}
	}
	return "unknown-cpu"
}

// getDiskSerial 获取主磁盘序列号
func getDiskSerial() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sh", "-c", "diskutil info disk0 2>/dev/null | grep 'Disk / Partition UUID' | awk '{print $NF}'").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
		out, err = exec.Command("sh", "-c", "system_profiler SPStorageDataType 2>/dev/null | grep 'Serial Number' | head -1 | awk -F': ' '{print $2}'").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	case "linux":
		out, err := exec.Command("sh", "-c", "lsblk -ndo SERIAL $(lsblk -ndo NAME,TYPE /dev/sda 2>/dev/null | head -1 | awk '{print \"/dev/\"$1}') 2>/dev/null | head -1").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
		out, err = exec.Command("sh", "-c", "cat /sys/block/sda/device/serial 2>/dev/null || cat /sys/block/vda/device/serial 2>/dev/null").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	case "windows":
		out, err := exec.Command("cmd", "/c", "wmic diskdrive get SerialNumber /format:list 2>nul").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "SerialNumber=") {
					return strings.TrimPrefix(line, "SerialNumber=")
				}
			}
		}
	}
	return "unknown-disk"
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return hostname
}
