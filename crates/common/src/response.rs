use serde::{Deserialize, Serialize};

/// Standard API response envelope.
///
/// All external API responses use this structure:
/// ```json
/// { "status": "...", "reason": "...", "next_action_hint": "..." }
/// ```
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct ApiResponse {
    pub status: String,
    pub reason: String,
    pub next_action_hint: String,
}

impl ApiResponse {
    pub fn new(
        status: impl Into<String>,
        reason: impl Into<String>,
        next_action_hint: impl Into<String>,
    ) -> Self {
        Self {
            status: status.into(),
            reason: reason.into(),
            next_action_hint: next_action_hint.into(),
        }
    }

    pub fn ok(reason: impl Into<String>, next_action_hint: impl Into<String>) -> Self {
        Self::new("ok", reason, next_action_hint)
    }

    pub fn error(reason: impl Into<String>, next_action_hint: impl Into<String>) -> Self {
        Self::new("error", reason, next_action_hint)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn new_sets_all_fields() {
        let resp = ApiResponse::new("healthy", "service is running", "proceed with requests");

        assert_eq!(resp.status, "healthy");
        assert_eq!(resp.reason, "service is running");
        assert_eq!(resp.next_action_hint, "proceed with requests");
    }

    #[test]
    fn ok_sets_status_to_ok() {
        let resp = ApiResponse::ok("done", "continue");
        assert_eq!(resp.status, "ok");
    }

    #[test]
    fn error_sets_status_to_error() {
        let resp = ApiResponse::error("failed", "retry later");
        assert_eq!(resp.status, "error");
    }

    #[test]
    fn serializes_to_expected_json() {
        let resp = ApiResponse::new("healthy", "all good", "send requests");
        let json = serde_json::to_value(&resp).unwrap();

        assert_eq!(json["status"], "healthy");
        assert_eq!(json["reason"], "all good");
        assert_eq!(json["next_action_hint"], "send requests");
    }

    #[test]
    fn deserializes_from_json() {
        let json = r#"{"status":"ok","reason":"done","next_action_hint":"proceed"}"#;
        let resp: ApiResponse = serde_json::from_str(json).unwrap();

        assert_eq!(resp.status, "ok");
        assert_eq!(resp.reason, "done");
        assert_eq!(resp.next_action_hint, "proceed");
    }
}
