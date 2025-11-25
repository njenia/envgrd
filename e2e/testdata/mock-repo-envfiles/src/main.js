// JavaScript using env vars from multiple file types
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const logLevel = process.env.LOG_LEVEL;
const prodOnly = process.env.PROD_ONLY_VAR;
const localOnly = process.env.LOCAL_ONLY_VAR;

// From .envrc
const envrcVar = process.env.ENVRC_VAR;
const envrcApiKey = process.env.ENVRC_API_KEY;

// From docker-compose.yml
const dockerDbHost = process.env.DOCKER_DB_HOST;
const dockerDbPort = process.env.DOCKER_DB_PORT;
const workerQueue = process.env.WORKER_QUEUE;

// From Kubernetes ConfigMap
const k8sApiUrl = process.env.K8S_API_URL;
const k8sLogLevel = process.env.K8S_LOG_LEVEL;

// From Kubernetes Secret
const k8sSecretKey = process.env.K8S_SECRET_KEY;
const k8sDbPassword = process.env.K8S_DB_PASSWORD;

// From systemd service file
const systemdPort = process.env.SYSTEMD_PORT;
const systemdDbUrl = process.env.SYSTEMD_DB_URL;

// From shell script
const shellVar = process.env.SHELL_SCRIPT_VAR;
const shellApiKey = process.env.SHELL_API_KEY;

console.log(apiKey, dbUrl, logLevel, prodOnly, localOnly, envrcVar, dockerDbHost, k8sApiUrl, systemdPort, shellVar);

