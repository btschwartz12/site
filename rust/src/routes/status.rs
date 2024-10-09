use actix_web::{HttpResponse, Responder};
use serde::Serialize;

#[derive(Serialize)]
struct Status {
    status: String,
    uptime: u64, // In seconds
}

pub async fn status() -> impl Responder {
    let server_status = Status {
        status: "Running".to_string(),
        uptime: 3600, // Example: 1 hour
    };

    HttpResponse::Ok().json(server_status)
}
