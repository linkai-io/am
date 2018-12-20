package am

import "errors"

var (
	ErrEmptyDBConfig     = errors.New("empty database connection string")
	ErrInvalidDBString   = errors.New("invalid db connection string")
	ErrOrgIDMismatch     = errors.New("org id does not user context")
	ErrUserNotAuthorized = errors.New("user is not authorized to perform this action")
	ErrLimitTooLarge     = errors.New("requested number of records too large")
	ErrNoResults         = errors.New("no results")

	// Scan Group Specific
	ErrScanGroupNotExists     = errors.New("scan group name does not exist")
	ErrScanGroupExists        = errors.New("scan group name already exists")
	ErrScanGroupVersionLinked = errors.New("scan group version is linked to this scan group")
	ErrAddressCopyCount       = errors.New("copy count of addresses did not match expected amount")
	ErrEmptyAddress           = errors.New("address data was nil")

	// Organization Specific
	ErrOrganizationExists    = errors.New("organization already exists")
	ErrOrganizationNotExists = errors.New("organization does not exist")

	// User Specific
	ErrUserExists   = errors.New("user already exists")
	ErrUserCIDEmpty = errors.New("user cid is empty")
)
