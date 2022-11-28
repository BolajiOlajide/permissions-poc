package main

import "fmt"

// User represents a user of the application.
type User struct {
	ID   int
	Name string
}

func (u *User) checkNamespaceAccess(namespace, action string) (bool, error) {
	var id int
	err := db.QueryRow(`SELECT
	p.id
FROM
	permissions p
INNER JOIN user_roles ur ON ur.user_id = $3
INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
WHERE rp.permission_id = p.id AND p.namespace = $1 AND p.action = $2
LIMIT 1;`, namespace, action, u.ID).Scan(&id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return id > 0, nil
}

var mockUserNames = []string{
	"kai",
	"elliot",
	"jalen",
	"adrian",
}

var userRoleMapping = map[string][]string{
	"kai":    {"DEFAULT", "SITE_ADMINISTRATOR"},
	"elliot": {"DEFAULT", "OPERATOR"},
	"jalen":  {"DEFAULT"},
	"adrian": {},
}

func seedUsers(c *Controller, roles ...*Role) ([]*User, error) {
	users := make([]*User, len(mockUserNames), len(mockUserNames))

	for idx, name := range mockUserNames {
		user, err := c.createUser(name)
		if err != nil {
			return users, err
		}

		// add users Kai and Jalen to the Acme org
		if name == "kai" || name == "jalen" {
			fmt.Printf("Adding user %s to Acme org.", name)
			org, err := c.getOrgByName("ACME")
			if err != nil {
				return users, err
			}
			err = c.addUserToOrg(user, org)
			if err != nil {
				return users, err
			}
		}

		fmt.Println("")
		roleNames := userRoleMapping[name]
		roles := getUserRoles(roleNames, roles)

		err = c.addRoleForUser(user, roles...)
		if err != nil {
			return users, err
		}

		users[idx] = user
	}

	return users, nil
}

func getUserRoles(roleNames []string, roles []*Role) (rs []*Role) {
	for _, r := range roles {
		for _, rn := range roleNames {
			if r.Name == rn {
				rs = append(rs, r)
			}
		}
	}

	return
}
