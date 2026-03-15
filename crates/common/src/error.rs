/// Errors that can occur in shared infrastructure.
#[derive(Debug, thiserror::Error)]
pub enum CommonError {
    #[error("json error: {0}")]
    Json(#[from] serde_json::Error),
}
