package collector

type DataFetchConfig struct {
	Module              string `json:"module"`
	Enabled             bool   `json:"enabled"`
	IntervalSeconds     int    `json:"interval_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	JitterSeconds       int    `json:"jitter_seconds"`
	FetchSource         string `json:"fetch_source"`
	Priority            int    `json:"priority"`
	MaxRetries          int    `json:"max_retries"`
	RetryBackoffSeconds int    `json:"retry_backoff_seconds"`
}

type dataFetchModuleDefault struct {
	module          string
	source          string
	intervalSeconds int
	priority        int
}

var defaultDataFetchModules = []dataFetchModuleDefault{
	{fetchModuleSummary, fetchSourceDashboard, 60, 10}, {fetchModuleHealth, fetchSourceDashboard, 60, 20},
	{fetchModuleHosts, fetchSourceCommand, 300, 30}, {fetchModuleDaemons, fetchSourceCommand, 300, 40},
	{fetchModuleServices, fetchSourceCommand, 300, 50}, {fetchModuleMonitors, fetchSourceCommand, 300, 60},
	{fetchModuleManagers, fetchSourceCommand, 300, 70}, {fetchModuleMDS, fetchSourceCommand, 300, 80},
	{fetchModuleOSDs, fetchSourceCommand, 300, 90}, {fetchModuleOSDFlags, fetchSourceCommand, 300, 100},
	{fetchModuleMgrModules, fetchSourceCommand, 600, 110}, {fetchModuleClusterConfiguration, fetchSourceCommand, 900, 120},
	{fetchModulePools, fetchSourceDashboard, 300, 130}, {fetchModuleRBDImages, fetchSourceDashboard, 300, 140},
	{fetchModuleCephFS, fetchSourceDashboard, 300, 150}, {fetchModuleRGWDaemons, fetchSourceDashboard, 300, 160},
	{fetchModuleRGWUsers, fetchSourceDashboard, 600, 170}, {fetchModuleRGWBuckets, fetchSourceDashboard, 600, 180},
	{fetchModuleSettings, fetchSourceDashboard, 600, 190}, {fetchModuleFeatureToggles, fetchSourceDashboard, 300, 200},
	{fetchModuleIntegrationStatus, fetchSourceDashboard, 300, 210},
}
