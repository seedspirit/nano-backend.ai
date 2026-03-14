/// Errors that can occur during manager startup or operation.
#[derive(Debug, thiserror::Error)]
pub enum ManagerError {
    #[error("failed to bind server: {0}")]
    Bind(std::io::Error),

    #[error("server error: {0}")]
    Serve(std::io::Error),
}
