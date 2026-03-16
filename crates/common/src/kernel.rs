use serde::{Deserialize, Serialize};
use std::fmt;

// ── Types ──────────────────────────────────────────────

/// Unique identifier for a kernel instance.
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct KernelId(pub String);

impl fmt::Display for KernelId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Specification for creating a new kernel.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KernelSpec {
    pub command: Vec<String>,
}

/// Current status of a kernel.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum KernelStatus {
    Running,
    Exited { code: i32 },
    Failed { reason: String },
}

// ── Error ──────────────────────────────────────────────

/// Errors from kernel operations.
#[derive(Debug, thiserror::Error)]
pub enum KernelError {
    #[error("kernel not found: {0}")]
    NotFound(KernelId),

    #[error("kernel already exists: {0}")]
    AlreadyExists(KernelId),

    #[error("runtime error: {0}")]
    Runtime(String),
}

// ── Trait ───────────────────────────────────────────────

/// Abstraction over kernel lifecycle management.
///
/// Implementations handle the actual creation, destruction, and status
/// querying of kernel processes (e.g., local process, Docker, K8s).
pub trait KernelRuntime {
    fn create(
        &self,
        spec: KernelSpec,
    ) -> impl std::future::Future<Output = Result<KernelId, KernelError>> + Send;

    fn destroy(
        &self,
        id: KernelId,
    ) -> impl std::future::Future<Output = Result<(), KernelError>> + Send;

    fn status(
        &self,
        id: KernelId,
    ) -> impl std::future::Future<Output = Result<KernelStatus, KernelError>> + Send;
}

// ── Tests ──────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn kernel_id_equality() {
        let a = KernelId("k-001".to_string());
        let b = KernelId("k-001".to_string());
        let c = KernelId("k-002".to_string());

        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn kernel_id_display() {
        let id = KernelId("k-123".to_string());
        assert_eq!(id.to_string(), "k-123");
    }

    #[test]
    fn kernel_spec_serializes() {
        let spec = KernelSpec {
            command: vec![
                "python".to_string(),
                "-c".to_string(),
                "print('hi')".to_string(),
            ],
        };
        let json = serde_json::to_value(&spec).unwrap();

        assert_eq!(
            json["command"],
            serde_json::json!(["python", "-c", "print('hi')"])
        );
    }

    #[test]
    fn kernel_spec_deserializes() {
        let json = r#"{"command":["sleep","10"]}"#;
        let spec: KernelSpec = serde_json::from_str(json).unwrap();

        assert_eq!(spec.command, vec!["sleep", "10"]);
    }

    #[test]
    fn kernel_status_running_serializes() {
        let status = KernelStatus::Running;
        let json = serde_json::to_string(&status).unwrap();

        assert!(json.contains("Running"));
    }

    #[test]
    fn kernel_status_exited_roundtrip() {
        let status = KernelStatus::Exited { code: 0 };
        let json = serde_json::to_string(&status).unwrap();
        let restored: KernelStatus = serde_json::from_str(&json).unwrap();

        assert_eq!(status, restored);
    }

    #[test]
    fn kernel_status_failed_roundtrip() {
        let status = KernelStatus::Failed {
            reason: "out of memory".to_string(),
        };
        let json = serde_json::to_string(&status).unwrap();
        let restored: KernelStatus = serde_json::from_str(&json).unwrap();

        assert_eq!(status, restored);
    }

    #[test]
    fn kernel_error_not_found_display() {
        let err = KernelError::NotFound(KernelId("k-999".to_string()));
        assert_eq!(err.to_string(), "kernel not found: k-999");
    }

    #[test]
    fn kernel_error_already_exists_display() {
        let err = KernelError::AlreadyExists(KernelId("k-001".to_string()));
        assert_eq!(err.to_string(), "kernel already exists: k-001");
    }

    #[test]
    fn kernel_error_runtime_display() {
        let err = KernelError::Runtime("connection refused".to_string());
        assert_eq!(err.to_string(), "runtime error: connection refused");
    }
}
