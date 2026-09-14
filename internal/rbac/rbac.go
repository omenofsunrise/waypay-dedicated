package rbac

type User struct {
	ID          string
	Roles       []string
	Permissions []string
}

type Role struct {
	Name        string
	Permissions []string
}

type Engine struct {
	roles map[string]Role
}

func NewEngine() *Engine {
	return &Engine{roles: map[string]Role{}}
}

func (e *Engine) AddRole(r Role) {
	e.roles[r.Name] = r
}

func (e *Engine) HasRole(name string) bool {
	if _, ok := e.roles[name]; ok {
		return true
	}
	return false
}

func (e *Engine) Can(u User, perm string) bool {
	if contains(u.Permissions, perm) {
		return true
	}

	for _, strR := range u.Roles {
		if r, ok := e.roles[strR]; ok {
			if contains(r.Permissions, perm) {
				return true
			}
		}
	}
	return false
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
