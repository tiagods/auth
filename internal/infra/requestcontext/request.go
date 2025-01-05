package requestcontext

type RequestContext struct {
	Cid    string   `json:"x-cid"`
	Tenant string   `json:"x-tenant"`
	Roles  []string `json:"roles"`
}

const CID = "x-cid"
const Tenant = "x-tenant"

const ContextKey = "ContextKey"
