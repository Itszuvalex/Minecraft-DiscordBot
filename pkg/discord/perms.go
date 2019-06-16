package discord // "github.com/itszuvalex/mcdiscord/pkg/discord"
import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/itszuvalex/mcdiscord/pkg/api"
)

type PermManager struct {
	Roots map[string]*PermNode `json:"roots"`
}

func NewPermHandler() api.IPermHandler {
	return &PermManager{Roots: make(map[string]*PermNode)}
}

type PermNode struct {
	NodeName    string                `json:"name"`
	Fullname    string                `json:"fullname"`
	Children    map[string]*PermNode  `json:"children"`
	GuildPerms  map[string]*GuildPerm `json:"guildperms"`
	Permdefault api.PermDefault       `json:"permdefault"`
}

type GuildPerm struct {
	ID        string           `json:"id"`
	UserPerms map[string]*Perm `json:"userperms"`
	RolePerms map[string]*Perm `json:"roleperms"`
}

func (g *GuildPerm) GetUserPerm(user string) (api.IPerm, error) {
	perm, ok := g.UserPerms[user]
	if !ok {
		return nil, errors.New("User not found")
	}
	return perm, nil
}

func (g *GuildPerm) GetRolePerm(user string) (api.IPerm, error) {
	perm, ok := g.RolePerms[user]
	if !ok {
		return nil, errors.New("Role not found")
	}
	return perm, nil
}

type Perm struct {
	ID      string `json:"id"`
	Allowed bool   `json:"y"`
}

func (p *Perm) PermID() string {
	return p.ID
}

func (p *Perm) PermAllowed() bool {
	return p.Allowed
}

func (perm *PermManager) GetRoot(name string) (api.IPermNode, error) {
	root, ok := perm.Roots[name]
	if ok {
		return root, nil
	}
	return nil, errors.New("Root permission '" + name + "' not found.")
}

func (perm *PermManager) GetOrAddRoot(name string) api.IPermNode {
	root, err := perm.GetRoot(name)
	if err == nil {
		return root
	}
	fmt.Println("Added perm root:", name)
	newroot := &PermNode{NodeName: name, Fullname: name, Children: make(map[string]*PermNode), GuildPerms: make(map[string]*GuildPerm), Permdefault: api.Block}
	perm.Roots[name] = newroot
	return newroot
}

func (perm *PermManager) GetPermNode(name string) (api.IPermNode, error) {
	return perm.GetPermNodeByNodes(strings.Split(name, "."))
}

func (perm *PermManager) GetPermNodeByNodes(nodes []string) (api.IPermNode, error) {
	root, err := perm.GetRoot(nodes[0])
	if err != nil {
		return nil, err
	}
	if len(nodes) > 1 {
		return root.GetPermNodeByNodes(nodes[1:])
	} else {
		return root, nil
	}
}

func (node *PermNode) GetGuildPerm(guildid string) (api.IGuildPerm, error) {
	guild, ok := node.GuildPerms[guildid]
	if !ok {
		return nil, errors.New("Guild not found")
	}
	return guild, nil
}

func (node *PermNode) GetOrAddPermNode(name string) api.IPermNode {
	return node.GetOrAddPermNodeByNodes(strings.Split(name, "."))
}

func (node *PermNode) GetOrAddPermNodeByNodes(subnodes []string) api.IPermNode {
	if len(subnodes) == 0 {
		return nil
	}
	child, ok := node.Children[subnodes[0]]
	if ok {
		if len(subnodes) > 1 {
			return child.GetOrAddPermNodeByNodes(subnodes[1:])
		}
		return child
	}
	fmt.Println("Added perm node: ", node.FullName()+"."+subnodes[0])
	newchild := &PermNode{NodeName: subnodes[0], Fullname: node.FullName() + "." + subnodes[0], Children: make(map[string]*PermNode), GuildPerms: make(map[string]*GuildPerm), Permdefault: api.Inherit}
	node.Children[subnodes[0]] = newchild
	if len(subnodes) > 1 {
		return newchild.GetOrAddPermNodeByNodes(subnodes[1:])
	}
	return newchild
}

func (node *PermNode) GetPermNode(name string) (api.IPermNode, error) {
	return node.GetPermNodeByNodes(strings.Split(name, "."))
}

func (node *PermNode) GetPermNodeByNodes(subnodes []string) (api.IPermNode, error) {
	if len(subnodes) == 0 {
		return nil, errors.New("No subnodes left to check.")
	}
	child, ok := node.Children[subnodes[0]]
	if ok {
		if len(subnodes) > 1 {
			return child.GetPermNodeByNodes(subnodes[1:])
		}
		return child, nil
	}
	return nil, errors.New("Cannot find nodes : " + strings.Join(subnodes, "."))
}

func (node *PermNode) SetPermDefault(permd api.PermDefault) {
	node.Permdefault = permd
}

func (node *PermNode) PermDefault() api.PermDefault {
	return node.Permdefault
}

func (node *PermNode) Name() string {
	return node.NodeName
}

func (node *PermNode) FullName() string {
	return node.Fullname
}

func (node *PermNode) AddOrSetRolePerm(guildid, role string, allow bool) {
	pguild := node.getOrAddGuildPerm(guildid)
	prole, ok := pguild.RolePerms[role]
	if !ok {
		pguild.RolePerms[role] = &Perm{ID: role, Allowed: allow}
	} else {
		prole.Allowed = allow
	}
}

func (node *PermNode) AddOrSetUserPerm(guildid, user string, allow bool) {
	pguild := node.getOrAddGuildPerm(guildid)
	puser, ok := pguild.UserPerms[user]
	if !ok {
		pguild.UserPerms[user] = &Perm{ID: user, Allowed: allow}
	} else {
		puser.Allowed = allow
	}
}

func (node *PermNode) getOrAddGuildPerm(guildid string) *GuildPerm {
	pguild, ok := node.GuildPerms[guildid]
	if !ok {
		pguild = &GuildPerm{ID: guildid, UserPerms: make(map[string]*Perm), RolePerms: make(map[string]*Perm)}
		node.GuildPerms[guildid] = pguild
	}
	return pguild
}

func (node *PermNode) IsChild(p api.IPermNode) bool {
	return IsPathChild(node.FullName(), p.FullName())
}

func IsPathChild(child, parent string) bool {
	return strings.HasPrefix(child, parent)
}

func (perm *PermManager) WriteJson() (json.RawMessage, error) {
	msgdata, err := json.Marshal(perm)
	if err != nil {
		fmt.Println("Error marshalling Permissions, ", err)
		return nil, err
	}
	return msgdata, nil
}

func (perm *PermManager) ReadJson(data json.RawMessage) error {
	return json.Unmarshal(data, perm)
}

type permhandler func(api.IPermNode) (bool, error)

func (perm *PermManager) IsUserAllowed(nodepath, guildid, user string) (api.PermCheck, error) {
	return perm.isAllowed(nodepath, func(node api.IPermNode) (bool, error) {
		guild, err := node.GetGuildPerm(guildid)
		if err != nil {
			return false, err
		}
		user, err := guild.GetUserPerm(user)
		if err != nil {
			return false, err
		}
		fmt.Println("Found explicit user setting:", user.PermAllowed(), " at path:", nodepath)
		return user.PermAllowed(), nil
	})
}

func (perm *PermManager) IsRoleAllowed(nodepath, guildid, role string) (api.PermCheck, error) {
	return perm.isAllowed(nodepath, func(node api.IPermNode) (bool, error) {
		guild, err := node.GetGuildPerm(guildid)
		if err != nil {
			return false, err
		}
		role, err := guild.GetRolePerm(role)
		if err != nil {
			return false, err
		}
		fmt.Println("Found explicit role setting:", role.PermAllowed(), " at path:", nodepath)
		return role.PermAllowed(), nil
	})
}

func (perm *PermManager) isAllowed(nodepath string, checkFunc permhandler) (api.PermCheck, error) {
	nodes := strings.Split(nodepath, ".")
	implicit := false
	implicitValue := false
	implicitNode := ""
	for i := 0; i < len(nodes); i++ {
		subnodes := nodes[:len(nodes)-i]

		fmt.Println("Checking node:", strings.Join(subnodes, "."))
		// Walk up from the deepest node
		node, err := perm.GetPermNodeByNodes(subnodes)
		if err != nil {
			return api.PermCheck{false, strings.Join(subnodes, "."), false}, err
		}

		// Find explicit yes/no
		// We only do NOT receive an error if we find an explicit
		check, err := checkFunc(node)
		if err == nil {
			return api.PermCheck{check, strings.Join(subnodes, "."), true}, nil
		}

		// Explicits ANYWHERE in the path ovewrite ANY implicits found
		// That's why we check the full path even if we find an implicit.

		// Otherwise, find the deepest implicit in the path
		if !implicit {
			switch node.PermDefault() {
			case api.Allow:
				implicit = true
				implicitValue = true
				implicitNode = strings.Join(subnodes, ".")
				fmt.Println("Found implicit Allow at:", implicitNode)
			case api.Block:
				implicit = true
				implicitValue = false
				implicitNode = strings.Join(subnodes, ".")
				fmt.Println("Found implicit Block at:", implicitNode)
			case api.Inherit:
			}
		}

	}

	return api.PermCheck{implicitValue, implicitNode, !implicit}, nil
}

// Cases:
//   User explicitly allowed anywhere in the path : allow
//   User explicitly disallowed anywhere in the path : disallow
//   Use furthest path explicit set for any role result
func (perm *PermManager) UserWithRolesAllowed(nodepath, guildid, user string, roles []string) (api.PermCheck, error) {
	userCheckResult, err := perm.IsUserAllowed(nodepath, guildid, user)
	if err != nil {
		return userCheckResult, err
	}

	// If user explicitly allowed/disallowed
	if userCheckResult.Explicit {
		if userCheckResult.Allowed {
			return userCheckResult, nil
		} else {
			return userCheckResult, errors.New("User is explicitly disallowed from running this command at path: " + userCheckResult.Path)
		}
	}

	type permcheckbyrole struct {
		perm *api.PermCheck
		role string
	}

	var roleCheckResultsAllowImplicit *permcheckbyrole
	var roleCheckResultsExplicit []permcheckbyrole
	for _, role := range roles {
		res, err := perm.IsRoleAllowed(nodepath, guildid, role)
		if err != nil {
			return userCheckResult, err
		}
		if res.Explicit {
			roleCheckResultsExplicit = append(roleCheckResultsExplicit, permcheckbyrole{&res, role})
			break
		} else if roleCheckResultsAllowImplicit == nil && res.Allowed {
			roleCheckResultsAllowImplicit = &permcheckbyrole{&res, role}
		}
	}

	if len(roleCheckResultsExplicit) > 0 {
		deepest := roleCheckResultsExplicit[0]
		for _, check := range roleCheckResultsExplicit[1:] {
			if IsPathChild(check.perm.Path, deepest.perm.Path) {
				deepest = check
			}
		}

		if deepest.perm.Allowed {
			return *deepest.perm, nil
		} else {
			return *deepest.perm, errors.New("Role " + deepest.role + " disallowed access at path:" + deepest.perm.Path)
		}
	}

	// Implicitly allow user
	if userCheckResult.Allowed {
		fmt.Println("Implicitly allowed user at node: ", userCheckResult.Path)
		return userCheckResult, nil
	}

	if roleCheckResultsAllowImplicit != nil && roleCheckResultsAllowImplicit.perm.Allowed {
		fmt.Println("Implicitly allowed by role:", roleCheckResultsAllowImplicit.role, " at node: ", roleCheckResultsAllowImplicit.perm.Path)
		return *roleCheckResultsAllowImplicit.perm, nil
	}

	return userCheckResult, errors.New("Could not find any allowed paths.")
}
