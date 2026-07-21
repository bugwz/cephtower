package v1

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, cephClient CephClient) {
	api := &API{ceph: cephClient}

	registerClusterRoutes(mux, api)
	registerHostRoutes(mux, api)
	registerOSDRoutes(mux, api)
	registerMonitorRoutes(mux, api)
	registerMgrRoutes(mux, api)
	registerDaemonRoutes(mux, api)
	registerServiceRoutes(mux, api)
	registerPoolRoutes(mux, api)
	registerBlockRoutes(mux, api)
	registerFilesystemRoutes(mux, api)
	registerObjectRoutes(mux, api)
	registerConfigurationRoutes(mux, api)
	registerLogRoutes(mux, api)
}
