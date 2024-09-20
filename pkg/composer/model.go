package composer

type Composer struct {
	ID          string `json:"id"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	BirthDate   string `json:"birthDate"`
	DeathDate   string `json:"deathDate"`
	Era         string `json:"era"`
	Nationality string `json:"nationality"`
}
