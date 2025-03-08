package Config

type VeleroReportorConfig struct {
	MattermostUrl      string
	MattermostToken    string
	MattermostTimeout  string
	MattermostChannel  string
	MattermostUsername string

	KubeConfigType string
	KubeConfigPath string

	HttpHost string
	HttpPort string

	LogLevel string
}
