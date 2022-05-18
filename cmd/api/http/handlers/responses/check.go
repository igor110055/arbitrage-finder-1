package responses

type Health struct {
	Status string `json:"status"`
}

type Info struct {
	Status    string `json:"status,omitempty"`
	Build     string `json:"build,omitempty"`
	Host      string `json:"host,omitempty"`
	Pod       string `json:"pod,omitempty"`
	PodIP     string `json:"podIP,omitempty"`
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}
