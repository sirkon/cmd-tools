package portallog

// PortalLogger логирование операций над хранилищем порталов.
type PortalLogger interface {
	AddPortal(name string, path string) error
	DeletePortal(name string) error
}
