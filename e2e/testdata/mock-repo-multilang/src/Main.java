public class Main {
    public static void main(String[] args) {
        // Java environment variable usage
        String apiKey = System.getenv("API_KEY");
        String dbUrl = System.getenv().get("DATABASE_URL");
        String secret = System.getenv("SECRET_KEY");
        String port = System.getenv("PORT");
        
        System.out.println(apiKey + " " + dbUrl + " " + secret + " " + port);
    }
}

