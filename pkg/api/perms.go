package api // "github.com/itszuvalex/mcdiscord/pkg/api"

import (
	"encoding/json"
)

type PermDefault int

const (
	Inherit PermDefault = 0
	Allow   PermDefault = 1
	Block   PermDefault = 2
)

type PermCheck struct {
	Allowed  bool
	Path     string
	Explicit bool
}

type IPermHandler interface {
	GetRoot(name string) (IPermNode, error)
	GetOrAddRoot(name string) IPermNode
	GetPermNode(name string) (IPermNode, error)
	GetPermNodeByNodes(nodes []string) (IPermNode, error)
	IsUserAllowed(nodepath, guildid, user string) (PermCheck, error)
	IsRoleAllowed(nodepath, guildid, role string) (PermCheck, error)
	UserWithRolesAllowed(nodepath, guildid, user string, roles []string) (PermCheck, error)
	WriteJson() (json.RawMessage, error)
	ReadJson(json json.RawMessage) error
	RootNodes() []IPermNode
	RecursiveGetAllNodes() []IPermNode
}

type IPermNode interface {
	Name() string
	FullName() string
	ChildNodes() []IPermNode
	GetOrAddPermNode(name string) IPermNode
	GetOrAddPermNodeByNodes(subnodes []string) IPermNode
	GetPermNode(name string) (IPermNode, error)
	GetPermNodeByNodes(subnodes []string) (IPermNode, error)
	SetPermDefault(permd PermDefault)
	PermDefault() PermDefault
	AddOrSetRolePerm(guildid, role string, allow bool)
	AddOrSetUserPerm(guildid, user string, allow bool)
	RemoveRolePerm(guildid, role string)
	RemoveUserPerm(guildid, role string)
	IsChild(p IPermNode) bool
	GetGuildPerm(guildid string) (IGuildPerm, error)
	RecursiveGetAllChildren() []IPermNode
}

type IGuildPerm interface {
	GetUserPerm(user string) (IPerm, error)
	GetRolePerm(role string) (IPerm, error)
	PermsUser() []IPerm
	PermsRole() []IPerm
}

type IPerm interface {
	PermID() string
	PermAllowed() bool
}
