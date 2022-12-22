package dashboardexecute

type RuntimeDependencySubscriber interface {
	RuntimeDependencyPublisher
	GetBaseDependencySubscriber() RuntimeDependencySubscriber
}
