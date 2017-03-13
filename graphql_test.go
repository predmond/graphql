package graphql

import (
	"fmt"
	"strings"
	"testing"
)

func diff(a, b string) error {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return nil
	}
	return fmt.Errorf("got:\n%s\nexpected:\n%s", a, b)
}

func TestArguments(t *testing.T) {
	query := struct {
		Human struct {
			Name   string
			Height float32 `graphql:"unit: FOOT"`
		} `graphql:"id: \"1000\""`
	}{}

	expectedQuery := `{
  human(id: "1000") {
    name
    height(unit: FOOT)
  }
}`

	if q, err := NewQuery("", query).MarshalString(); err != nil {
		t.Error(err)
	} else if err := diff(q, expectedQuery); err != nil {
		t.Error("wrong query:\n", err)
	}

	/*
		result := `{
			"data": {
				"human": {
					"name": "Luke Skywalker",
					"height": 5.6430448
				}
			}
		}`

		if err := Unmarshal([]byte(result), &query); err != nil {
			t.Error(err)
		}
		if query.Human.Name != "Luke Skywalker" {
			t.Error("wrong name", query.Human.Name)
		}
		if query.Human.Height != 5.6430448 {
			t.Error("wrong height", query.Human.Height)
		}
	*/
}

func TestMarshalGraphQL(t *testing.T) {
	query := struct {
		Human struct {
			Name   string
			Height float32 `graphql:"unit: FOOT"`
			Test   struct {
				HideStruct
				Name string
			}
		} `graphql:"id: \"1000\""`
	}{}

	expectedQuery := `{
  human(id: "1000") {
    name
    height(unit: FOOT)
    test
  }
}`

	if q, err := NewQuery("", query).MarshalString(); err != nil {
		t.Error(err)
	} else if err := diff(q, expectedQuery); err != nil {
		t.Error("wrong query:\n", err)
	}
}

/*
func TestAlias(t *testing.T) {
	type Hero struct {
		MakeQueryable
		Name string
	}
	query := struct {
		EmpireHero Hero `graphql:"episode: EMPIRE"`
		JediHero   Hero `graphql:"episode: JEDI"`
	}{}

	_ = query

	expected := `{
	empireHero: hero(episode: EMPIRE) {
		name
	}
	jediHero: hero(episode: JEDI) {
		name
	}
}`
	_ = expected

	test := `{
		"data": {
			"empireHero": {
				"name": "Luke Skywalker"
			},
			"jediHero": {
				"name": "R2-D2"
			}
		}
	}`

	_ = test
}
*/

func TestVariables(t *testing.T) {
	query := struct {
		Hero struct {
			Name    string
			Friends []struct {
				Name string
			}
		} `graphql:"episode: $episode"`
	}{}

	expectedQuery := `query HeroNameAndFriends($episode: Episode) {
  hero(episode: $episode) {
    name
    friends {
      name
    }
  }
}`

	if q, err := NewQuery("HeroNameAndFriends", query).DefineVariable("episode", "Episode").MarshalString(); err != nil {
		t.Error(err)
	} else if err := diff(q, expectedQuery); err != nil {
		t.Error("wrong query:\n", err)
	}

	result := `{
  "data": {
    "hero": {
      "name": "R2-D2",
      "friends": [
        {
          "name": "Luke Skywalker"
        },
        {
          "name": "Han Solo"
        },
        {
          "name": "Leia Organa"
        }
      ]
    }
  }
}`
	_ = result

}

func TestConnection(t *testing.T) {
	expectedQuery := `{
  hero {
    name
    friendsConnection(first:2 after:"Y3Vyc29yMQ==") {
      edges {
        node {
          name
        }
        cursor
      }
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
    }
  }
}`

	query := struct {
		Hero struct {
			Name              string
			FriendsConnection struct {
				Edges []struct {
					Node struct {
						Name string
					}
					Cursor string
				}
				Connection
			} `graphql:"first:2 after:\"Y3Vyc29yMQ==\""`
		}
	}{}

	if q, err := NewQuery("", query).MarshalString(); err != nil {
		t.Error(err)
	} else if err := diff(q, expectedQuery); err != nil {
		t.Error("wrong query:\n", err)
	}

	response := `{
  "data": {
    "hero": {
      "name": "R2-D2",
      "friendsConnection": {
        "totalCount": 3,
        "edges": [
          {
            "node": {
              "name": "Han Solo"
            },
            "cursor": "Y3Vyc29yMg=="
          },
          {
            "node": {
              "name": "Leia Organa"
            },
            "cursor": "Y3Vyc29yMw=="
          }
        ],
        "pageInfo": {
          "hasNextPage": false,
          "startCursor": "Y3Vyc29yMg==",
          "endCursor": "Y3Vyc29yMw=="
        }
      }
    }
  }
}`

	if err := Unmarshal([]byte(response), &query); err != nil {
		t.Error(err)
	}
	if query.Hero.Name != "R2-D2" {
		t.Error("wrong hero")
	}
	if query.Hero.FriendsConnection.TotalCount != 3 {
		t.Error("wrong total count")
	}
	if len(query.Hero.FriendsConnection.Edges) != 2 {
		t.Error("wrong edge count")
	}
	if query.Hero.FriendsConnection.Edges[0].Node.Name != "Han Solo" {
		t.Error("wrong friend 0")
	}
	if query.Hero.FriendsConnection.Edges[0].Cursor != "Y3Vyc29yMg==" {
		t.Error("wrong friend 0 cursor")
	}
	if query.Hero.FriendsConnection.Edges[1].Node.Name != "Leia Organa" {
		t.Error("wrong friend 1")
	}
	if query.Hero.FriendsConnection.Edges[1].Cursor != "Y3Vyc29yMw==" {
		t.Error("wrong friend 1 cursor")
	}
	if query.Hero.FriendsConnection.PageInfo.HasNextPage != false {
		t.Error("wrong HasNextPage")
	}
	if query.Hero.FriendsConnection.PageInfo.HasPreviousPage != false {
		t.Error("wrong HasPreviousPage")
	}
	if query.Hero.FriendsConnection.PageInfo.StartCursor != "Y3Vyc29yMg==" {
		t.Error("wrong StartCursor")
	}
	if query.Hero.FriendsConnection.PageInfo.EndCursor != "Y3Vyc29yMw==" {
		t.Error("wrong EndCursor")
	}
}
