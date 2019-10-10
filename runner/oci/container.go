type Container interface {
	volumes        []ContainerVolume
	id             string
	name           string
	logPath        string
	image          string
	sandbox        string
	netns          string
	runtimeHandler string
	// this is the /var/run/storage/... directory, erased on reboot
	bundlePath string
	// this is the /var/lib/storage/... directory
	dir                string
	stopSignal         string
	imageName          string
	imageRef           string
	mountPoint         string
	seccompProfilePath string
	conmonCgroupfsPath string
	labels             fields.Set
	annotations        fields.Set
	crioAnnotations    fields.Set
	state              *ContainerState
	metadata           *pb.ContainerMetadata
	opLock             sync.RWMutex
	spec               *specs.Spec
	idMappings         *idtools.IDMappings
	terminal           bool
	stdin              bool
	stdinOnce          bool
	privileged         bool
	created            bool
}
