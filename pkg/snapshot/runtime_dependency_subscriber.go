package snapshot

type RuntimeDependencySubscriber interface {
	RuntimeDependencyPublisher
	GetBaseDependencySubscriber() RuntimeDependencySubscriber
}
