// TypeScript environment variable usage
const apiKey = process.env.API_KEY;
const dbUrl = process.env["DATABASE_URL"];
const secret = process.env.SECRET_KEY;
const port = process.env.PORT;

console.log(apiKey, dbUrl, secret, port);

