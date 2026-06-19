package exporter

type InfoData struct {
	HardwareVersion string `json:"hardwareVersion"`
	SoftwareVersion string `json:"softwareVersion"`
}

// GET http://${address}/rest/v1/system/info_
type Info struct {
	InfoData InfoData `json:"info"`
}
