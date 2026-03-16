/// Errors that can occur during agent startup or operation.
#[derive(Debug, thiserror::Error)]
#[allow(dead_code)]
pub enum AgentError {
    #[error("failed to bind server: {0}")]
    Bind(std::io::Error),

    #[error("server error: {0}")]
    Serve(std::io::Error),
}
