package requestcontext

type RequestContext struct {
	Cid      string   `json:"x-cid"`
	Tenant   string   `json:"x-tenant"`
	UserID   int64    `json:"x-user-id"`
	UserName string   `json:"x-user-name"`
	Roles    []string `json:"roles"`
}

const CID = "x-cid"
const Tenant = "x-tenant"

const ContextKey = "ContextKey"
