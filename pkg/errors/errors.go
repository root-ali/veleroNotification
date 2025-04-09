package errors

import "errors"

// Kubernetes errors
var KUBERNETES_HEALTH_ERROR = errors.New("cannot check kubernetes health api")
var KUBERNETES_API_NOT_READY = errors.New("kubernetes API not ready")
var KUBERNETES_API_ERROR = errors.New("kubernetes API error")
var KUBERNETES_CREATE_CONFIGMAP_ERROR = errors.New("cannot create configmap")

// Velero errors
var VELERO_RETIERIVE_STATUS_ERROR = errors.New("cannot get Velero status")
var VELERO_STATUS_MISSING = errors.New("cannot get Velero status")
var VELERO_CANNOT_MARSHALL_STATUS = errors.New("cannot convert Velero status into struct")
var VELERO_ERROR_RETIEVIE_CONFIGMAP = errors.New("cannot retierive configmap")
var VELERO_RESOURCEVERSION_IS_NULL = errors.New("resource version in configmap is null")
var VELERO_CANNOT_CONVERT_RESOURCE_VERSION_TO_INT = errors.New("cannot convert resource version to int")
var VELERO_UPDATE_CONFIGMAP_ERROR = errors.New("cannot update configmap")
var VELERO_BACKUP_NOT_COMPLETED = errors.New("velero backup not in complete status")

// Mattermost errors
var MATTERMOST_CANNOT_CONVERT_BODY_TO_JSON = errors.New("cannot convert Velero status into JSON")
var MATTERMOST_CANNOT_CREATE_REQUEST = errors.New("cannot create mattermost request ")
var MATTERMOST_ERROR_SENDING_REQUEST = errors.New("cannot send request to mattermost server")
var MATTERMOST_ERROR = errors.New("mattermost server returned an error")
var MATTERMOST_TOO_MANY_REQUEST = errors.New("mattermost returned Too many requests")
