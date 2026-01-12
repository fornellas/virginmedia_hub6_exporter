package exporter


// General Configuration
type CableModem struct {
	// Config file
    BootFilename string `json:"bootFilename"`
    // DOCSIS Mode
    DocsisVersion string `json:"docsisVersion"`
    MacAddress string `json:"macAddress"`
    SerialNumber string `json:"serialNumber"`
    UpTime uint64 `json:"upTime"`
    // Network access
    AccessAllowed bool `json:"accessAllowed"`
    Status string `json:"status"`
    // Maximum Number of CPEs
    MaxCpEs uint64 `json:"maxCPEs"`
    // Baseline Privacy
    BaselinePrivacyEnabled bool `json:"baselinePrivacyEnabled"`
}

// /rest/v1/cablemodem/state_
type State struct {
	CableModem CableModem `json:cablemodem"`
}
