package spawner

type Spawn struct {
	env     string
	region  string
	spawner Spawner
}

func New(env, region string) *Spawn {
	s := &Spawn{env: env, region: region}
	if env == "local" {
		s.spawner = NewLocalSpawner(env, region)
	} else {
		s.spawner = NewAWSSpawner(env, region)
	}
	return s
}
