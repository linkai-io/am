package mock

type UserContext struct {
	GetTraceIDFn        func() string
	GetTradeIDFnInvoked bool

	GetOrgIDFn        func() int32
	GetOrgIDFnInvoked bool

	GetUserIDFn        func() int32
	GetUserIDFnInvoked bool

	GetRolesFn        func() []string
	GetRolesFnInvoked bool

	GetIPAddressFn        func() string
	GetIPAddressFnInvoked bool
}

func (u *UserContext) GetTraceID() string {
	u.GetTradeIDFnInvoked = true
	return u.GetTraceIDFn()
}

func (u *UserContext) GetOrgID() int32 {
	u.GetOrgIDFnInvoked = true
	return u.GetOrgIDFn()
}

func (u *UserContext) GetUserID() int32 {
	u.GetUserIDFnInvoked = true
	return u.GetUserIDFn()
}

func (u *UserContext) GetRoles() []string {
	u.GetRolesFnInvoked = true
	return u.GetRolesFn()
}

func (u *UserContext) GetIPAddress() string {
	u.GetIPAddressFnInvoked = true
	return u.GetIPAddressFn()
}
