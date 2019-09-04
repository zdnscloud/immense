package client

type Df struct {
	Nodes   []Node   `json:"nodes"`
	Stray   []string `json:"stray"`
	Summary Sum      `json:"summary"`
}

type Node struct {
	ID          int64              `json:"id"`
	DeviceClass string             `json:"device_class"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	TypeID      int64              `json:"type_id"`
	CrushWeight float64            `json:"crush_weight"`
	Depth       int64              `json:"depth"`
	PoolWeight  map[string]float64 `json:"pool_weights"`
	Reweight    float64            `json:"reweight"`
	Total       int64              `json:"kb"`
	Used        int64              `json:"kb_used"`
	Avail       int64              `json:"kb_avail"`
	Util        float64            `json:"utilization"`
	Var         float64            `json:"var"`
	Pgs         int64              `json:"pgs"`
	Status      string             `json:"status"`
}

type Sum struct {
	Total  int64   `json:"total_kb"`
	Used   int64   `json:"total_kb_used"`
	Avail  int64   `json:"total_kb_avail"`
	Util   float64 `json:"average_utilization"`
	MinVar float64 `json:"min_var"`
	MaxVar float64 `json:"max_var"`
	dev    float64 `json:"dev"`
}

type MonDump struct {
	Epoch    int    `json:"epoch"`
	FSID     string `json:"fsid"`
	Modified string `json:"modified"`
	Created  string `json:"created"`
	Features `json:"features"`
	Mons     []Mon `json:"mons"`
	Quorum   []int `json:"quorum"`
}

type Features struct {
	Persistent []string `json:"persistent"`
	Optional   []string `json:"optional"`
}
type Mon struct {
	Rank       int    `json:"rank"`
	Name       string `json:"name"`
	Addr       string `json:"addr"`
	PublicAddr string `json:"public_addr"`
}
