package execute

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/darkyzhou/seele/runj/cmd/runj/cgroup"
	"github.com/darkyzhou/seele/runj/cmd/runj/entities"
	"github.com/darkyzhou/seele/runj/cmd/runj/utils"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs2"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/samber/lo"
	"golang.org/x/sys/unix"
)

func initContainerFactory(overlayfsConfig string) (libcontainer.Factory, error) {
	return libcontainer.New(
		".",
		libcontainer.NewuidmapPath("/usr/bin/newuidmap"),
		libcontainer.NewgidmapPath("/usr/bin/newgidmap"),
		libcontainer.InitArgs(os.Args[0], "init", overlayfsConfig),
	)
}

func getCgroupPath(parentCgroupPath string, rootless bool) (string, string, error) {
	var (
		parentPath     string
		fullCgroupPath string
		err            error
	)

	if parentCgroupPath != "" {
		fullCgroupPath, err = cgroup.GetCgroupPathViaFs(parentCgroupPath)
	} else {
		if rootless {
			parentPath, fullCgroupPath, err = cgroup.GetCgroupPathViaSystemd()
		} else {
			fullCgroupPath, err = cgroup.GetCgroupPathViaFs(fs2.UnifiedMountpoint)
		}
	}

	if err != nil {
		return "", "", err
	}

	return parentPath, fullCgroupPath, nil
}

func getIdMappings(config *entities.UserNamespaceConfig) ([]specs.LinuxIDMapping, []specs.LinuxIDMapping) {
	return []specs.LinuxIDMapping{
			{
				HostID:      config.RootUid,
				ContainerID: 0,
				Size:        1,
			},
			{
				HostID:      config.UidMapBegin,
				ContainerID: 1,
				Size:        config.UidMapCount,
			},
		}, []specs.LinuxIDMapping{
			{
				HostID:      config.RootGid,
				ContainerID: 0,
				Size:        1,
			},
			{
				HostID:      config.GidMapBegin,
				ContainerID: 1,
				Size:        config.GidMapCount,
			},
		}
}

func prepareFds(config *entities.FdConfig) (*os.File, *os.File, *os.File, error) {
	if config != nil && config.StdErrToStdOut && config.StdOutToStdErr {
		return nil, nil, nil, fmt.Errorf("Cannot have both StdErrToStdOut and StdOutToStdErr set")
	}

	var err error

	stdInFilePath := lo.TernaryF(
		config == nil || config.StdIn == "",
		func() string {
			return os.DevNull
		},
		func() string {
			return config.StdIn
		},
	)
	stdInFile, err := os.Open(stdInFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error opening the stdin file %s: %w", stdInFilePath, err)
	}

	var (
		stdOutFile *os.File
		stdErrFile *os.File
	)
	if config != nil && config.StdOutToStdErr {
		if config.StdOut != "" {
			return nil, nil, nil, fmt.Errorf("Cannot have both StdOut and StdOutToStdErr set")
		}

		stdErrFile, err = prepareOutFd(false, config)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error preparing the stderr file: %w", err)
		}

		stdOutFile = stdErrFile
	} else if config != nil && config.StdErrToStdOut {
		if config.StdErr != "" {
			return nil, nil, nil, fmt.Errorf("Cannot have both StdErr and StdErrToStdOut set")
		}

		stdOutFile, err = prepareOutFd(true, config)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error preparing the stdout file: %w", err)
		}

		stdErrFile = stdOutFile
	} else {
		stdOutFile, err = prepareOutFd(true, config)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error preparing the stdout file: %w", err)
		}

		stdErrFile, err = prepareOutFd(false, config)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error preparing the stderr file: %w", err)
		}
	}

	return stdInFile, stdOutFile, stdErrFile, nil
}

func prepareOutFd(stdout bool, config *entities.FdConfig) (*os.File, error) {
	if config == nil || (stdout && config.StdOut == "") || (!stdout && config.StdErr == "") {
		mask := unix.Umask(0)
		file, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0o664)
		unix.Umask(mask)
		if err != nil {
			return nil, fmt.Errorf("Error opening the %s: %w", os.DevNull, err)
		}
		return file, nil
	} else {
		path := lo.Ternary(stdout, config.StdOut, config.StdErr)

		modes := os.O_WRONLY | os.O_TRUNC
		if _, err := os.Stat(path); os.IsNotExist(err) {
			modes = modes | os.O_CREATE | os.O_EXCL
		}

		mask := unix.Umask(0)
		file, err := os.OpenFile(path, modes, 0o664)
		unix.Umask(mask)
		if err != nil {
			return nil, fmt.Errorf("Error opening the file: %w", err)
		}
		return file, nil
	}
}

func prepareOverlayfs(userNamespaceConfig *entities.UserNamespaceConfig, config *entities.OverlayfsConfig) (string, error) {
	// FIXME: In seele bare work mode, 'others' bits are not important
	if err := utils.CheckPermission(config.LowerDirectory, 0b000_101_101); err != nil {
		return "", fmt.Errorf("Error checking lower directory's permissions: %w", err)
	}
	if err := utils.CheckPermission(config.UpperDirectory, 0b000_111_111); err != nil {
		return "", fmt.Errorf("Error checking upper directory's permissions: %w", err)
	}
	if err := utils.CheckPermission(config.MergedDirectory, 0b111_000_000); err != nil {
		return "", fmt.Errorf("Error checking merged directory's permissions: %w", err)
	}

	if userNamespaceConfig != nil && userNamespaceConfig.Enabled {
		if err := os.Chown(config.UpperDirectory, int(userNamespaceConfig.RootUid), int(userNamespaceConfig.RootGid)); err != nil {
			return "", fmt.Errorf("Error chowning upper directory: %w", err)
		}
		if err := os.Chown(config.MergedDirectory, int(userNamespaceConfig.RootUid), int(userNamespaceConfig.RootGid)); err != nil {
			return "", fmt.Errorf("Error chowning merged directory: %w", err)
		}
	}

	workdirEmpty, err := utils.DirectoryEmpty(config.WorkDirectory)
	if err != nil {
		return "", fmt.Errorf("Error checking work directory: %w", err)
	}
	if !workdirEmpty {
		return "", fmt.Errorf("The workdir is not empty")
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("Error serializing the config: %w", err)
	}
	return string(data), nil
}

func checkIsOOM(cgroupPath string) (bool, error) {
	memoryEvents, err := cgroups.ReadFile(cgroupPath, "memory.events")
	if err != nil {
		return false, fmt.Errorf("Error reading memory events: %w", err)
	}
	index := strings.LastIndex(memoryEvents, "oom_kill")
	// TODO: should handle the case when index+9 is out of bounds
	return index > 0 && memoryEvents[index+9] != '0', nil
}

func readMemoryPeak(cgroupPath string) (uint64, error) {
	data, err := cgroups.ReadFile(cgroupPath, "memory.peak")
	if err != nil {
		return 0, fmt.Errorf("Error reading memory.peak: %w", err)
	}
	memoryUsage, err := strconv.Atoi(strings.Trim(data, "\n "))
	if err != nil || memoryUsage <= 0 {
		return 0, fmt.Errorf("Unexpected memory.peak value: %s", data)
	}

	return uint64(memoryUsage), nil
}
