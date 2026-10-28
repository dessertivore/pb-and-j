package main

type Side string

const (
	SideBlue Side = "blue"
	SideRed  Side = "red"
)

type ActionType string

const (
	ActionBan  ActionType = "ban"
	ActionPick ActionType = "pick"
)

type Action struct {
	Side     Side       `json:"side"`
	Type     ActionType `json:"type"`
	Champion string     `json:"champion,omitempty" doc:"Empty for a skipped ban"`
	Role     string     `json:"role" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role. User will have to guess this for opponent."`
}

type PickBan struct {
	ID      string   `json:"id"`
	Actions []Action `json:"actions" doc:"In action order"`
}

// By returns the champions a side picked or banned, in action order.
func (d PickBan) By(side Side, t ActionType) []string {
	var out []string
	for _, a := range d.Actions {
		if a.Side == side && a.Type == t {
			out = append(out, a.Champion)
		}
	}
	return out
}
