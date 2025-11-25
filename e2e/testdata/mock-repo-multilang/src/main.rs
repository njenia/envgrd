use std::env;

fn main() {
    // Rust environment variable usage
    let api_key = env::var("API_KEY").unwrap_or_default();
    let db_url = std::env::var("DATABASE_URL").unwrap_or_default();
    let secret = env::var_os("SECRET_KEY");
    let port = std::env::var_os("PORT");
    
    println!("{} {} {:?} {:?}", api_key, db_url, secret, port);
}

