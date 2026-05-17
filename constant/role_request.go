package constant

// RoleRequestStatus values represent the lifecycle of a role request.
const (
	RoleRequestStatusRequested = "REQUESTED"
	RoleRequestStatusApproved  = "APPROVED"
	RoleRequestStatusRejected  = "REJECTED"
)

// RoleRequestRole values — only user-requestable roles.
const (
	RoleRequestRoleBidder = "BIDDER"
	RoleRequestRoleSeller = "SELLER"
)
