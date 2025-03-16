package errors

import "errors"

// Kubernetes errors
var KUBERNETES_HEALTH_ERROR = errors.New("Cannot check kubernetes health api")
var KUBERNETES_API_NOT_READY = errors.New("Kubernetes API not ready")
var KUBERNETES_API_ERROR = errors.New("Kubernetes API error")
var KUBERNETES_CREATE_CONFIGMAP_ERROR = errors.New("Cannot create configmap")

// Velero errors
var VELERO_RETIERIVE_STATUS_ERROR = errors.New("Cannot get Velero status")
var VELERO_STATUS_MISSING = errors.New("Cannot get Velero status")
var VELERO_CANNOT_MARSHALL_STATUS = errors.New("Cannot convert Velero status into struct")

// Mattermost errors
var MATTERMOST_CANNOT_CONVERT_BODY_TO_JSON = errors.New("Cannot convert Velero status into JSON")
var MATTERMOST_CANNOT_CREATE_REQUEST = errors.New("Cannot create mattermost request ")
var MATTERMOST_ERROR_SENDING_REQUEST = errors.New("Cannot send request to mattermost server")
var MATTERMOST_ERROR = errors.New("Mattermost server returned an error")
var MATTERMOST_TOO_MANY_REQUEST = errors.New("Mattermost returned Too many requests")
