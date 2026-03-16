use serde::{Deserialize, Serialize};
use std::fmt;
use uuid::Uuid;

// ── Types ──────────────────────────────────────────────

/// Unique identifier for a kernel instance.
///
/// Wraps a UUID v4 to guarantee uniqueness and prevent construction
/// from arbitrary strings.
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct KernelID(Uuid);

impl KernelID {
    /// Generate a new random kernel identifier.
    pub fn new() -> Self {
        Self(Uuid::new_v4())
    }

    /// Reconstruct a `KernelID` from an existing UUID.
    pub fn from_uuid(uuid: Uuid) -> Self {
        Self(uuid)
    }

    /// Return the inner UUID.
    pub fn as_uuid(&self) -> &Uuid {
        &self.0
    }
}

impl Default for KernelID {
    fn default() -> Self {
        Self::new()
    }
}

impl fmt::Display for KernelID {
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
    NotFound(KernelID),

    #[error("kernel already exists: {0}")]
    AlreadyExists(KernelID),

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
    ) -> impl std::future::Future<Output = Result<KernelID, KernelError>> + Send;

    fn destroy(
        &self,
        id: KernelID,
    ) -> impl std::future::Future<Output = Result<(), KernelError>> + Send;

    fn status(
        &self,
        id: KernelID,
    ) -> impl std::future::Future<Output = Result<KernelStatus, KernelError>> + Send;
}

// ── Tests ──────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn kernel_id_new_is_unique() {
        let a = KernelID::new();
        let b = KernelID::new();
        assert_ne!(a, b, "two new KernelIDs should differ");
    }

    #[test]
    fn kernel_id_equality_from_same_uuid() {
        let uuid = Uuid::new_v4();
        let a = KernelID::from_uuid(uuid);
        let b = KernelID::from_uuid(uuid);
        assert_eq!(a, b);
    }

    #[test]
    fn kernel_id_display_matches_uuid() {
        let uuid = Uuid::new_v4();
        let id = KernelID::from_uuid(uuid);
        assert_eq!(id.to_string(), uuid.to_string());
    }

    #[test]
    fn kernel_id_serialization_roundtrip() {
        let id = KernelID::new();
        let json = serde_json::to_string(&id).expect("KernelID serialization should succeed");
        let restored: KernelID =
            serde_json::from_str(&json).expect("KernelID deserialization should succeed");
        assert_eq!(id, restored);
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
        let json = serde_json::to_value(&spec).expect("KernelSpec serialization should succeed");

        assert_eq!(
            json["command"],
            serde_json::json!(["python", "-c", "print('hi')"])
        );
    }

    #[test]
    fn kernel_spec_deserializes() {
        let json = r#"{"command":["sleep","10"]}"#;
        let spec: KernelSpec =
            serde_json::from_str(json).expect("KernelSpec deserialization should succeed");

        assert_eq!(spec.command, vec!["sleep", "10"]);
    }

    #[test]
    fn kernel_status_running_serializes() {
        let status = KernelStatus::Running;
        let json = serde_json::to_string(&status)
            .expect("KernelStatus::Running serialization should succeed");

        assert!(json.contains("Running"));
    }

    #[test]
    fn kernel_status_exited_roundtrip() {
        let status = KernelStatus::Exited { code: 0 };
        let json = serde_json::to_string(&status)
            .expect("KernelStatus::Exited serialization should succeed");
        let restored: KernelStatus = serde_json::from_str(&json)
            .expect("KernelStatus::Exited deserialization should succeed");

        assert_eq!(status, restored);
    }

    #[test]
    fn kernel_status_failed_roundtrip() {
        let status = KernelStatus::Failed {
            reason: "out of memory".to_string(),
        };
        let json = serde_json::to_string(&status)
            .expect("KernelStatus::Failed serialization should succeed");
        let restored: KernelStatus = serde_json::from_str(&json)
            .expect("KernelStatus::Failed deserialization should succeed");

        assert_eq!(status, restored);
    }

    #[test]
    fn kernel_error_not_found_display() {
        let id = KernelID::new();
        let expected = format!("kernel not found: {id}");
        let err = KernelError::NotFound(id);
        assert_eq!(err.to_string(), expected);
    }

    #[test]
    fn kernel_error_already_exists_display() {
        let id = KernelID::new();
        let expected = format!("kernel already exists: {id}");
        let err = KernelError::AlreadyExists(id);
        assert_eq!(err.to_string(), expected);
    }

    #[test]
    fn kernel_error_runtime_display() {
        let err = KernelError::Runtime("connection refused".to_string());
        assert_eq!(err.to_string(), "runtime error: connection refused");
    }
}
