# Velero Reporter

[![Go Report Card](https://goreportcard.com/badge/github.com/root-ali/velero-reporter)](https://goreportcard.com/report/github.com/root-ali/velero-reporter)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Velero Reporter is a Kubernetes application designed to monitor Velero backup operations and send notifications to Mattermost when backups succeed or fail. It leverages the Kubernetes API, Velero's custom resources, and Mattermost's incoming webhooks to provide real-time feedback on backup status.

## Features

*   **Real-time Backup Monitoring:** Watches Velero Backup resources for changes in status.
*   **Mattermost Notifications:** Sends messages to a specified Mattermost channel when backups complete successfully or fail.
*   **Configurable:** Easily configure Mattermost webhook URL, API key, and other settings via environment variables.
*   **Robust Error Handling:** Gracefully handles errors when interacting with the Kubernetes API, Mattermost, and other components.
*   **Efficient Resource Usage:** Uses Kubernetes informers to efficiently watch for changes without polling.
*   **Easy to deploy:** You can deploy it in your kubernetes cluster.
*   **Easy to configure:** You can configure it with environment variable.

## Architecture

The application consists of the following main components:

*   **Kubernetes Client:** Interacts with the Kubernetes API to watch Velero Backup resources and manage ConfigMaps.
*   **Velero Watcher:** Uses a Kubernetes informer to watch for changes to Velero Backup resources.
*   **Mattermost Client:** Sends messages to Mattermost via incoming webhooks.
*   **ConfigMap Manager:** Manages a ConfigMap to store the last processed resource version.
*   **Health check:** check if the kubernetes api is ready.

## Prerequisites

*   **Kubernetes Cluster:** A running Kubernetes cluster with Velero installed.
*   **Velero:** Velero must be installed and configured in your cluster.
*   **Mattermost:** A Mattermost instance with an incoming webhook configured.
*   **Go:** Go 1.20 or higher.

## Installation and Deployment

1.  **Clone the Repository:**

```bash
 git clone https://github.com/root-ali/velero-reporter.git && cd velero-reporter
```

2.  **Build the Application:**
```bash
docker image build -t velero-reporter .
```

3.  **Customize `values.yaml`:**


*  Edit the `values.yaml` file to configure the following:
    * `image.repository`: Set this to your Docker image repository (e.g., `your-docker-registry/velero-reporter`).
    * `image.tag`: Set this to the tag of your Docker image (e.g., `latest`).
    * `env.mattermostUrl`: Set this to your Mattermost URL.
    * `env.mattermostToken`: Set this to your Mattermost webhook token.
    * `env.kubeConfigType`: Set this to `in-cluster` or `out-of-cluster`.
    * `env.kubeConfigPath`: Set this to the path of your kubeconfig file (if `out-of-cluster`).
    * `env.httpHost`: Set this to the host for the http server.
    * `env.httpPort`: Set this to the port for the http server.
    * `env.logLevel`: Set this to the log level.
    * You can also customize other settings in `values.yaml` as needed.

4.  **Install the Chart:**

```bash
helm install my-velero-reporter . -n velero --create-namespace
```

 *   Replace `my-velero-reporter` with the desired release name.
    *   Replace `.` with the path to your chart directory if you are not in it.
    * `--create-namespace` will create the namespace if it does not exist.

5.  **Uninstall the Chart:**

```bash
helm uninstall my-velero-reporter
```

## Configuration

The following environment variables can be used to configure Velero Reporter:

*   **`VELERO_REPORTER_MATTERMOST_URL`:** (Required) The base URL of your Mattermost instance (e.g., `https://mattermost.example.com`).
*   **`VELERO_REPORTER_MATTERMOST_TOKEN`:** (Required) The API key (webhook token) for the Mattermost incoming webhook.
*   **`VELERO_REPORTER_KUBECONFIG_TYPE`:** (Optional) Set to `in-cluster` (default) if running inside the cluster or `out-of-cluster` if running outside.
*   **`VELERO_REPORTER_KUBECONFIG_PATH`:** (Optional) The path to your kubeconfig file if `VELERO_REPORTER_KUBECONFIG_TYPE` is `out-of-cluster`.
*   **`VELERO_REPORTER_HTTP_HOST`:** (Optional) The host for the http server.
*   **`VELERO_REPORTER_HTTP_PORT`:** (Optional) The port for the http server.
*   **`LOG_LEVEL`:** (Optional) The logging level (`info` or `debug`). Default is `info`.


## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License.