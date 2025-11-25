import os

# Python environment variable usage
api_key = os.getenv("API_KEY")
db_url = os.environ["DATABASE_URL"]
secret = os.getenv("SECRET_KEY")
port = os.environ.get("PORT", "3000")

print(api_key, db_url, secret, port)

