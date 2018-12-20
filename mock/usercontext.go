package mock

type UserContext struct {
	GetTraceIDFn        func() string
	GetTradeIDFnInvoked bool

	GetOrgIDFn        func() int
	GetOrgIDFnInvoked bool

	GetOrgCIDFn        func() string
	GetOrgCIDFnInvoked bool

	GetUserIDFn        func() int
	GetUserIDFnInvoked bool

	GetUserCIDFn        func() string
	GetUserCIDFnInvoked bool

	GetRolesFn        func() []string
	GetRolesFnInvoked bool

	GetIPAddressFn        func() string
	GetIPAddressFnInvoked bool
}

func (u *UserContext) GetTraceID() string {
	u.GetTradeIDFnInvoked = true
	if u.GetTraceIDFn == nil {
		return ""
	}
	return u.GetTraceIDFn()
}

func (u *UserContext) GetOrgID() int {
	u.GetOrgIDFnInvoked = true
	return u.GetOrgIDFn()
}

func (u *UserContext) GetOrgCID() string {
	u.GetOrgCIDFnInvoked = true
	return u.GetOrgCIDFn()
}

func (u *UserContext) GetUserID() int {
	u.GetUserIDFnInvoked = true
	return u.GetUserIDFn()
}

func (u *UserContext) GetUserCID() string {
	u.GetUserCIDFnInvoked = true
	return u.GetUserCIDFn()
}

func (u *UserContext) GetRoles() []string {
	u.GetRolesFnInvoked = true
	return u.GetRolesFn()
}

func (u *UserContext) GetIPAddress() string {
	u.GetIPAddressFnInvoked = true
	return u.GetIPAddressFn()
}
