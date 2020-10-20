package allocator

type GameServer struct {
	UID             string            `json:"uid,omitempty"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	ResourceVersion string            `json:"resource_version,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Status          *GameServerStatus `json:"status,omitempty"`
}

type GameServerStatus struct {
	State   string        `json:"state,omitempty"`
	Address string        `json:"address,omitempty"`
	Players *PlayerStatus `json:"players,omitempty"`
}

type PlayerStatus struct {
	Count    int64    `json:"count"`
	Capacity int64    `json:"capacity"`
	IDs      []string `json:"ids"`
}
