package entities

type RunjConfig struct {
	UserNamespace *UserNamespaceConfig `mapstructure:"user_namespace"`
	Overlayfs     *OverlayfsConfig     `mapstructure:"overlayfs" validate:"required"`
	CgroupPath    string               `mapstructure:"cgroup_path"`
	Cwd           string               `mapstructure:"cwd" validate:"required"`
	Command       []string             `mapstructure:"command" validate:"required,dive,required"`
	Paths         []string             `mapstructure:"paths" validate:"dive,required"`
	Fd            *FdConfig            `mapstructure:"fd"`
	Mounts        []*MountConfig       `mapstructure:"mounts"`
	Limits        *LimitsConfig        `mapstructure:"limits" validate:"required"`
	NoNewKeyring  bool                 `mapstructure:"no_new_keyring" validate:"required"`
}

type UserNamespaceConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	RootUid      uint32 `mapstructure:"root_uid" validate:"required"`
	UidMapBegin  uint32 `mapstructure:"uid_map_begin" validate:"required"`
	UidMapCount  uint32 `mapstructure:"uid_map_count" validate:"required"`
	RootGid      uint32 `mapstructure:"root_gid" validate:"required"`
	GidMapBegin  uint32 `mapstructure:"gid_map_begin" validate:"required"`
	GidMapCount  uint32 `mapstructure:"gid_map_count" validate:"required"`
	ContainerUID uint32 `mapstructure:"container_uid" validate:"required"`
	ContainerGID uint32 `mapstructure:"container_gid" validate:"required"`
}

type OverlayfsConfig struct {
	LowerDirectory  string `mapstructure:"lower_dir" validate:"required"`
	UpperDirectory  string `mapstructure:"upper_dir" validate:"required"`
	WorkDirectory   string `mapstructure:"work_dir" validate:"required"`
	MergedDirectory string `mapstructure:"merged_dir" validate:"required"`
}

type FdConfig struct {
	StdIn          string `mapstructure:"stdin"`
	StdOut         string `mapstructure:"stdout"`
	StdErr         string `mapstructure:"stderr"`
	StdOutToStdErr bool   `mapstructure:"stdout_to_stderr"`
	StdErrToStdOut bool   `mapstructure:"stderr_to_stdout"`
}

type MountConfig struct {
	From    string   `mapstructure:"from" validate:"required"`
	To      string   `mapstructure:"to" validate:"required"`
	Options []string `mapstructure:"options"`
}

type LimitsConfig struct {
	TimeMs uint64        `mapstructure:"time_ms" validate:"required"`
	Cgroup *CgroupConfig `mapstructure:"cgroup" validate:"required"`
	Rlimit *RlimitConfig `mapstructure:"rlimit" validate:"required"`
}

type CgroupConfig struct {
	CpuShares  uint64 `mapstructure:"cpu_shares"`
	CpuQuota   int64  `mapstructure:"cpu_quota"`
	CpusetCpus string `mapstructure:"cpuset_cpus"`
	CpusetMems string `mapstructure:"cpuset_mems"`
	Memory     int64  `mapstructure:"memory" validate:"required"`
	PidsLimit  int64  `mapstructure:"pids_limit" validate:"required"`
}

type RlimitConfig struct {
	Core   *RlimitItem `mapstructure:"core" validate:"required"`
	Fsize  *RlimitItem `mapstructure:"fsize" validate:"required"`
	NoFile *RlimitItem `mapstructure:"no_file" validate:"required"`
}

type RlimitItem struct {
	Hard uint64 `mapstructure:"hard"`
	Soft uint64 `mapstructure:"soft"`
}
