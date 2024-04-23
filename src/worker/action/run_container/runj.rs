use std::{fmt::Display, path::PathBuf};

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RunjConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub user_namespace: Option<UserNamespaceConfig>,

    pub overlayfs: OverlayfsConfig,

    pub cgroup_path: PathBuf,

    pub cwd: PathBuf,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub paths: Option<Vec<PathBuf>>,

    pub command: Vec<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub fd: Option<FdConfig>,

    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub mounts: Vec<MountConfig>,

    pub limits: LimitsConfig,

    pub no_new_keyring: bool,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct UserNamespaceConfig {
    pub enabled: bool,
    pub root_uid: u32,
    pub uid_map_begin: u32,
    pub uid_map_count: u32,
    pub root_gid: u32,
    pub gid_map_begin: u32,
    pub gid_map_count: u32,
    pub container_uid: u32,
    pub container_gid: u32,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct OverlayfsConfig {
    pub lower_dir: PathBuf,
    pub upper_dir: PathBuf,
    pub work_dir: PathBuf,
    pub merged_dir: PathBuf,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct FdConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub stdin: Option<PathBuf>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub stdout: Option<PathBuf>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub stderr: Option<PathBuf>,

    #[serde(default)]
    pub stdout_to_stderr: bool,

    #[serde(default)]
    pub stderr_to_stdout: bool,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct MountConfig {
    pub from: PathBuf,
    pub to: PathBuf,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub options: Option<Vec<String>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct LimitsConfig {
    pub time_ms: u64,
    pub cgroup: CgroupConfig,
    pub rlimit: RlimitConfig,
}

#[derive(Debug, Default, Clone, Deserialize, Serialize)]
pub struct CgroupConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpu_shares: Option<u64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpu_quota: Option<i64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpuset_cpus: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub cpuset_mems: Option<String>,

    pub memory: i64,

    pub pids_limit: i64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RlimitConfig {
    pub core: RlimitItem,

    pub fsize: RlimitItem,

    pub no_file: RlimitItem,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RlimitItem {
    hard: u64,
    soft: u64,
}

impl RlimitItem {
    #[inline]
    pub fn new_single(value: u64) -> Self {
        Self { hard: value, soft: value }
    }
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ContainerExecutionReport {
    pub status: ContainerExecutionStatus,
    pub exit_code: i64,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub signal: Option<String>,

    pub wall_time_ms: u64,
    pub cpu_user_time_ms: u64,
    pub cpu_kernel_time_ms: u64,
    pub memory_usage_kib: u64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "SCREAMING_SNAKE_CASE")]
pub enum ContainerExecutionStatus {
    Normal,
    RuntimeError,
    SignalTerminate,
    UserTimeLimitExceeded,
    WallTimeLimitExceeded,
    MemoryLimitExceeded,
    OutputLimitExceeded,
    Unknown,
}

impl Display for ContainerExecutionStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "{}",
            match self {
                Self::Normal => "NORMAL",
                Self::RuntimeError => "RUNTIME_ERROR",
                Self::SignalTerminate => "SIGNAL_TERMINATE",
                Self::UserTimeLimitExceeded => "USER_TIME_LIMIT_EXCEEDED",
                Self::WallTimeLimitExceeded => "WALL_TIME_LIMIT_EXCEEDED",
                Self::MemoryLimitExceeded => "MEMORY_LIMIT_EXCEEDED",
                Self::OutputLimitExceeded => "OUTPUT_LIMIT_EXCEEDED",
                Self::Unknown => "UNKNOWN",
            }
        )
    }
}
