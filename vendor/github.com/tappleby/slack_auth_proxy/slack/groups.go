package slack

type GroupService struct {
	api *SlackClient
}

type Group struct {
	Id string
	Name string
	IsArchived bool
}

type GroupList []Group

func (g *GroupService) List() (GroupList, error) {

	req, _ := g.api.NewRequest(_GET, "groups.list", nil)

	type list struct {
		Ok bool
		Groups []Group `json:"groups"`
	}

	l := new(list)

	_, err := g.api.Do(req, l)
	
	if err != nil {
		return nil, err
	}

	return l.Groups, nil
}


func (list GroupList) FindName(name string) *Group {
	var group *Group = nil

	for _, g := range list {
		if g.Name == name {
			group = &g
			break

		}
	}

	return group
}
