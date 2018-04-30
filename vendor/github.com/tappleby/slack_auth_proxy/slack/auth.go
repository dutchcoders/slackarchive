package slack

type AuthService struct {
	api *SlackClient
}

type Auth struct {
	UserId 	 string  `json:"user_id"`
	Username string	 `json:"user"`
	Team 	 string  `json:"team"`
	TeamId 	 string  `json:"team_id"`
	TeamUrl  string  `json:"url"`
}

func (s *AuthService) Test() (*Auth, error) {

	req, _ := s.api.NewRequest(_GET, "auth.test", nil)

	auth := new(Auth)

	_, err := s.api.Do(req, auth)

	if err != nil {
		return nil, err
	}

	return auth, nil
}
