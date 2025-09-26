package models

type Target struct {
	IP       string
	Username string
	Password string
}

type ProbeResult struct {
	Protocol string
	Port     int
	Success  bool
	Error    string
	Banner   string
}
