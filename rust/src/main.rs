use actix_web::{web, App, HttpServer};
use std::env;

mod routes;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = env::var("PORT")
        .unwrap()
        .parse()
        .expect("RUST_PORT must be a number");

    let address = format!("0.0.0.0:{}", port);

    println!("Starting server at http://{}", address);

    HttpServer::new(|| {
        App::new()
            .route("/", web::get().to(routes::hello::hello))
            .route("/greet/{name}", web::get().to(routes::greet::greet))
            .route("/status", web::get().to(routes::status::status))
    })
    .bind(&address)?
    .run()
    .await
}