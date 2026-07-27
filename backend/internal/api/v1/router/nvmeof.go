package router

import "cephtower/backend/internal/api/v1/handler"

func nvmeofRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/nvmeof/gateway", h.GetNVMeOFGateway},
		{"GET", "/nvmeof/subsystems", h.ListNVMeOFSubsystems},
		{"POST", "/nvmeof/subsystem", h.CreateNVMeOFSubsystem},
		{"GET", "/nvmeof/subsystem", h.GetNVMeOFSubsystem},
		{"PATCH", "/nvmeof/subsystem", h.UpdateNVMeOFSubsystem},
		{"DELETE", "/nvmeof/subsystem", h.DeleteNVMeOFSubsystem},
		{"GET", "/nvmeof/subsystem/namespaces", h.ListNVMeOFNamespaces},
		{"POST", "/nvmeof/subsystem/namespace", h.CreateNVMeOFNamespace},
		{"GET", "/nvmeof/subsystem/namespace", h.GetNVMeOFNamespace},
		{"PATCH", "/nvmeof/subsystem/namespace", h.UpdateNVMeOFNamespace},
		{"DELETE", "/nvmeof/subsystem/namespace", h.DeleteNVMeOFNamespace},
		{"GET", "/nvmeof/subsystem/listeners", h.ListNVMeOFListeners},
		{"POST", "/nvmeof/subsystem/listener", h.CreateNVMeOFListener},
		{"DELETE", "/nvmeof/subsystem/listener", h.DeleteNVMeOFListener},
		{"GET", "/nvmeof/subsystem/hosts", h.ListNVMeOFHosts},
		{"POST", "/nvmeof/subsystem/host", h.CreateNVMeOFHost},
		{"DELETE", "/nvmeof/subsystem/host", h.DeleteNVMeOFHost},
		{"GET", "/nvmeof/subsystem/connections", h.ListNVMeOFConnections},
	}
}
