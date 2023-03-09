package dashboardexecute

type LeafRunOption = func(target *LeafRun)

func setName(name string) LeafRunOption {
	return func(target *LeafRun) {
		target.Name = name
	}
}
