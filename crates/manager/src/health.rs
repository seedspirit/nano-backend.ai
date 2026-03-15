use axum::Json;
use common::ApiResponse;

/// Handler for `GET /health`.
///
/// Returns a structured JSON response indicating the service is running.
pub async fn check() -> Json<ApiResponse> {
    tracing::debug!("health check requested");

    Json(ApiResponse::new(
        "healthy",
        "manager is running",
        "proceed with API requests",
    ))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn check_returns_healthy_status() {
        let Json(response) = check().await;

        assert_eq!(response.status, "healthy");
        assert_eq!(response.reason, "manager is running");
        assert_eq!(response.next_action_hint, "proceed with API requests");
    }
}
