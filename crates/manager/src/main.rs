mod app;
mod error;
mod health;

use std::net::SocketAddr;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

#[tokio::main]
async fn main() -> Result<(), error::ManagerError> {
    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let app = app::build_router();

    // Port 8080: default for the manager HTTP API.
    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    tracing::info!(%addr, "starting manager");

    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .map_err(error::ManagerError::Bind)?;

    axum::serve(listener, app)
        .await
        .map_err(error::ManagerError::Serve)?;

    Ok(())
}
